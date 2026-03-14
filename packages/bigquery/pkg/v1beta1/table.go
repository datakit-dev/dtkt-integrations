package v1beta1

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"runtime"
	"slices"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	storage "cloud.google.com/go/bigquery/storage/apiv1"
	"cloud.google.com/go/bigquery/storage/apiv1/storagepb"
	"cloud.google.com/go/bigquery/storage/managedwriter"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"github.com/googleapis/gax-go/v2"
	goavro "github.com/linkedin/goavro/v2"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	AVRO_FORMAT        = "avro"
	ARROW_FORMAT       = "arrow"
	MAX_READ_MSG_SIZE  = 1024 * 1024 * 129
	MAX_WRITE_MSG_SIZE = 1024 * 1024 * 10
)

var (
	minReadStreams = runtime.NumCPU() * 2
	rpcOpts        = gax.WithGRPCOptions(
		grpc.MaxCallRecvMsgSize(MAX_READ_MSG_SIZE),
	)
)

type (
	TableService struct {
		catalogv1beta1.UnimplementedTableServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
	Table struct {
		table *bigquery.Table
		meta  *bigquery.TableMetadata
		desc  protoreflect.MessageDescriptor
	}
	Tables []*Table
)

func NewTableService(mux v1beta1.InstanceMux[*Instance]) *TableService {
	return &TableService{
		mux: mux,
	}
}

func NewTable(table *bigquery.Table, meta *bigquery.TableMetadata) *Table {
	return &Table{
		table: table,
		meta:  meta,
	}
}

func NewTableWriter(ctx context.Context, client *lib.Client) (*managedwriter.Client, error) {
	return managedwriter.NewClient(ctx, client.Config.ProjectId, client.Options...)
}

func NewTableReader(ctx context.Context, client *lib.Client) (*storage.BigQueryReadClient, error) {
	return storage.NewBigQueryReadClient(ctx, client.Options...)
}

func (s *TableService) GetTable(ctx context.Context, req *catalogv1beta1.GetTableRequest) (*catalogv1beta1.GetTableResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewDataset(inst.client, req.Schema).GetTable(ctx, req)
}

func (s *TableService) CreateTable(ctx context.Context, req *catalogv1beta1.CreateTableRequest) (*catalogv1beta1.CreateTableResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewDataset(inst.client, req.Schema).CreateTable(ctx, req)
}

func (s *TableService) UpdateTable(ctx context.Context, req *catalogv1beta1.UpdateTableRequest) (*catalogv1beta1.UpdateTableResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewDataset(inst.client, req.Schema).UpdateTable(ctx, req)
}

func (s *TableService) DeleteTable(ctx context.Context, req *catalogv1beta1.DeleteTableRequest) (*catalogv1beta1.DeleteTableResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewDataset(inst.client, req.Schema).DeleteTable(ctx, req)
}

func (s *TableService) ListTables(req *catalogv1beta1.ListTablesRequest, stream grpc.ServerStreamingServer[catalogv1beta1.ListTablesResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewDataset(inst.client, req.Schema).ListTables(stream)
}

func (s *TableService) ReadTable(req *catalogv1beta1.ReadTableRequest, stream grpc.ServerStreamingServer[catalogv1beta1.ReadTableResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	reader, err := NewTableReader(stream.Context(), inst.client)
	if err != nil {
		return err
	}

	return NewDataset(inst.client, req.Schema).ReadTable(req, stream, reader, &storagepb.ReadSession_TableReadOptions{
		SelectedFields: req.SelectedFields,
	})
}

func (s *TableService) WriteTables(stream grpc.ClientStreamingServer[catalogv1beta1.WriteTablesRequest, catalogv1beta1.WriteTablesResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	writer, err := NewTableWriter(ctx, inst.client)
	if err != nil {
		return err
	}
	defer writer.Close()

	var (
		writeStreams = make(map[string]*managedwriter.ManagedStream)
		writeBatches = map[string]writeRowsData{}
		writeResps   = map[string]*catalogv1beta1.WriteTableResponse{}
		tableDescs   = map[string]protoreflect.MessageDescriptor{}
		mut          sync.Mutex
		writeCh      = make(chan writeRowsBatch, 5)
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case batch := <-writeCh:
				start := time.Now()
				writeResp, err := batch.result.FullResponse(ctx)
				if err != nil {
					fmt.Printf("Write rows error: %s\n", err)
				}

				if writeResp != nil {
					if len(writeResp.RowErrors) > 0 {
						var newBatch writeRowsData
						for idx, row := range batch.rows {
							if !slices.ContainsFunc(writeResp.RowErrors, func(err *storagepb.RowError) bool {
								skip := int64(idx) == err.Index
								if skip {
									fmt.Printf("Write row error: %s\n", err.Message)
								}
								return skip
							}) {
								newBatch = append(newBatch, row)
							}
						}

						mut.Lock()
						writeBatches[batch.tableID] = append(newBatch, writeBatches[batch.tableID]...)
						mut.Unlock()
					} else {
						mut.Lock()
						batch.resp.RowsWritten += int64(len(batch.rows))
						fmt.Printf("Wrote %d rows in %s\n", int64(len(batch.rows)), time.Since(start))
						mut.Unlock()
					}
				}
			}
		}
	}()

	var nextBatch = func(tableID string, rows writeRowsData) error {
		start := time.Now()
		result, err := rows.Write(ctx, writeStreams[tableID], tableDescs[tableID])
		if err != nil {
			return err
		}

		writeCh <- writeRowsBatch{
			tableID:     tableID,
			writeStream: writeStreams[tableID],
			tableDesc:   tableDescs[tableID],
			resp:        writeResps[tableID],
			result:      result,
			rows:        rows,
		}

		log.Info(ctx, fmt.Sprintf("Submitted %d rows in %s\n", len(rows), time.Since(start)))
		return nil
	}

	defer func() {
		for _, ws := range writeStreams {
			ws.Close()
		}
	}()

	var streamErr error
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			streamErr = err
			break
		} else if req.GetTable() == nil {
			streamErr = fmt.Errorf("table cannot be nil")
			break
		} else if req.GetTable().GetSchema() == nil {
			streamErr = fmt.Errorf("schema cannot be nil")
			break
		} else if req.GetTable().GetSchema().GetCatalog() == nil {
			streamErr = fmt.Errorf("catalog cannot be nil")
			break
		} else if req.GetRow() == nil {
			streamErr = fmt.Errorf("row cannot be nil")
			break
		}

		var tableID = fmt.Sprintf("%s.%s.%s", req.Table.Schema.Catalog.Name, req.Table.Schema.Name, req.Table.Name)
		if _, ok := writeStreams[tableID]; !ok {
			table, err := NewDataset(inst.client, req.Table.Schema).NewTable(stream.Context(), req.Table.Name, false)
			if err != nil {
				streamErr = err
				break
			}

			desc, err := table.Descriptor(stream.Context())
			if err != nil {
				streamErr = err
				break
			}

			writeStream, err := table.NewWriteStream(ctx, writer)
			if err != nil {
				streamErr = err
				break
			}

			tableDescs[tableID] = desc
			writeStreams[tableID] = writeStream
			writeResps[tableID] = &catalogv1beta1.WriteTableResponse{
				Table: req.Table,
			}
		}

		row, err := NewWriteRow(req.Row)
		if err != nil {
			streamErr = err
			break
		}

		mut.Lock()
		writeBatches[tableID] = append(writeBatches[tableID], row)

		var (
			batchLen    int
			batchSize   int
			shouldWrite bool
		)
		for row := range slices.Values(writeBatches[tableID]) {
			if batchSize+row.Size() < MAX_WRITE_MSG_SIZE {
				batchSize += row.Size()
				batchLen++
			} else {
				shouldWrite = true
				break
			}
		}
		mut.Unlock()

		if shouldWrite {
			var batch writeRowsData

			mut.Lock()
			batch = slices.Clone(writeBatches[tableID][0:batchLen])
			writeBatches[tableID] = writeBatches[tableID][batchLen-1:]
			mut.Unlock()

			if err := nextBatch(tableID, batch); err != nil {
				streamErr = err
				break
			}
		}
	}

	if streamErr != nil {
		log.Error(ctx, fmt.Sprintf("Stream error: %s", streamErr))
	}
	log.Info(ctx, "Writing remaining rows...")

	finalCtx, finalCancel := context.WithTimeout(ctx, 5*time.Second)
	defer finalCancel()

	var ticker = time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			var remaining int
			for tableID := range writeBatches {
				mut.Lock()
				remaining += len(writeBatches[tableID])
				var rows = slices.Clone(writeBatches[tableID])
				writeBatches[tableID] = nil
				mut.Unlock()

				if len(rows) > 0 {
					if err := nextBatch(tableID, rows); err != nil {
						log.Error(ctx, "Remaining rows error", log.Err(err))
					}
				}
			}

			log.Error(ctx, "Remaining rows", slog.Int("count", remaining))
		case <-finalCtx.Done():
			return stream.SendAndClose(&catalogv1beta1.WriteTablesResponse{
				Tables: slices.Collect(maps.Values(writeResps)),
			})
		}
	}
}

func (t *Table) loadMeta(ctx context.Context) error {
	meta, err := t.table.Metadata(ctx)
	if err != nil {
		return err
	}
	t.meta = meta
	return nil
}

func (t *Table) Identifier(f bigquery.IdentifierFormat) (string, error) {
	return t.table.Identifier(f)
}

func (t *Table) Inserter() *bigquery.Inserter {
	return t.table.Inserter()
}

func (t *Table) Name() string {
	return t.table.TableID
}

func (t *Table) Description(ctx context.Context) string {
	if t.meta == nil {
		if err := t.loadMeta(ctx); err != nil {
			return ""
		}
	}

	return t.meta.Description
}

func (t *Table) Type(ctx context.Context) *string {
	if t.meta == nil {
		if err := t.loadMeta(ctx); err != nil {
			return nil
		}
	}

	tt := string(t.meta.Type)
	return &tt
}

func (t *Table) TotalRows(ctx context.Context) *int64 {
	if t.meta == nil {
		if err := t.loadMeta(ctx); err != nil {
			return nil
		}
	}

	tr := int64(t.meta.NumRows)
	return &tr
}

func (t *Table) TotalBytes(ctx context.Context) *int64 {
	if t.meta == nil {
		if err := t.loadMeta(ctx); err != nil {
			return nil
		}
	}

	nb := int64(t.meta.NumBytes)
	return &nb
}

func (t *Table) Descriptor(ctx context.Context) (protoreflect.MessageDescriptor, error) {
	if t.desc == nil {
		if t.meta == nil {
			if err := t.loadMeta(ctx); err != nil {
				return nil, err
			}
		}

		tableSchema, err := adapt.BQSchemaToStorageTableSchema(t.meta.Schema)
		if err != nil {
			return nil, err
		}

		schemaDescriptor, err := adapt.StorageSchemaToProto2Descriptor(tableSchema, "root")
		if err != nil {
			return nil, err
		}

		messageDescriptor, ok := schemaDescriptor.(protoreflect.MessageDescriptor)
		if !ok {
			return nil, fmt.Errorf("schema descriptor is not a message descriptor")
		}

		t.desc = messageDescriptor
	}

	return t.desc, nil
}

func (t *Table) NewWriteStream(ctx context.Context, writer *managedwriter.Client) (*managedwriter.ManagedStream, error) {
	tableName, err := t.Identifier(bigquery.StorageAPIResourceID)
	if err != nil {
		return nil, err
	}

	descriptor, err := t.Descriptor(ctx)
	if err != nil {
		return nil, err
	}

	return writer.NewManagedStream(ctx,
		managedwriter.WithDestinationTable(tableName),
		managedwriter.WithType(managedwriter.DefaultStream),
		managedwriter.WithSchemaDescriptor(protodesc.ToDescriptorProto(descriptor)),
	)
}

func (t *Table) Read(ctx context.Context, reader *storage.BigQueryReadClient, readRow readRowFunc, opts *storagepb.ReadSession_TableReadOptions) error {
	tableName, err := t.Identifier(bigquery.StorageAPIResourceID)
	if err != nil {
		return err
	}

	var (
		dataFormat = storagepb.DataFormat_AVRO
		sessionReq = &storagepb.CreateReadSessionRequest{
			Parent: fmt.Sprintf("projects/%s", t.table.ProjectID),
			ReadSession: &storagepb.ReadSession{
				Table:       tableName,
				DataFormat:  dataFormat,
				ReadOptions: opts,
			},
			PreferredMinStreamCount: int32(minReadStreams),
		}
	)

	session, err := reader.CreateReadSession(ctx, sessionReq, rpcOpts)
	if err != nil {
		return err
	}

	var streams = session.GetStreams()
	if len(streams) == 0 {
		return fmt.Errorf("no streams in session")
	}

	var (
		readCh    = make(chan *storagepb.ReadRowsResponse)
		streamWg  sync.WaitGroup
		processWg sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for readStream := range slices.Values(session.GetStreams()) {
		// Use a waitgroup to coordinate the reading and decoding goroutines.
		streamWg.Add(1)
		go func() {
			defer streamWg.Done()

			if err := processStream(ctx, reader, readStream.Name, readCh); err != nil {
				if err != io.EOF {
					fmt.Printf("processStream failure: %v\n", err)
					cancel()
				}
				return
			}
		}()
	}

	// Start Avro processing and decoding in another goroutine.
	processWg.Add(1)
	go func() {
		defer processWg.Done()

		if err := processAvro(ctx, session.GetAvroSchema().GetSchema(), readCh, readRow); err != nil {
			if err != io.EOF && !errors.Is(err, context.Canceled) {
				fmt.Printf("error processing avro: %v\n", err)
				cancel()
			}
			return
		}
	}()

	streamWg.Wait()
	cancel()
	processWg.Wait()

	return nil
}

func processStream(ctx context.Context, client *storage.BigQueryReadClient, stream string, readCh chan<- *storagepb.ReadRowsResponse) error {
	var offset int64

	// Streams may be long-running.  Rather than using a global retry for the
	// stream, implement a retry that resets once progress is made.
	retryLimit := 3
	retries := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Send the initiating request to start streaming row blocks.
			rowStream, err := client.ReadRows(ctx, &storagepb.ReadRowsRequest{
				ReadStream: stream,
				Offset:     offset,
			}, rpcOpts)
			if err != nil {
				return fmt.Errorf("couldn't invoke ReadRows: %w", err)
			}

			// Process the streamed resonses.
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					r, err := rowStream.Recv()
					if err != nil {
						if errors.Is(err, io.EOF) {
							return err
						}

						// If there is an error, check whether it is a retryable
						// error with a retry delay and sleep instead of increasing
						// retries count.
						var retryDelayDuration time.Duration
						if errorStatus, ok := status.FromError(err); ok && errorStatus.Code() == codes.ResourceExhausted {
							for _, detail := range errorStatus.Details() {
								retryInfo, ok := detail.(*errdetails.RetryInfo)
								if !ok {
									continue
								}
								var retryDelay = retryInfo.GetRetryDelay()
								retryDelayDuration = time.Duration(retryDelay.Seconds)*time.Second + time.Duration(retryDelay.Nanos)*time.Nanosecond
								break
							}
						}
						if retryDelayDuration != 0 {
							log.Info(ctx, fmt.Sprintf("processStream failed with a retryable error, retrying in %v", retryDelayDuration))
							time.Sleep(retryDelayDuration)
						} else {
							retries++
							if retries >= retryLimit {
								return fmt.Errorf("processStream retries exhausted: %w", err)
							}
						}
						// break the inner loop, and try to recover by starting a new streaming
						// ReadRows call at the last known good offset.
						break
					} else {
						// Reset retries after a successful resonse.
						retries = 0
					}

					var rc = r.GetRowCount()
					if rc > 0 {
						// Bookmark our progress in case of retries and send the rowblock on the channel.
						offset = offset + rc
						// We're making progress, reset retries.
						retries = 0
						readCh <- r
					}
				}
			}
		}
	}
}

func processAvro(ctx context.Context, schema string, readCh <-chan *storagepb.ReadRowsResponse, readRow readRowFunc) error {
	// Establish a decoder that can process blocks of messages using the
	// reference schema. All blocks share the same schema, so the decoder
	// can be long-lived.
	codec, err := goavro.NewCodec(schema)
	if err != nil {
		return fmt.Errorf("couldn't create codec: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			// Context was cancelled.  Stop.
			return ctx.Err()
		case rowsRes, ok := <-readCh:
			if !ok {
				// Channel closed, no further avro messages.  Stop.
				return nil
			}

			var bytes = rowsRes.GetAvroRows().GetSerializedBinaryRows()
			for len(bytes) > 0 {
				row, remainingBytes, err := codec.NativeFromBinary(bytes)

				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					return fmt.Errorf("decoding error with %d bytes remaining: %v", len(bytes), err)
				}

				if rowMap, ok := row.(map[string]any); ok {
					if err := readRow(flattenAvroRow(rowMap)); err != nil {
						return err
					}
				} else {
					return fmt.Errorf("expected row map, got: %T", row)
				}

				bytes = remainingBytes
			}
		}
	}
}

func flattenAvroRow(row map[string]any) map[string]any {
	for k, v := range row {
		row[k] = flattenAvroData(v)
	}
	return row
}

func flattenAvroData(data any) any {
	switch v := data.(type) {
	case map[string]any:
		// Check for union-type wrapper
		if len(v) == 1 {
			for _, val := range v {
				// This is a union-type value like {"string": "value"}
				return flattenAvroData(val)
			}
		}
		// Normal object with nested fields
		flattened := make(map[string]interface{})
		for key, val := range v {
			flattened[key] = flattenAvroData(val)
		}
		return flattened

	case []any:
		// Flatten each element in the array
		for i, item := range v {
			v[i] = flattenAvroData(item)
		}
		return v

	default:
		// Primitive types (string, int, etc.)
		return v
	}
}

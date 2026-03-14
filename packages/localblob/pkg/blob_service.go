package pkg

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path"

	"gocloud.dev/blob"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	grpc "google.golang.org/grpc"

	// localblobv1beta1 "github.com/datakit-dev/dtkt-integrations/localblob/pkg/proto/localblob/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/encoding"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	blobv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/blob/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
)

const (
	DefaultLineDelim = "\n"
	DefaultPageSize  = 10
)

type (
	BlobService struct {
		blobv1beta1.UnimplementedBlobServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
)

func NewBlobService(mux v1beta1.InstanceMux[*Instance]) *BlobService {
	return &BlobService{
		mux: mux,
	}
}

func (s *BlobService) ListBlobs(ctx context.Context, req *blobv1beta1.ListBlobsRequest) (*blobv1beta1.ListBlobsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	var (
		pageToken []byte
		pageSize  int
	)
	if req.PageToken == "" {
		pageToken = blob.FirstPageToken
	} else {
		pageToken = []byte(req.PageToken)
	}

	if req.PageSize <= 0 {
		pageSize = DefaultPageSize
	} else {
		pageSize = int(req.PageSize)
	}

	bucket, err := GetBucket(&inst.buckets, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, nextPage, err := bucket.ListPage(ctx, pageToken, pageSize, &blob.ListOptions{
		Prefix: req.GetPrefix(),
	})
	if err != nil {
		return nil, err
	}

	return &blobv1beta1.ListBlobsResponse{
		Results: util.SliceMap(resp, func(r *blob.ListObject) *blobv1beta1.ListResult {
			return &blobv1beta1.ListResult{
				Key:     r.Key,
				ModTime: timestamppb.New(r.ModTime),
				Size:    r.Size,
				Md5:     string(r.MD5),
				IsDir:   r.IsDir,
			}
		}),
		NextPageToken: string(nextPage),
	}, nil
}

func (s *BlobService) GetBlob(ctx context.Context, req *blobv1beta1.GetBlobRequest) (*blobv1beta1.GetBlobResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	bucket, err := GetBucket(&inst.buckets, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	attrs, err := GetBlob(ctx, bucket, req.Key)
	if err != nil {
		return nil, err
	}

	return &blobv1beta1.GetBlobResponse{
		Key:                req.Key,
		CacheControl:       attrs.CacheControl,
		ContentDisposition: attrs.ContentDisposition,
		ContentEncoding:    attrs.ContentEncoding,
		ContentLanguage:    attrs.ContentLanguage,
		ContentType:        attrs.ContentType,
		Metadata:           attrs.Metadata,
		CreateTime:         timestamppb.New(attrs.CreateTime),
		ModTime:            timestamppb.New(attrs.ModTime),
		Size:               attrs.Size,
		Md5:                hex.EncodeToString(attrs.MD5),
		Etag:               attrs.ETag,
	}, nil
}

func (s *BlobService) WriteBlobLines(stream grpc.ClientStreamingServer[blobv1beta1.WriteBlobLinesRequest, blobv1beta1.WriteBlobLinesResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	var writers util.SyncMap[string, *BlobLineWriter]
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		bucket, err := GetBucket(&inst.buckets, req)
		if err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}

		key := path.Join(req.Bucket, req.Key)
		lineWriter, ok := writers.Load(key)
		if !ok {
			writer, err := bucket.NewWriter(ctx, req.Key, &blob.WriterOptions{})
			if err != nil {
				return err
			}
			defer func() {
				if err := writer.Close(); err != nil {
					log.Error(ctx, "failed to close blob writer", log.Err(err))
				}
			}()

			lineWriter = &BlobLineWriter{
				Writer: writer,
				bucket: bucket,
				key:    req.Key,
			}

			writers.Store(key, lineWriter)
		}

		_, err = fmt.Fprintln(lineWriter, string(req.Data))
		if err != nil {
			return err
		}

		lineWriter.linesWritten++
	}

	var results []*blobv1beta1.WriteBlobLinesResponse_KeyResult
	for _, writer := range writers.Values() {
		err := writer.Close()
		if err != nil {
			return err
		}

		result, err := writer.Result(ctx)
		if err != nil {
			return err
		}
		results = append(results, result)
	}

	return stream.SendAndClose(&blobv1beta1.WriteBlobLinesResponse{
		Results: results,
	})
}

func (s *BlobService) WriteBlob(ctx context.Context, req *blobv1beta1.WriteBlobRequest) (*blobv1beta1.WriteBlobResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	bucket, err := GetBucket(&inst.buckets, req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Data) > 0 {
		writer, err := bucket.NewWriter(ctx, req.Key, &blob.WriterOptions{
			ContentType:        req.ContentType,
			ContentLanguage:    req.ContentLanguage,
			ContentEncoding:    req.ContentEncoding,
			CacheControl:       req.CacheControl,
			ContentDisposition: req.ContentDisposition,
			Metadata:           req.Metadata,
			IfNotExist:         req.IfNotExist,
			// IfMatch:            req.IfMatch,
			// IfNoneMatch:        req.IfNoneMatch,
		})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		defer func() {
			if err := writer.Close(); err != nil {
				log.Error(ctx, "failed to close blob writer", log.Err(err))
			}
		}()

		// If decoding string represenation of bytes fails, then the input was probably a string and not bytes
		_, err = base64.StdEncoding.DecodeString(string(req.Data))
		if err != nil {
			req.Data = []byte(string(req.Data))
		}

		_, err = writer.Write(req.Data)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if err := writer.Close(); err != nil {
			log.Error(ctx, "failed to close blob writer", log.Err(err))
		}

		blob, err := s.GetBlob(ctx, &blobv1beta1.GetBlobRequest{
			Key:    req.Key,
			Bucket: req.Bucket,
		})
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		return &blobv1beta1.WriteBlobResponse{
			Key:     blob.Key,
			Size:    blob.Size,
			ModTime: blob.ModTime,
			Etag:    blob.Etag,
		}, nil
	}

	return nil, status.Errorf(codes.InvalidArgument, "no data provided")
}

func (s *BlobService) DeleteBlob(ctx context.Context, req *blobv1beta1.DeleteBlobRequest) (*blobv1beta1.DeleteBlobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteBlob not implemented")
}

func (s *BlobService) CopyBlob(ctx context.Context, req *blobv1beta1.CopyBlobRequest) (*blobv1beta1.CopyBlobResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CopyBlob not implemented")
}

func (s *BlobService) GenerateSignedURL(ctx context.Context, req *blobv1beta1.GenerateSignedURLRequest) (*blobv1beta1.GenerateSignedURLResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GenerateSignedURL not implemented")
}

func (s *BlobService) ReadBlobLines(req *blobv1beta1.ReadBlobLinesRequest, stream grpc.ServerStreamingServer[blobv1beta1.ReadBlobLinesResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	bucket, err := GetBucket(&inst.buckets, req)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if exists, err := bucket.Exists(ctx, req.Key); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if !exists {
		return status.Errorf(codes.InvalidArgument, "blob not found for key: %s", req.Key)
	}

	reader, err := bucket.NewReader(ctx, req.Key, nil)
	if err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error(ctx, "failed to close blob reader", log.Err(err))
		}
	}()

	if req.Delimiter == "" {
		req.Delimiter = DefaultLineDelim
	}

	if req.LineFormat <= 0 {
		req.LineFormat = blobv1beta1.LineFormat_LINE_FORMAT_RAW
	}

	log.Info(ctx, "reading blob lines", slog.String("key", req.Key), slog.String("line_format", req.LineFormat.String()), slog.String("delimiter", req.Delimiter))

	switch req.LineFormat {
	case blobv1beta1.LineFormat_LINE_FORMAT_RAW:
		var (
			scanner = bufio.NewScanner(reader)
			data    []byte
		)

		scanner.Split(encoding.DelimSplitFunc(req.Delimiter))

		for scanner.Scan() {
			data = scanner.Bytes()
			if len(data) == 0 || (req.MaxSize > 0 && len(data) > int(req.MaxSize)) {
				continue
			}

			err = scanner.Err()
			if err != nil {
				log.Error(ctx, "scanner error", log.Err(err))
				return status.Errorf(codes.Aborted, "scanner error: %v", err)
			}

			res := &blobv1beta1.ReadBlobLinesResponse{
				Key: req.Key,
				Line: &blobv1beta1.ReadBlobLinesResponse_Data{
					Data: data,
				},
			}

			err = stream.Send(res)
			if err != nil {
				log.Error(ctx, "stream send error", log.Err(err))
				return status.Errorf(codes.Aborted, "failed to send stream response: %v", err)
			}
		}
	case blobv1beta1.LineFormat_LINE_FORMAT_JSON:
		var (
			decode = encoding.NewJSONDecoderV2(
				encoding.WithDecodeDelim(req.Delimiter),
				encoding.WithDecodeJSONStream(req.JsonArrayMode),
			).StreamDecode(reader)
			json *structpb.Value
		)

		for {
			json = &structpb.Value{}
			err = decode(json)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return status.Error(codes.OK, err.Error())
				}
				log.Error(ctx, "json decode error", log.Err(err))
				return status.Errorf(codes.Aborted, "json decode error: %v", err)
			}

			res := &blobv1beta1.ReadBlobLinesResponse{
				Key: req.Key,
				Line: &blobv1beta1.ReadBlobLinesResponse_Json{
					Json: json,
				},
			}

			err = stream.Send(res)
			if err != nil {
				log.Error(ctx, "stream send error", log.Err(err))
				return status.Errorf(codes.Aborted, "failed to send stream response: %v", err)
			}
		}
	}

	return nil
}

func (s *BlobService) ReadBlobRange(req *blobv1beta1.ReadBlobRangeRequest, stream grpc.ServerStreamingServer[blobv1beta1.ReadBlobRangeResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	bucket, err := GetBucket(&inst.buckets, req)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if exists, err := bucket.Exists(ctx, req.Key); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	} else if !exists {
		return status.Errorf(codes.InvalidArgument, "blob not found for key: %s", req.Key)
	}

	reader, err := bucket.NewRangeReader(ctx, req.Key, int64(req.Offset), req.Length, nil)
	if err != nil {
		return status.Error(codes.Aborted, err.Error())
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Error(ctx, "failed to close blob reader", log.Err(err))
		}
	}()

	var buf = make([]byte, 1024*64)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			if err := stream.Send(&blobv1beta1.ReadBlobRangeResponse{
				Key:  req.Key,
				Data: buf[:n],
			}); err != nil {
				return status.Error(codes.Aborted, err.Error())
			}
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				return status.Error(codes.OK, err.Error())
			}
			return status.Error(codes.Aborted, err.Error())
		}
	}
}

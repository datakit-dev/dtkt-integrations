package v1beta1

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/bigquery"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib"
	"github.com/datakit-dev/dtkt-integrations/bigquery/pkg/lib/zetasql"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	catalogv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/catalog/v1beta1"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const QueryDialect = "GoogleSQL"

type (
	QueryService struct {
		catalogv1beta1.UnimplementedQueryServiceServer
		mux v1beta1.InstanceMux[*Instance]
	}
	Query struct {
		client     *lib.Client
		sql        string
		valid      bool
		err        error
		params     v1beta1.Params
		schema     bigquery.Schema
		fields     v1beta1.Fields
		totalBytes int64
		result     QueryResults
	}
)

func NewQueryService(mux v1beta1.InstanceMux[*Instance]) *QueryService {
	return &QueryService{
		mux: mux,
	}
}

func NewQuery(client *lib.Client, sql string, params v1beta1.Params) *Query {
	return &Query{client: client, sql: sql, params: params}
}

func (s *QueryService) GetQueryDialect(context.Context, *catalogv1beta1.GetQueryDialectRequest) (*catalogv1beta1.GetQueryDialectResponse, error) {
	return v1beta1.NewQueryDialectResponse(QueryDialect), nil
}

func (s *QueryService) ValidateQuery(ctx context.Context, req *catalogv1beta1.ValidateQueryRequest) (*catalogv1beta1.ValidateQueryResponse, error) {
	v, err := zetasql.
		NewValidation(req.Query, req.Params).
		ValidateQuery(req.Accessible...)
	if err != nil {
		if !zetasql.IgnoreError(err) {
			return nil, err
		}
	}

	return v1beta1.NewValidateQueryRes(
		v1beta1.NewQuery(
			v.Dialect(),
			v.Query(),
			v1beta1.WithQueryValid(v.Valid()),
			v1beta1.WithQueryError(v.Error()),
			v1beta1.WithQueryFields(v.Fields()...),
			v1beta1.WithQueryParams(v.Params()...),
		),
		v.Accessed(),
	), nil
}

func (s *QueryService) GetQueryResults(ctx context.Context, req *catalogv1beta1.GetQueryResultsRequest) (*catalogv1beta1.GetQueryResultsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	return NewQuery(inst.client, req.Query, req.Params).ResultsPage(ctx, req.PageSize, req.Page)
}

func (s *QueryService) StreamQueryResults(req *catalogv1beta1.StreamQueryResultsRequest, stream grpc.ServerStreamingServer[catalogv1beta1.StreamQueryResultsResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	var (
		query   = NewQuery(inst.client, req.Query, req.Params)
		sendRow = func(result resultRow) error {
			row, err := result.Row()
			if err != nil {
				return err
			}

			return stream.Send(&catalogv1beta1.StreamQueryResultsResponse{
				Query: v1beta1.NewQuery(
					query.Dialect(),
					query.Query(),
					v1beta1.WithQueryValid(query.Valid()),
					v1beta1.WithQueryError(query.Error()),
					v1beta1.WithQueryFields(query.Fields()...),
					v1beta1.WithQueryParams(query.Params()...),
				),
				Row: row,
			})
		}
	)

	return query.StreamResults(stream.Context(), sendRow)
}

func (q *Query) Dialect() string {
	return QueryDialect
}

func (q *Query) runQuery(ctx context.Context, query *bigquery.Query) (*bigquery.Job, error) {
	job, err := query.Run(ctx)
	if err != nil {
		q.valid = false
		q.err = err
		return job, err
	}

	q.valid = true
	return job, err
}

func (q *Query) loadSchema(schema bigquery.Schema) {
	if schema != nil {
		q.schema = schema
		q.fields = FieldsToProto(schema)
	}
}

func (q *Query) loadStats(status *bigquery.JobStatus) error {
	stats := status.Statistics
	if stats == nil {
		q.valid = false
		q.err = fmt.Errorf("query statistics not found")
		return q.err
	}

	q.totalBytes = stats.TotalBytesProcessed
	if details, ok := stats.Details.(*bigquery.QueryStatistics); ok {
		q.loadSchema(details.Schema)
	}

	return nil
}

func (q *Query) prepareQuery(dryRun bool) (*bigquery.Query, error) {
	query := q.client.Client.Query(q.sql)
	query.DryRun = dryRun

	if len(q.params) > 0 {
		queryParams, err := ParamsFromProto(q.params)
		if err != nil {
			return nil, err
		}

		query.Parameters = queryParams
	}
	return query, nil
}

func (q *Query) runValidation(ctx context.Context) error {
	query, err := q.prepareQuery(true)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	job, err := q.runQuery(ctx, query)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	status := job.LastStatus()
	if err := status.Err(); err != nil {
		q.valid = false
		q.err = err
		return err
	}

	if err := q.loadStats(status); err != nil {
		q.valid = false
		q.err = err
		return err
	}

	return nil
}

func (q *Query) StreamResults(ctx context.Context, getRow readRowFunc) error {
	query, err := q.prepareQuery(false)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	job, err := q.runQuery(ctx, query)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	} else if job == nil {
		return fmt.Errorf("query job not found")
	}

	status, err := job.Wait(ctx)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	if status.Err() != nil {
		q.valid = false
		q.err = status.Err()
		return q.err
	}

	it, err := job.Read(ctx)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			var row resultRow
			err = it.Next(&row)
			if err != nil {
				if err == iterator.Done {
					return io.EOF
				}
				return err
			}

			if len(q.schema) != len(it.Schema) {
				q.loadSchema(it.Schema)
			}

			err = getRow(row)
			if err != nil {
				return err
			}
		}
	}
}

func (q *Query) LoadResults(ctx context.Context, pageSize int, pageEncoded string) error {
	var (
		job      *bigquery.Job
		currPage *queryPage
	)

	if pageEncoded != "" {
		currPage, q.err = decodeResultPage[queryPage](pageEncoded)
		if q.err != nil {
			q.valid = false
			return q.err
		}

		job, q.err = q.client.JobFromID(ctx, currPage.JobID)
		if q.err != nil {
			q.valid = false
			return q.err
		} else if job == nil {
			return fmt.Errorf("query job not found")
		}
	} else {
		query, err := q.prepareQuery(false)
		if err != nil {
			q.valid = false
			q.err = err
			return err
		}

		job, err = q.runQuery(ctx, query)
		if err != nil {
			q.valid = false
			q.err = err
			return err
		} else if job == nil {
			return fmt.Errorf("query job not found")
		}
	}

	status, err := job.Wait(ctx)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}
	if status.Err() != nil {
		q.valid = false
		q.err = status.Err()
		return q.err
	}

	it, err := job.Read(ctx)
	if err != nil {
		q.valid = false
		q.err = err
		return err
	}

	q.result = QueryResults{pageSize: pageSize}
	q.result.totalRows = int64(it.TotalRows)

	it.PageInfo().MaxSize = pageSize

	if currPage != nil {
		it.PageInfo().Token = currPage.PageToken

		if currPage.PrevPage != nil {
			q.result.prevPage = currPage.PrevPage
		}
	}

	for {
		err = it.Next(&q.result)
		nextPageToken := it.PageInfo().Token

		// if len(q.schema) != len(it.Schema) {
		q.loadSchema(it.Schema)
		// }

		if q.result.totalRows != int64(it.TotalRows) {
			q.result.totalRows = int64(it.TotalRows)
		}

		if nextPageToken != "" {
			q.result.nextPage = newQueryPage(job.ID(), nextPageToken, currPage)
		}

		if len(q.result.rows) == pageSize {
			return nil
		}

		if err != nil {
			if err == iterator.Done {
				return nil
			}
			return err
		}
	}
}

func (q *Query) ResultsPage(ctx context.Context, pageSize int32, pageEncoded string) (*catalogv1beta1.GetQueryResultsResponse, error) {
	q.err = q.LoadResults(ctx, int(pageSize), pageEncoded)
	if q.err != nil {
		return nil, q.err
	}

	var (
		errMsg string
		// rows               []*structpb.Struct
		prevPage, nextPage string
	)

	// rows, q.err = q.result.Rows()
	if q.err != nil {
		q.valid = false
		errMsg = q.err.Error()
	} else {
		if q.result.PrevPage() != nil {
			prevPage = *q.result.PrevPage()
		}
		if q.result.NextPage() != nil {
			nextPage = *q.result.NextPage()
		}
	}

	return &catalogv1beta1.GetQueryResultsResponse{
		Query: v1beta1.NewQuery(q.Dialect(), q.Query(),
			v1beta1.WithQueryError(errMsg),
			v1beta1.WithQueryFields(q.Fields()...),
			v1beta1.WithQueryParams(q.Params()...),
		),
		ResultsPage: &catalogv1beta1.QueryResultsPage{
			PrevPage:   prevPage,
			NextPage:   nextPage,
			TotalPages: q.result.TotalPages(),
			TotalRows:  q.result.TotalRows(),
			RowsCount:  q.result.RowsCount(),
			Rows:       q.result.Rows(),
		},
	}, nil
}

func (q *Query) Query() string {
	return q.sql
}

func (q *Query) Valid() bool {
	return q.valid
}

func (q *Query) Err() error {
	return q.err
}

func (q *Query) Error() string {
	if q.err != nil {
		return q.err.Error()
	}
	return ""
}

func (q *Query) Params() v1beta1.Params {
	return v1beta1.Params(q.params)
}

func (q *Query) Fields() v1beta1.Fields {
	return q.fields
}

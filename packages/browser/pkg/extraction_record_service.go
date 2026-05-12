package pkg

import (
	"context"
	"log/slog"
	"time"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/extractionrecord"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/predicate"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/protostoresdk/entadapter"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ExtractionRecordService struct {
	browserv1beta.UnimplementedExtractionRecordServiceServer
	mux   v1beta1.InstanceMux[*Instance]
	log   *slog.Logger
	db    *ent.Client
	pager entadapter.PaginateOptions[*ent.ExtractionRecordQuery, *ent.ExtractionRecord, extractionrecord.OrderOption, predicate.ExtractionRecord]
}

func NewExtractionRecordService(mux v1beta1.InstanceMux[*Instance]) (*ExtractionRecordService, error) {
	dataDir, err := mux.GetDataRoot()
	if err != nil {
		return nil, err
	}

	dbClient, err := db.GetClient(context.Background(), mux.Logger(), dataDir)
	if err != nil {
		return nil, err
	}

	return &ExtractionRecordService{
		mux: mux,
		db:  dbClient,
		pager: entadapter.PaginateOptions[*ent.ExtractionRecordQuery, *ent.ExtractionRecord, extractionrecord.OrderOption, predicate.ExtractionRecord]{
			IDField:   extractionrecord.FieldRowid,
			TimeField: extractionrecord.FieldUpdateTime,
			GetID: func(r *ent.ExtractionRecord) int64 {
				return r.Rowid
			},
			GetTime: func(r *ent.ExtractionRecord) time.Time {
				return r.UpdateTime
			},
		},
	}, nil
}

func ExtractionRecordToProto(r *ent.ExtractionRecord) *browserv1beta.ExtractionRecord {
	proto := &browserv1beta.ExtractionRecord{
		Id:         r.ID.String(),
		SchemaId:   r.SchemaID.String(),
		TaskId:     r.TaskID.String(),
		CreateTime: timestamppb.New(r.CreateTime),
		UpdateTime: timestamppb.New(r.UpdateTime),
	}
	proto.Values = r.Values
	proto.Captures = r.Captures
	return proto
}

func (s *ExtractionRecordService) CreateExtractionRecord(ctx context.Context, req *browserv1beta.CreateExtractionRecordRequest) (*browserv1beta.CreateExtractionRecordResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	p := req.GetRecord()

	schemaID, err := uuid.Parse(p.GetSchemaId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid schema_id: "+err.Error())
	}

	taskID, err := uuid.Parse(p.GetTaskId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid task_id: "+err.Error())
	}

	create := s.db.ExtractionRecord.Create().
		SetSchemaID(schemaID).
		SetTaskID(taskID).
		SetValues(p.GetValues()).
		SetCaptures(p.GetCaptures()).
		SetExtensionID(inst.GetExtensionId())

	if id := p.GetId(); id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		create = create.SetID(uid)
	}

	created, err := create.Save(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.CreateExtractionRecordResponse{Record: ExtractionRecordToProto(created)}, nil
}

func (s *ExtractionRecordService) GetExtractionRecord(ctx context.Context, req *browserv1beta.GetExtractionRecordRequest) (*browserv1beta.GetExtractionRecordResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	record, err := s.db.ExtractionRecord.Query().Where(
		extractionrecord.ID(id),
		extractionrecord.ExtensionID(inst.GetExtensionId()),
	).Only(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.GetExtractionRecordResponse{Record: ExtractionRecordToProto(record)}, nil
}

func (s *ExtractionRecordService) ListExtractionRecords(ctx context.Context, req *browserv1beta.ListExtractionRecordsRequest) (*browserv1beta.ListExtractionRecordsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	preds := []predicate.ExtractionRecord{
		extractionrecord.ExtensionID(inst.GetExtensionId()),
	}

	if sid := req.GetSchemaId(); sid != "" {
		id, err := uuid.Parse(sid)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid schema_id: "+err.Error())
		}
		preds = append(preds, extractionrecord.SchemaID(id))
	}

	if tid := req.GetTaskId(); tid != "" {
		id, err := uuid.Parse(tid)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid task_id: "+err.Error())
		}
		preds = append(preds, extractionrecord.TaskID(id))
	}

	nextPageToken, records, err := s.pager.GetNextPage(ctx, req, s.db.ExtractionRecord.Query().Where(preds...))
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.ListExtractionRecordsResponse{
		Records:       util.SliceMap(records, ExtractionRecordToProto),
		NextPageToken: nextPageToken,
	}, nil
}

func (s *ExtractionRecordService) UpdateExtractionRecord(ctx context.Context, req *browserv1beta.UpdateExtractionRecordRequest) (*browserv1beta.UpdateExtractionRecordResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	p := req.GetRecord()
	id, err := uuid.Parse(p.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	upd := s.db.ExtractionRecord.UpdateOneID(id).Where(
		extractionrecord.ExtensionID(inst.GetExtensionId()),
	)

	applyPath := func(path string) {
		switch path {
		case "values":
			upd.SetValues(p.GetValues())
		case "captures":
			upd.SetCaptures(p.GetCaptures())
		}
	}

	mask := req.GetUpdateMask()
	if mask == nil || len(mask.GetPaths()) == 0 {
		applyPath("values")
		applyPath("captures")
	} else {
		for _, path := range mask.GetPaths() {
			applyPath(path)
		}
	}

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.UpdateExtractionRecordResponse{Record: ExtractionRecordToProto(updated)}, nil
}

func (s *ExtractionRecordService) DeleteExtractionRecord(ctx context.Context, req *browserv1beta.DeleteExtractionRecordRequest) (*browserv1beta.DeleteExtractionRecordResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.db.ExtractionRecord.DeleteOneID(id).Where(
		extractionrecord.ExtensionID(inst.GetExtensionId()),
	).Exec(ctx); err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.DeleteExtractionRecordResponse{}, nil
}

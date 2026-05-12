package pkg

import (
	"context"
	"log/slog"
	"time"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/extractionschema"
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

type ExtractionSchemaService struct {
	browserv1beta.UnimplementedExtractionSchemaServiceServer
	mux   v1beta1.InstanceMux[*Instance]
	log   *slog.Logger
	db    *ent.Client
	pager entadapter.PaginateOptions[*ent.ExtractionSchemaQuery, *ent.ExtractionSchema, extractionschema.OrderOption, predicate.ExtractionSchema]
}

func NewExtractionSchemaService(mux v1beta1.InstanceMux[*Instance]) (*ExtractionSchemaService, error) {
	dataDir, err := mux.GetDataRoot()
	if err != nil {
		return nil, err
	}

	dbClient, err := db.GetClient(context.Background(), mux.Logger(), dataDir)
	if err != nil {
		return nil, err
	}

	return &ExtractionSchemaService{
		mux: mux,
		db:  dbClient,
		pager: entadapter.PaginateOptions[*ent.ExtractionSchemaQuery, *ent.ExtractionSchema, extractionschema.OrderOption, predicate.ExtractionSchema]{
			IDField:   extractionschema.FieldRowid,
			TimeField: extractionschema.FieldUpdateTime,
			GetID: func(s *ent.ExtractionSchema) int64 {
				return s.Rowid
			},
			GetTime: func(s *ent.ExtractionSchema) time.Time {
				return s.UpdateTime
			},
		},
	}, nil
}

func ExtractionSchemaToProto(s *ent.ExtractionSchema) *browserv1beta.ExtractionSchema {
	proto := &browserv1beta.ExtractionSchema{
		Id:         s.ID.String(),
		Name:       s.Name,
		CreateTime: timestamppb.New(s.CreateTime),
		UpdateTime: timestamppb.New(s.UpdateTime),
	}
	if s.Description != nil {
		proto.Description = *s.Description
	}
	proto.Fields = s.Fields
	return proto
}

func (s *ExtractionSchemaService) CreateExtractionSchema(ctx context.Context, req *browserv1beta.CreateExtractionSchemaRequest) (*browserv1beta.CreateExtractionSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	p := req.GetSchema()

	create := s.db.ExtractionSchema.Create().
		SetName(p.GetName()).
		SetFields(p.GetFields()).
		SetExtensionID(inst.GetExtensionId())

	if desc := p.GetDescription(); desc != "" {
		create = create.SetDescription(desc)
	}

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

	return &browserv1beta.CreateExtractionSchemaResponse{Schema: ExtractionSchemaToProto(created)}, nil
}

func (s *ExtractionSchemaService) GetExtractionSchema(ctx context.Context, req *browserv1beta.GetExtractionSchemaRequest) (*browserv1beta.GetExtractionSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	schema, err := s.db.ExtractionSchema.Query().Where(
		extractionschema.ID(id),
		extractionschema.ExtensionID(inst.GetExtensionId()),
	).Only(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.GetExtractionSchemaResponse{Schema: ExtractionSchemaToProto(schema)}, nil
}

func (s *ExtractionSchemaService) ListExtractionSchemas(ctx context.Context, req *browserv1beta.ListExtractionSchemasRequest) (*browserv1beta.ListExtractionSchemasResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	preds := []predicate.ExtractionSchema{
		extractionschema.ExtensionID(inst.GetExtensionId()),
	}

	nextPageToken, schemas, err := s.pager.GetNextPage(ctx, req, s.db.ExtractionSchema.Query().Where(preds...))
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.ListExtractionSchemasResponse{
		Schemas:       util.SliceMap(schemas, ExtractionSchemaToProto),
		NextPageToken: nextPageToken,
	}, nil
}

func (s *ExtractionSchemaService) UpdateExtractionSchema(ctx context.Context, req *browserv1beta.UpdateExtractionSchemaRequest) (*browserv1beta.UpdateExtractionSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	p := req.GetSchema()
	id, err := uuid.Parse(p.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	upd := s.db.ExtractionSchema.UpdateOneID(id).Where(
		extractionschema.ExtensionID(inst.GetExtensionId()),
	)

	applyPath := func(path string) {
		switch path {
		case "name":
			upd.SetName(p.GetName())
		case "description":
			if desc := p.GetDescription(); desc != "" {
				upd.SetDescription(desc)
			} else {
				upd.ClearDescription()
			}
		case "fields":
			upd.SetFields(p.GetFields())
		}
	}

	mask := req.GetUpdateMask()
	if mask == nil || len(mask.GetPaths()) == 0 {
		for _, path := range []string{"name", "description", "fields"} {
			applyPath(path)
		}
	} else {
		for _, path := range mask.GetPaths() {
			applyPath(path)
		}
	}

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.UpdateExtractionSchemaResponse{Schema: ExtractionSchemaToProto(updated)}, nil
}

func (s *ExtractionSchemaService) DeleteExtractionSchema(ctx context.Context, req *browserv1beta.DeleteExtractionSchemaRequest) (*browserv1beta.DeleteExtractionSchemaResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.db.ExtractionSchema.DeleteOneID(id).Where(
		extractionschema.ExtensionID(inst.GetExtensionId()),
	).Exec(ctx); err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.DeleteExtractionSchemaResponse{}, nil
}

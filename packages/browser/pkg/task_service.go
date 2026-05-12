package pkg

import (
	"context"
	"log/slog"
	"time"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/predicate"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/task"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/protostoresdk/entadapter"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskService struct {
	browserv1beta.UnimplementedTaskServiceServer
	mux    v1beta1.InstanceMux[*Instance]
	log    *slog.Logger
	db     *ent.Client
	broker *TaskBroker
	pager  entadapter.PaginateOptions[*ent.TaskQuery, *ent.Task, task.OrderOption, predicate.Task]
}

func NewTaskService(mux v1beta1.InstanceMux[*Instance]) (*TaskService, error) {
	dataDir, err := mux.GetDataRoot()
	if err != nil {
		return nil, err
	}

	db, err := db.GetClient(context.Background(), mux.Logger(), dataDir)
	if err != nil {
		return nil, err
	}

	return &TaskService{
		mux:    mux,
		db:     db,
		broker: NewTaskBroker(context.Background(), db),
		pager: entadapter.PaginateOptions[*ent.TaskQuery, *ent.Task, task.OrderOption, predicate.Task]{
			IDField:   task.FieldRowid,
			TimeField: task.FieldUpdateTime,
			GetID: func(task *ent.Task) int64 {
				return task.Rowid
			},
			GetTime: func(task *ent.Task) time.Time {
				return task.UpdateTime
			},
		},
	}, nil
}

func TaskToProto(task *ent.Task) *browserv1beta.Task {
	var completeTime *timestamppb.Timestamp
	if task.CompleteTime != nil {
		completeTime = timestamppb.New(*task.CompleteTime)
	}

	extraction := &browserv1beta.ExtractionTask{}
	if task.SchemaID != nil {
		extraction.SchemaId = task.SchemaID.String()
	}
	if task.RecordID != nil {
		extraction.RecordId = task.RecordID.String()
	}

	return &browserv1beta.Task{
		Id:    task.ID.String(),
		Title: task.Title,
		Url:   task.URL,
		State: task.State,
		Payload: &browserv1beta.Task_Extraction{
			Extraction: extraction,
		},
		CreateTime:   timestamppb.New(task.CreateTime),
		UpdateTime:   timestamppb.New(task.UpdateTime),
		CompleteTime: completeTime,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, req *browserv1beta.CreateTaskRequest) (*browserv1beta.CreateTaskResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	task := req.GetTask()
	state := task.GetState()
	if state == browserv1beta.TaskState_TASK_STATE_UNSPECIFIED {
		state = browserv1beta.TaskState_TASK_STATE_PENDING
	}

	create := s.db.Task.Create().
		SetTitle(task.GetTitle()).
		SetURL(task.GetUrl()).
		SetState(state).
		SetExtensionID(inst.GetExtensionId())

	if id := task.GetId(); id != "" {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		create = create.SetID(uid)
	}

	if ext := task.GetExtraction(); ext != nil {
		if sid := ext.GetSchemaId(); sid != "" {
			uid, err := uuid.Parse(sid)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid extraction.schema_id: "+err.Error())
			}
			create = create.SetSchemaID(uid)
		} else {
			return nil, status.Error(codes.InvalidArgument, "extraction.schema_id is required")
		}
		if rid := ext.GetRecordId(); rid != "" {
			uid, err := uuid.Parse(rid)
			if err != nil {
				return nil, status.Error(codes.InvalidArgument, "invalid extraction.record_id: "+err.Error())
			}
			create = create.SetRecordID(uid)
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "task.extraction is required")
	}

	created, err := create.Save(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.CreateTaskResponse{Task: TaskToProto(created)}, nil
}

func (s *TaskService) GetTask(ctx context.Context, req *browserv1beta.GetTaskRequest) (*browserv1beta.GetTaskResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	task, err := s.db.Task.Query().Where(
		task.ID(id),
		task.ExtensionID(inst.GetExtensionId()),
	).Only(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.GetTaskResponse{Task: TaskToProto(task)}, nil
}

func (s *TaskService) ListTasks(ctx context.Context, req *browserv1beta.ListTasksRequest) (*browserv1beta.ListTasksResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	preds := []predicate.Task{
		task.ExtensionID(inst.GetExtensionId()),
	}

	if len(req.GetStateFilter()) > 0 {
		preds = append(preds, task.StateIn(req.GetStateFilter()...))
	}

	nextPageToken, tasks, err := s.pager.GetNextPage(ctx, req, s.db.Task.Query().Where(preds...))
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.ListTasksResponse{
		Tasks:         util.SliceMap(tasks, TaskToProto),
		NextPageToken: nextPageToken,
	}, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, req *browserv1beta.UpdateTaskRequest) (*browserv1beta.UpdateTaskResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	proto := req.GetTask()
	id, err := uuid.Parse(proto.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	upd := s.db.Task.UpdateOneID(id).Where(
		task.ExtensionID(inst.GetExtensionId()),
	)

	var (
		stateUpdated    bool
		newState        browserv1beta.TaskState
		completeTimeSet bool
	)

	applyPath := func(path string) {
		switch path {
		case "title":
			upd.SetTitle(proto.GetTitle())
		case "url":
			upd.SetURL(proto.GetUrl())
		case "state":
			upd.SetState(proto.GetState())
			stateUpdated = true
			newState = proto.GetState()
		case "extraction":
			if ext := proto.GetExtraction(); ext != nil {
				if sid := ext.GetSchemaId(); sid != "" {
					uid, err := uuid.Parse(sid)
					if err == nil {
						upd.SetSchemaID(uid)
					}
				}
				if rid := ext.GetRecordId(); rid != "" {
					uid, err := uuid.Parse(rid)
					if err == nil {
						upd.SetRecordID(uid)
					}
				}
			}
		case "complete_time":
			completeTimeSet = true
			if ct := proto.GetCompleteTime(); ct != nil {
				upd.SetCompleteTime(ct.AsTime())
			} else {
				upd.ClearCompleteTime()
			}
		}
	}

	mask := req.GetUpdateMask()
	if mask == nil || len(mask.GetPaths()) == 0 {
		// Full replacement — apply all mutable fields.
		for _, p := range []string{"title", "url", "state", "extraction", "complete_time"} {
			applyPath(p)
		}
	} else {
		for _, p := range mask.GetPaths() {
			applyPath(p)
		}
	}

	// Auto-set complete_time on terminal state transitions unless it was
	// explicitly provided in the request.
	if stateUpdated && !completeTimeSet {
		if newState == browserv1beta.TaskState_TASK_STATE_COMPLETED ||
			newState == browserv1beta.TaskState_TASK_STATE_DISMISSED {
			upd.SetCompleteTime(time.Now())
		}
	}

	updated, err := upd.Save(ctx)
	if err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.UpdateTaskResponse{Task: TaskToProto(updated)}, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, req *browserv1beta.DeleteTaskRequest) (*browserv1beta.DeleteTaskResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := s.db.Task.DeleteOneID(id).Where(
		task.ExtensionID(inst.GetExtensionId()),
	).Exec(ctx); err != nil {
		return nil, db.ToConnectError(err)
	}

	return &browserv1beta.DeleteTaskResponse{}, nil
}

func (s *TaskService) StreamTaskUpdates(req *browserv1beta.StreamTaskUpdatesRequest, stream grpc.ServerStreamingServer[browserv1beta.StreamTaskUpdatesResponse]) error {
	inst, err := s.mux.GetInstance(stream.Context())
	if err != nil {
		return err
	}

	sub, ch := s.broker.Subscribe(inst.GetExtensionId())
	defer s.broker.Unsubscribe(sub)

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(evt.resp); err != nil {
				return err
			}
		case <-stream.Context().Done():
			return nil
		}
	}
}

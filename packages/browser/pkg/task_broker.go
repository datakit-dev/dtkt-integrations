package pkg

import (
	"context"
	"sync"

	"entgo.io/ent"
	"github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent/hook"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/google/uuid"

	entpkg "github.com/datakit-dev/dtkt-integrations/browser/pkg/db/ent"
)

// TaskBroker fans out task mutation events to all active StreamTaskUpdates
// subscribers. It is wired into the Ent client via RegisterHooks.
type (
	TaskBroker struct {
		ctx  context.Context
		mu   sync.RWMutex
		subs map[TaskSub]chan TaskEvent
	}
	TaskSub struct {
		id          uuid.UUID
		extensionId string
	}
	TaskEvent struct {
		resp        *browserv1beta.StreamTaskUpdatesResponse
		extensionId string
	}
)

func NewTaskBroker(ctx context.Context, db *entpkg.Client) *TaskBroker {
	b := &TaskBroker{
		ctx:  ctx,
		subs: make(map[TaskSub]chan TaskEvent),
	}

	RegisterTaskHooks(b, db)

	// When the server context is cancelled, close all subscriber channels so
	// active StreamTaskUpdates RPCs return cleanly.
	context.AfterFunc(ctx, b.closeAll)
	return b
}

// RegisterHooks attaches Ent mutation hooks to the provided client so that
// every Create, Update and Delete publishes to the broker.
func RegisterTaskHooks(broker *TaskBroker, client *entpkg.Client) {
	// ── Delete: capture ID before the row is removed, publish afterward ────────
	//
	// Two-hook approach: a pre-mutation hook stashes the task ID in a
	// context value; the post-mutation hook reads it and publishes DELETED.
	type deletedIDKey struct{}

	client.Task.Use(
		// Pre-hook: record the ID being deleted.
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return hook.TaskFunc(func(ctx context.Context, mut *entpkg.TaskMutation) (ent.Value, error) {
					if id, exists := mut.ID(); exists {
						t, err := mut.Client().Task.Get(ctx, id)
						if err == nil {
							ctx = context.WithValue(ctx, deletedIDKey{}, t)
						}
					}
					return next.Mutate(ctx, mut)
				})
			},
			entpkg.OpDeleteOne,
		),

		// Post-hook: fires after all ops; dispatches the appropriate event.
		hook.On(
			func(next ent.Mutator) ent.Mutator {
				return hook.TaskFunc(func(ctx context.Context, mut *entpkg.TaskMutation) (ent.Value, error) {
					val, err := next.Mutate(ctx, mut)
					if err != nil {
						return val, err
					}

					switch mut.Op() {
					case entpkg.OpCreate:
						if t, ok := val.(*entpkg.Task); ok {
							broker.Publish(TaskEvent{
								resp: &browserv1beta.StreamTaskUpdatesResponse{
									Task:       TaskToProto(t),
									UpdateType: browserv1beta.StreamTaskUpdatesResponse_UPDATE_TYPE_CREATED,
								},
								extensionId: t.ExtensionID,
							})
						}

					case entpkg.OpUpdateOne:
						if t, ok := val.(*entpkg.Task); ok {
							broker.Publish(TaskEvent{
								resp: &browserv1beta.StreamTaskUpdatesResponse{
									Task:       TaskToProto(t),
									UpdateType: browserv1beta.StreamTaskUpdatesResponse_UPDATE_TYPE_UPDATED,
								},
								extensionId: t.ExtensionID,
							})
						}
					case entpkg.OpDeleteOne:
						t, ok := ctx.Value(deletedIDKey{}).(*entpkg.Task)
						if ok {
							broker.Publish(TaskEvent{
								resp: &browserv1beta.StreamTaskUpdatesResponse{
									Task:       TaskToProto(t),
									UpdateType: browserv1beta.StreamTaskUpdatesResponse_UPDATE_TYPE_DELETED,
								},
								extensionId: t.ExtensionID,
							})
						}
					}

					return val, nil
				})
			},
			entpkg.OpCreate|entpkg.OpUpdateOne|entpkg.OpDeleteOne,
		),
	)
}

// Subscribe returns a channel that receives task updates. The caller must call
// Unsubscribe with the returned ID when done (e.g. via defer).
func (b *TaskBroker) Subscribe(extensionID string) (TaskSub, <-chan TaskEvent) {
	sub := TaskSub{
		id:          uuid.New(),
		extensionId: extensionID,
	}

	ch := make(chan TaskEvent, 16)
	b.mu.Lock()
	b.subs[sub] = ch
	b.mu.Unlock()
	return sub, ch
}

// Unsubscribe removes the subscriber and closes its channel.
func (b *TaskBroker) Unsubscribe(sub TaskSub) {
	b.mu.Lock()
	if ch, ok := b.subs[sub]; ok {
		delete(b.subs, sub)
		close(ch)
	}
	b.mu.Unlock()
}

// Publish sends an update to all current subscribers. Slow subscribers are
// skipped (non-blocking send) to avoid back-pressure on mutations. Publishes
// are dropped if the broker context is already done.
func (b *TaskBroker) Publish(evt TaskEvent) {
	select {
	case <-b.ctx.Done():
		return
	default:
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	for sub, ch := range b.subs {
		if sub.extensionId != evt.extensionId {
			continue
		}

		select {
		case ch <- evt:
		case <-b.ctx.Done():
			return
		default:
		}
	}
}

// closeAll closes every subscriber channel. Called when the broker context ends.
func (b *TaskBroker) closeAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for id, ch := range b.subs {
		delete(b.subs, id)
		close(ch)
	}
}

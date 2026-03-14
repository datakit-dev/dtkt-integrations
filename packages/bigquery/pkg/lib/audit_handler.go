package lib

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	bigqueryv1beta "github.com/datakit-dev/dtkt-integrations/bigquery/pkg/proto/integration/bigquery/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"google.golang.org/genproto/googleapis/cloud/audit"
	"google.golang.org/protobuf/proto"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"

	"cloud.google.com/go/pubsub"
)

var _ v1beta1.PullSourceHandler[*bigqueryv1beta.AuditLogConfig] = (*AuditLogHandler[AuditLogInstance])(nil)

const AuditLogEventSourceName = "AuditLogs"

type (
	AuditLogHandler[I AuditLogInstance] struct {
		log          *slog.Logger
		inst         I
		config       *bigqueryv1beta.AuditLogConfig
		client       *pubsub.Client
		subscription *pubsub.Subscription

		ctx    context.Context
		cancel context.CancelFunc
		msgCh  chan *pubsub.Message
	}
	AuditLogInstance interface {
		v1beta1.InstanceType
		Client() *Client
		GetAuditLogEvent(context.Context, v1beta1.RegisteredEvent, *audit.AuditLog) (any, error)
	}
)

func AuditLogSource[I AuditLogInstance]() v1beta1.RegisterSourceFunc[I] {
	return v1beta1.NewPullSource(
		AuditLogEventSourceName,
		"BigQuery Audit Logs event source provides a stream of log events performed in BigQuery.",
		false, true,
		time.Second,
		NewAuditLogHandler[I],
	)
}

func NewAuditLogHandler[I AuditLogInstance](ctx context.Context, mux v1beta1.InstanceMux[I], config *bigqueryv1beta.AuditLogConfig) (*AuditLogHandler[I], error) {
	inst, err := mux.GetInstance(ctx)
	if err != nil {
		return nil, err
	}

	if config == nil && inst.Client().Config.AuditLog == nil {
		return nil, fmt.Errorf("audit log config required")
	} else if config == nil {
		config = inst.Client().Config.AuditLog
	}

	client, err := pubsub.NewClient(ctx, inst.Client().Config.ProjectId, inst.Client().Options...)
	if err != nil {
		return nil, err
	}

	logger := log.FromCtx(ctx).WithGroup("AuditLogHandler")
	ctx, cancel := context.WithCancel(ctx)

	handler := &AuditLogHandler[I]{
		log:          logger,
		inst:         inst,
		config:       config,
		client:       client,
		subscription: client.Subscription(config.Subscription),

		ctx:    ctx,
		cancel: cancel,
		msgCh:  make(chan *pubsub.Message, 100),
	}

	go handler.Receive()

	return handler, nil
}

func (h *AuditLogHandler[I]) Receive() error {
	return h.subscription.Receive(h.ctx, func(ctx context.Context, msg *pubsub.Message) {
		select {
		case <-ctx.Done():
		case h.msgCh <- msg:
			msg.Ack()
		}
	})
}

func (h *AuditLogHandler[I]) PullEvent(ctx context.Context, events *v1beta1.EventRegistry) (*v1beta1.EventWithPayload, error) {
	select {
	case <-ctx.Done():
		h.cancel()
		return nil, ctx.Err()
	case msg := <-h.msgCh:
		return ProcessAuditLogEvent(ctx, h.inst, events, msg)
	}
}

func (h *AuditLogHandler[I]) ConfigEqual(config *bigqueryv1beta.AuditLogConfig) bool {
	return proto.Equal(h.config, config)
}

func (h *AuditLogHandler[I]) Close() error {
	h.cancel()
	return h.client.Close()
}

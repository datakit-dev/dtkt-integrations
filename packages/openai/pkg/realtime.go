package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/coder/websocket"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
)

const (
	RealtimeEventSourceName = "RealtimeEvents"
	RealtimeEndpoint        = "wss://api.openai.com/v1/realtime"
	RealtimeDefaultModel    = "gpt-realtime"
)

var realtimeEndpoint, _ = url.Parse(RealtimeEndpoint)

type (
	RealtimeHandler struct {
		config *RealtimeConfig
		conn   *websocket.Conn
	}
	RealtimeConfig struct {
		ID    string `json:"id"`
		Model string `json:"model,omitempty"`

		apiKey string
	}
	RealtimeClientEventInput[T any] struct {
		Config *RealtimeConfig `json:"config"`
		Event  T               `json:"event"`
	}
	RealtimeClientEventOutput struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
)

func RealtimeSource() v1beta1.RegisterSourceFunc[*Instance] {
	// v1beta1.NewPushSource(
	// 	"RealtimeEvents",
	// 	"OpenAI's Realtime API enables you to build low-latency, multi-modal conversational experiences. It currently supports text and audio as both input and output, as well as function calling.",
	// 	true, true,
	// 	func(ctx context.Context, config *RealtimeConfig) (*RealtimeHandler, error) {
	// 		handler, err := NewRealtimeHandler(ctx, inst, config)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		inst.handlers.Store(config.ID, handler)

	// 		return handler, nil
	// 	},
	// ),
	return v1beta1.NewPullSource(
		RealtimeEventSourceName,
		"OpenAI's Realtime API enables you to build low-latency, multi-modal conversational experiences. It currently supports text and audio as both input and output, as well as function calling.",
		false, true, 0,
		func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], config *RealtimeConfig) (*RealtimeHandler, error) {
			inst, err := mux.GetInstance(ctx)
			if err != nil {
				return nil, err
			}

			config.apiKey = inst.config.APIKey

			handler, err := NewRealtimeHandler(ctx, config)
			if err != nil {
				return nil, err
			}

			inst.handlers.Store(config.ID, handler)

			return handler, nil
		},
	)
}

func NewRealtimeAction[T any](name string) v1beta1.RegisterActionFunc[*Instance] {
	return v1beta1.NewAction(name, EventDescriptions[name],
		func(ctx context.Context, mux v1beta1.InstanceMux[*Instance], input *RealtimeClientEventInput[T]) (*RealtimeClientEventOutput, error) {
			inst, err := mux.GetInstance(ctx)
			if err != nil {
				return nil, err
			}

			handler, ok := inst.handlers.Load(input.Config.ID)
			if ok {
				event, err := common.MarshalJSON[[]byte](input.Event)
				if err != nil {
					return &RealtimeClientEventOutput{
						Success: false,
						Error:   err.Error(),
					}, nil
				}

				if err = handler.conn.Write(ctx, websocket.MessageText, event); err != nil {
					return &RealtimeClientEventOutput{
						Success: true,
						Error:   err.Error(),
					}, nil
				}

				return &RealtimeClientEventOutput{
					Success: true,
				}, nil
			}

			return nil, fmt.Errorf("realtime handler not found")
		},
	)
}

func NewRealtimeHandler(ctx context.Context, config *RealtimeConfig) (*RealtimeHandler, error) {
	if config.Model == "" {
		config.Model = RealtimeDefaultModel
	}

	url := *realtimeEndpoint
	q := url.Query()
	q.Set("model", config.Model)
	url.RawQuery = q.Encode()

	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+config.apiKey)

	conn, _, err := websocket.Dial(ctx, url.String(), &websocket.DialOptions{
		HTTPHeader: headers,
	})
	if err != nil {
		return nil, err
	}

	conn.SetReadLimit(-1)

	return &RealtimeHandler{
		config: config,
		conn:   conn,
	}, nil
}

func (h *RealtimeHandler) ConfigEqual(c *RealtimeConfig) bool {
	return c != nil && h.config != nil && *c == *h.config
}

func (h *RealtimeHandler) PullEvent(ctx context.Context, events *v1beta1.EventRegistry) (*v1beta1.EventWithPayload, error) {
	_, eventBytes, err := h.conn.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("error reading websocket: %w", err)
	}

	log.Info(ctx, "PullEvent", slog.String("event", string(eventBytes)))

	eventMap, err := common.UnmarshalJSON[common.JSONMap](eventBytes)
	if err != nil {
		return nil, fmt.Errorf("error unmashalling event to map: %w", err)
	}

	eventType, ok := common.GetJSONValue[string](eventMap, ".type")
	if !ok {
		return nil, fmt.Errorf("error unmashalling event type: %s", eventType)
	}

	event, err := events.Find(eventType)
	if err != nil {
		return nil, fmt.Errorf("invalid event type: %s", eventType)
	}

	return event.WithPayload(map[string]any(eventMap))
}

func (h *RealtimeHandler) Close() error {
	if h.conn != nil {
		return h.conn.Close(websocket.StatusNormalClosure, "")
	}
	return nil
}

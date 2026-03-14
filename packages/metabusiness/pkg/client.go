package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	fb "github.com/huandu/facebook/v2"
	"github.com/invopop/jsonschema"
)

type (
	Client struct {
		log    *slog.Logger
		app    *fb.App
		config *Config
	}
	Input  map[string]any
	Output map[string]any
)

func NewClient(log *slog.Logger, app *fb.App, config *Config) *Client {
	return &Client{
		app:    app,
		config: config,
	}
}

func (c *Client) performRequest(ctx context.Context, method string, input Input) (fb.Result, error) {
	session, err := NewSession(ctx, c.app, NewAuthConfig(c.config, DefaultScopes))
	if err != nil {
		return nil, err
	}

	if input["access_token"] != nil {
		if accessToken, ok := input["access_token"].(string); ok {
			session.SetAccessToken(accessToken)
		}
	}

	path, ok := input["path"].(string)
	if ok {
		return nil, fmt.Errorf("input.path not specified")
	}

	var params map[string]any
	if input["params"] != nil {
		if params, ok = input["params"].(map[string]any); !ok {
			return nil, fmt.Errorf("input.params must be object")
		}
	}

	switch method {
	case http.MethodGet:
		return session.Get(path, params)
	case http.MethodPost:
		return session.Post(path, params)
	case http.MethodPut:
		return session.Put(path, params)
	case http.MethodDelete:
		return session.Delete(path, params)
	}

	return nil, fmt.Errorf("unsupported method: %s", method)
}

func (c *Client) prepareResponse(resp fb.Result) (Output, error) {
	return Output(resp), nil
}

func (c *Client) Get(ctx context.Context, input Input) (Output, error) {
	resp, err := c.performRequest(ctx, http.MethodGet, input)
	if err != nil {
		c.log.Error(http.MethodGet, slog.Any("DebugInfo", resp.DebugInfo()))
		return nil, err
	}

	return c.prepareResponse(resp)
}

func (c *Client) Post(ctx context.Context, input Input) (Output, error) {
	resp, err := c.performRequest(ctx, http.MethodPost, input)
	if err != nil {
		c.log.Error(http.MethodPost, slog.Any("DebugInfo", resp.DebugInfo()))
		return nil, err
	}

	return c.prepareResponse(resp)
}

func (c *Client) Put(ctx context.Context, input Input) (Output, error) {
	resp, err := c.performRequest(ctx, http.MethodPut, input)
	if err != nil {
		c.log.Error(http.MethodPut, slog.Any("DebugInfo", resp.DebugInfo()))
		return nil, err
	}

	return c.prepareResponse(resp)
}

func (c *Client) Delete(ctx context.Context, input Input) (Output, error) {
	resp, err := c.performRequest(ctx, http.MethodDelete, input)
	if err != nil {
		c.log.Error(http.MethodDelete, slog.Any("DebugInfo", resp.DebugInfo()))
		return nil, err
	}

	return c.prepareResponse(resp)
}

func (Input) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

func (Output) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: jsonschema.TrueSchema,
	}
}

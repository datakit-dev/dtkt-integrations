package pkg

import (
	"context"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/env"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/middleware"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/network"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg/chrome"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
)

// Integration instance struct
type Instance struct {
	config *browserv1beta.Config
}

// Creates a new instance
func NewInstance(ctx context.Context, config *browserv1beta.Config) (*Instance, error) {
	if config.GetChrome() != nil {
		chromeConfig := config.GetChrome()
		if chromeConfig.Context == nil {
			req, err := middleware.RequestFromContext(ctx)
			if err != nil {
				return nil, err
			}

			addr, err := network.ParseAddr(env.GetVar(env.ContextAddress))
			if err != nil {
				return nil, err
			}

			chromeConfig.Context = &browserv1beta.Context{
				Address:    addr.HTTP().JoinPath("/proxy").String(),
				Connection: req.AddrName(),
			}
		}

		configDir, err := GetConfigDir()
		if err != nil {
			return nil, err
		}

		err = WriteConfig(configDir, config)
		if err != nil {
			return nil, err
		}

		err = chrome.InstallBridge(log.FromCtx(ctx), chromeConfig.GetExtensionId())
		if err != nil {
			return nil, err
		}
	}

	return &Instance{
		config: config,
	}, nil
}

func (i *Instance) GetExtensionId() string {
	return i.config.GetChrome().GetExtensionId()
}

// Orchestrates OAuth checks/exchanges.
func (i *Instance) CheckAuth(ctx context.Context, req *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

// Close is called to release instance resources (e.g. client connections)
func (i *Instance) Close() error {
	return nil
}

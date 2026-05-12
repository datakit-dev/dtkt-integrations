package pkg

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/datakit-dev/dtkt-integrations/browser/pkg/chrome"
	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/network"
	"google.golang.org/grpc"
)

func HasExtensionArg(args ...string) (_ string, _ bool) {
	if len(args) < 2 {
		return
	}

	switch {
	case strings.HasPrefix(args[1], "chrome-extension://"):
		return args[1], true
	}

	return
}

func InstallExtension(log *slog.Logger) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	return chrome.InstallExtension(log, configDir)
}

func StartExtension(extArg string) error {
	switch {
	case strings.HasPrefix(extArg, "chrome-extension://"):
		configDir, err := GetConfigDir()
		if err != nil {
			return err
		}

		config, err := ReadConfig(configDir, chrome.GetExtensionIdFromArg(extArg))
		if err != nil {
			return err
		}

		log := log.NewLogger(
			log.WithSlogDefault(false),
			log.WithFile(filepath.Join(configDir, chrome.LogFile)),
		).With(
			slog.String("address", config.GetChrome().GetBridgeAddress()),
			slog.String("configDir", configDir),
		)

		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()

		return StartServer(ctx, log, network.Addr(network.TCP, config.GetChrome().GetBridgeAddress()), func(ctx context.Context, reg grpc.ServiceRegistrar) {
			browserv1beta.RegisterChromeServiceServer(reg, chrome.NewChromeService(ctx, log, config.GetChrome()))
		})
	}

	return nil
}

func GetExtensionClient(mux v1beta1.InstanceMux[*Instance]) integrationsdk.GetProxyConnFunc {
	return func(ctx context.Context, opts ...grpc.DialOption) (context.Context, *grpc.ClientConn, error) {
		inst, err := mux.GetInstance(ctx)
		if err != nil {
			return nil, nil, err
		}

		switch {
		case inst.config.GetChrome() != nil:
			addr := network.Addr(network.TCP, inst.config.GetChrome().GetBridgeAddress())
			client, err := network.DialGRPC(addr, opts...)
			if err != nil {
				return nil, nil, err
			}

			return ctx, client, nil
		}

		return nil, nil, fmt.Errorf("invalid browser config")
	}
}

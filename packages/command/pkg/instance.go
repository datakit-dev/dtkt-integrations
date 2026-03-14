package pkg

import (
	"context"
	"fmt"

	"github.com/datakit-dev/dtkt-integrations/command/pkg/executor"
	commandintgr "github.com/datakit-dev/dtkt-integrations/command/pkg/proto/integration/command/v1beta"
	basev1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/base/v1beta1"
	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type Instance struct {
	config *commandintgr.Config

	exec executor.CommandExecutor

	commands map[string]*commandv1beta1.Command
}

func NewInstance(ctx context.Context, config *commandintgr.Config) (*Instance, error) {
	commands := make(map[string]*commandv1beta1.Command)
	for _, cmd := range config.Commands {
		commands[cmd.Name] = cmd
	}

	var err error
	var exec executor.CommandExecutor
	if config.SshConfig != nil {
		exec, err = executor.NewRemoteExecutor(config.SshConfig)
	} else {
		exec = executor.NewLocalExecutor()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create command executor: %v", err)
	}

	return &Instance{
		config:   config,
		exec:     exec,
		commands: commands,
	}, nil
}

func (i *Instance) CheckAuth(context.Context, *basev1beta1.CheckAuthRequest) (*basev1beta1.CheckAuthResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAuth not implemented")
}

func (i *Instance) Close() error {
	return i.exec.Close()
}

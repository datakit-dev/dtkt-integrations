package pkg

// import (
// 	"context"
// 	"testing"

// 	commandintgr "github.com/datakit-dev/dtkt-integrations/command/pkg/proto/integration/command/v1beta"
// 	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
// )

// func TestLocalService(t *testing.T) {
// 	RunServiceTestSuite(t, ServiceTestConfig{
// 		Name: "Local",
// 		NewInstance: func(t *testing.T) *Instance {
// 			config := &commandintgr.Config{
// 				AllowShell:   true,
// 				ShellCommand: "sh",
// 				Commands: []*commandv1beta1.Command{
// 					{Name: "echo"},
// 					{Name: "cat"},
// 					{Name: "sleep"},
// 					{Name: "test"},     // For exit code tests
// 					{Name: "pwd"},      // For workdir tests
// 					{Name: "printenv"}, // For env tests
// 				},
// 			}

// 			instance, err := NewInstance(context.Background(), config)
// 			if err != nil {
// 				t.Fatalf("failed to create instance: %v", err)
// 			}
// 			return instance
// 		},
// 		SupportsExecWorkdir: true,
// 		SupportsExecEnv:     true,
// 	})
// }

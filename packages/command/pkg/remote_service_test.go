package pkg

// import (
// 	"context"
// 	"os"
// 	"testing"

// 	"github.com/datakit-dev/dtkt-integrations/command/pkg/test"

// 	commandintgr "github.com/datakit-dev/dtkt-integrations/command/pkg/proto/integration/command/v1beta"
// 	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
// )

// // TestMain handles setup and teardown for all tests in this package
// func TestMain(m *testing.M) {
// 	code := m.Run()

// 	// Cleanup: terminate the shared SSH container if it was started
// 	test.TerminateSSHContainer()

// 	os.Exit(code)
// }

// func TestRemoteService(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping integration test in short mode")
// 	}

// 	container := test.SetupSSHContainer(t)

// 	RunServiceTestSuite(t, ServiceTestConfig{
// 		Name: "Remote",
// 		NewInstance: func(t *testing.T) *Instance {
// 			config := &commandintgr.Config{
// 				SshConfig:    container.SSHConfig(),
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
// 		SupportsExecWorkdir: false, // SSH sessions don't support workdir in non-shell mode
// 		SupportsExecEnv:     false, // SSH sessions don't support env in non-shell mode
// 	})
// }

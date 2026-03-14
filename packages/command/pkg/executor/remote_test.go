package executor

import (
	"testing"
)

// TestRemoteExecutor runs the shared executor test suite against RemoteExecutor
func TestRemoteExecutor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	RunExecutorTestSuite(t, ExecutorTestConfig{
		Name: "RemoteExecutor",
		NewExecutor: func(t *testing.T) CommandExecutor {
			exec, err := NewRemoteExecutor(container.SSHConfig())
			if err != nil {
				t.Fatalf("NewRemoteExecutor failed: %v", err)
			}
			return exec
		},
		SupportsExecWorkdir: false,
		SupportsExecEnv:     false,
	})
}

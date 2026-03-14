package executor

import (
	"testing"
)

// TestLocalExecutor runs the shared executor test suite against LocalExecutor
func TestLocalExecutor(t *testing.T) {
	RunExecutorTestSuite(t, ExecutorTestConfig{
		Name: "LocalExecutor",
		NewExecutor: func(t *testing.T) CommandExecutor {
			return NewLocalExecutor()
		},
		SupportsExecWorkdir: true,
		SupportsExecEnv:     true,
	})
}

package pkg

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"

// 	commandintgr "github.com/datakit-dev/dtkt-integrations/command/pkg/proto/integration/command/v1beta"
// 	"github.com/datakit-dev/dtkt-integrations/command/pkg/test"
// 	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// 	"google.golang.org/protobuf/types/known/durationpb"
// )

// // ServiceTestConfig holds configuration for the shared service test suite
// type ServiceTestConfig struct {
// 	// Name is the test configuration name for test output
// 	Name string

// 	// NewInstance creates a new Instance for testing
// 	NewInstance func(t *testing.T) *Instance

// 	// Skip can be used to skip certain tests
// 	Skip func(t *testing.T)

// 	// SupportsExecWorkdir indicates whether the executor supports workdir in non-shell command execution.
// 	// Local executors support this, but remote (SSH) executors do not.
// 	SupportsExecWorkdir bool

// 	// SupportsExecEnv indicates whether the executor supports env in non-shell command execution.
// 	// Local executors support this, but remote (SSH) executors do not.
// 	SupportsExecEnv bool
// }

// // RunServiceTestSuite runs the shared test suite against any Instance configuration
// func RunServiceTestSuite(t *testing.T, cfg ServiceTestConfig) {
// 	if cfg.Skip != nil {
// 		cfg.Skip(t)
// 	}

// 	t.Run("ExecuteCommand", func(t *testing.T) {
// 		runExecuteCommandTests(t, cfg)
// 	})

// 	t.Run("ExecuteShellCommand", func(t *testing.T) {
// 		runExecuteShellCommandTests(t, cfg)
// 	})

// 	t.Run("ExecuteStreamedCommand", func(t *testing.T) {
// 		runExecuteStreamedCommandTests(t, cfg)
// 	})

// 	t.Run("ExecuteStreamedShellCommand", func(t *testing.T) {
// 		runExecuteStreamedShellCommandTests(t, cfg)
// 	})

// 	t.Run("ExecuteCommands", func(t *testing.T) {
// 		runExecuteCommandsTests(t, cfg)
// 	})

// 	t.Run("ExecuteBatchCommands", func(t *testing.T) {
// 		runExecuteBatchCommandsTests(t, cfg)
// 	})

// 	t.Run("ExecuteShellCommands", func(t *testing.T) {
// 		runExecuteShellCommandsTests(t, cfg)
// 	})

// 	t.Run("ExecuteBatchShellCommands", func(t *testing.T) {
// 		runExecuteBatchShellCommandsTests(t, cfg)
// 	})

// 	t.Run("TerminalSession", func(t *testing.T) {
// 		runTerminalSessionTests(t, cfg)
// 	})
// }

// func runExecuteCommandTests(t *testing.T, cfg ServiceTestConfig) {
// 	ctx := context.Background()

// 	t.Run("simple echo command", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"hello", "world"},
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Equal(t, "hello world\n", string(resp.Output.Stdout))
// 		assert.Empty(t, resp.Output.Stderr)
// 	})

// 	t.Run("command with stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "cat",
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("test input"),
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Equal(t, "test input", string(resp.Output.Stdout))
// 	})

// 	t.Run("command not found", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "nonexistent",
// 			},
// 		}

// 		_, err := instance.ExecuteCommand(ctx, req)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "not found")
// 	})

// 	t.Run("non zero exit", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "test",
// 				Args:    []string{"1", "-eq", "0"},
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.NotEqual(t, int32(0), resp.Result.ExitCode)
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		if !cfg.SupportsExecEnv {
// 			t.Skip("executor does not support env in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "printenv",
// 				Args:    []string{"MY_TEST_VAR"},
// 				Env:     map[string]string{"MY_TEST_VAR": "test_value_123"},
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Contains(t, string(resp.Output.Stdout), "test_value_123")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		if !cfg.SupportsExecWorkdir {
// 			t.Skip("executor does not support workdir in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		// macOS resolves /tmp to /private/tmp
// 		output := string(resp.Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("with_timeout", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"10"},
// 				Timeout: durationpb.New(100 * time.Millisecond),
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		// Command was killed by timeout - returns result, not error
// 		require.NoError(t, err)
// 		require.NotNil(t, resp.Result)
// 		// Exit code is -1 when killed by timeout
// 		assert.Equal(t, int32(-1), resp.Result.ExitCode)
// 		assert.Equal(t, commandv1beta1.CommandResult_REASON_TIMED_OUT, resp.Result.Reason)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		// cat with nonexistent file writes error to stderr and exits non-zero
// 		req := &commandv1beta1.ExecuteCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "cat",
// 				Args:    []string{"/nonexistent_file_for_stderr_test"},
// 			},
// 		}

// 		resp, err := instance.ExecuteCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.NotEqual(t, int32(0), resp.Result.ExitCode)
// 		assert.Contains(t, string(resp.Output.Stderr), "nonexistent_file_for_stderr_test")
// 	})
// }

// func runExecuteShellCommandTests(t *testing.T, cfg ServiceTestConfig) {
// 	ctx := context.Background()

// 	t.Run("simple shell command", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo hello",
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Equal(t, "hello\n", string(resp.Output.Stdout))
// 	})

// 	t.Run("shell command with pipes", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo -e 'line1\\nline2\\nline3' | wc -l",
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Contains(t, string(resp.Output.Stdout), "3")
// 	})

// 	t.Run("shell disabled", func(t *testing.T) {
// 		// Create instance with shell disabled
// 		config := &commandintgr.Config{
// 			AllowShell: false,
// 			Commands:   []*commandv1beta1.Command{},
// 		}
// 		instance, err := NewInstance(ctx, config)
// 		require.NoError(t, err)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo hello",
// 			},
// 		}

// 		_, err = instance.ExecuteShellCommand(ctx, req)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "not allowed")
// 	})

// 	t.Run("non zero exit", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "test 1 -eq 0",
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.NotEqual(t, int32(0), resp.Result.ExitCode)
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo $MY_SHELL_VAR",
// 				Env:     map[string]string{"MY_SHELL_VAR": "shell_test_value"},
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Contains(t, string(resp.Output.Stdout), "shell_test_value")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		output := string(resp.Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("with_timeout", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "sleep 10",
// 				Timeout: durationpb.New(100 * time.Millisecond),
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		// Command was killed by timeout - returns result, not error
// 		require.NoError(t, err)
// 		require.NotNil(t, resp.Result)
// 		// Exit code is -1 when killed by timeout
// 		assert.Equal(t, int32(-1), resp.Result.ExitCode)
// 		assert.Equal(t, commandv1beta1.CommandResult_REASON_TIMED_OUT, resp.Result.Reason)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		req := &commandv1beta1.ExecuteShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo stdout_shell && echo stderr_shell >&2",
// 			},
// 		}

// 		resp, err := instance.ExecuteShellCommand(ctx, req)
// 		require.NoError(t, err)
// 		assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		assert.Contains(t, string(resp.Output.Stdout), "stdout_shell")
// 		assert.Contains(t, string(resp.Output.Stderr), "stderr_shell")
// 	})
// }

// func runExecuteStreamedCommandTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("simple streamed echo", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"hello", "streamed"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "hello streamed\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("streamed command with stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("chunk1 "),
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("chunk2"),
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "chunk1 chunk2", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("command not found", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "nonexistent",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "not found")
// 	})

// 	t.Run("no command in first request", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("data"),
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid")
// 	})

// 	t.Run("context cancellation", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx, cancel := context.WithCancel(context.Background())
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"10"},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)
// 		cancel()

// 		select {
// 		case <-done:
// 		case <-time.After(2 * time.Second):
// 			t.Fatal("service did not respond to context cancellation")
// 		}
// 	})

// 	t.Run("client_immediate_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		// Immediately disconnect after sending command
// 		client.Close()

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly on immediate client disconnect")
// 		}
// 	})

// 	t.Run("client_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"30"},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		// Wait for command to start
// 		time.Sleep(100 * time.Millisecond)

// 		// Close stream (simulating client connection drop)
// 		client.Close()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly when client disconnected during busy execution")
// 		}
// 	})

// 	// Input handling tests
// 	runStreamedCommandInputTests(t, cfg)

// 	// Concurrency tests
// 	t.Run("concurrent commands", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		const numConcurrent = 10

// 		var wg sync.WaitGroup
// 		errors := make(chan error, numConcurrent)

// 		for i := 0; i < numConcurrent; i++ {
// 			wg.Add(1)
// 			go func(n int) {
// 				defer wg.Done()

// 				ctx := context.Background()
// 				client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 				client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 					Command: &commandv1beta1.ExecutableCommand{
// 						Command: "echo",
// 						Args:    []string{"concurrent", "test"},
// 					},
// 				})
// 				client.CloseSend()

// 				if err := instance.ExecuteStreamedCommand(server); err != nil {
// 					errors <- err
// 					return
// 				}

// 				responses := client.DrainResponses()
// 				var gotResult bool
// 				for _, resp := range responses {
// 					if resp.GetResult() != nil {
// 						gotResult = true
// 					}
// 				}
// 				if !gotResult {
// 					errors <- fmt.Errorf("no result in concurrent test %d", n)
// 				}
// 			}(i)
// 		}

// 		wg.Wait()
// 		close(errors)

// 		for err := range errors {
// 			t.Errorf("concurrent test error: %v", err)
// 		}
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		if !cfg.SupportsExecEnv {
// 			t.Skip("executor does not support env in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "printenv",
// 				Args:    []string{"STREAM_TEST_VAR"},
// 				Env:     map[string]string{"STREAM_TEST_VAR": "streamed_value"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Contains(t, string(collected.Stdout), "streamed_value")
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		if !cfg.SupportsExecWorkdir {
// 			t.Skip("executor does not support workdir in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		output := string(collected.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		// cat with nonexistent file writes error to stderr
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "cat",
// 				Args:    []string{"/nonexistent_file_for_stderr_test"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.NotEqual(t, int32(0), collected.ExitCode)
// 		assert.Contains(t, string(collected.Stderr), "nonexistent_file_for_stderr_test")
// 	})

// 	t.Run("non zero exit", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "test",
// 				Args:    []string{"1", "-eq", "0"}, // fails with non-zero exit
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		assert.NotEqual(t, int32(0), collected.ExitCode)
// 	})

// 	t.Run("with_timeout_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		// Start a command that expects stdin with a short timeout
// 		// The command will wait for stdin indefinitely, so the timeout should trigger
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 				Timeout:      durationpb.New(100 * time.Millisecond),
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		// Don't send any stdin data - let the timeout trigger
// 		// Eventually close the client to allow cleanup
// 		select {
// 		case <-done:
// 			// Service completed (due to timeout)
// 		case <-time.After(5 * time.Second):
// 			client.Close()
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		// Exit code is -1 when killed by timeout
// 		assert.Equal(t, int32(-1), collected.ExitCode)
// 		assert.Equal(t, commandv1beta1.CommandResult_REASON_TIMED_OUT, collected.Reason)
// 	})
// }

// func runExecuteStreamedShellCommandTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("simple streamed shell", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo 'hello shell'",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, "hello shell\n", string(collected.Stdout))
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("shell with interactive stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		for i := 0; i < 5; i++ {
// 			client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 				Input: &commandv1beta1.CommandInput{
// 					Stdin: []byte("line\n"),
// 				},
// 			})
// 		}

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "line\nline\nline\nline\nline\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("large output buffering", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "for i in $(seq 1 1000); do echo 'This is line number '$i' with some padding'; done",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		var totalBytes int
// 		for _, resp := range responses {
// 			if resp.GetOutput() != nil {
// 				totalBytes += len(resp.GetOutput().Stdout)
// 			}
// 		}

// 		t.Logf("Received %d bytes", totalBytes)
// 		assert.Greater(t, totalBytes, 10000, "should have received significant output")
// 	})

// 	// Input handling tests
// 	runStreamedShellInputTests(t, cfg)

// 	// Signal tests
// 	t.Run("send SIGTERM to running command", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "sleep 30",
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Signal: commandv1beta1.Signal_SIGNAL_SIGTERM,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("command did not terminate after SIGTERM")
// 		}
// 	})

// 	t.Run("client_immediate_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		// Immediately disconnect after sending command
// 		client.Close()

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly on immediate client disconnect")
// 		}
// 	})

// 	t.Run("client_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "sleep 30",
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		// Wait for command to start
// 		time.Sleep(100 * time.Millisecond)

// 		// Close stream (simulating client connection drop)
// 		client.Close()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly when client disconnected during busy execution")
// 		}
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo $SHELL_STREAM_VAR",
// 				Env:     map[string]string{"SHELL_STREAM_VAR": "shell_stream_value"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Contains(t, string(collected.Stdout), "shell_stream_value")
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		output := string(collected.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo stdout_shell_stream && echo stderr_shell_stream >&2",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Contains(t, string(collected.Stdout), "stdout_shell_stream")
// 		assert.Contains(t, string(collected.Stderr), "stderr_shell_stream")
// 	})

// 	t.Run("with_timeout_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		// Start a shell command that expects stdin with a short timeout
// 		// The command will wait for stdin indefinitely, so the timeout should trigger
// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 				Timeout:      durationpb.New(100 * time.Millisecond),
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		// Don't send any stdin data - let the timeout trigger
// 		// Eventually close the client to allow cleanup
// 		select {
// 		case <-done:
// 			// Service completed (due to timeout)
// 		case <-time.After(5 * time.Second):
// 			client.Close()
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		// Exit code is -1 when killed by timeout
// 		assert.Equal(t, int32(-1), collected.ExitCode)
// 		assert.Equal(t, commandv1beta1.CommandResult_REASON_TIMED_OUT, collected.Reason)
// 	})
// }

// func runStreamedCommandInputTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("command_stdin_eof_in_first_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("all in one message"),
// 				Eof:   true,
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "all in one message", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_eof_in_subsequent_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("final data"),
// 				Eof:   true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "final data", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_after_eof_returns_error", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("before eof"),
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte(" should cause error"),
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.Error(t, serviceErr)
// 		assert.Contains(t, serviceErr.Error(), "stdin data received after EOF")
// 	})

// 	t.Run("duplicate_command_ignored", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"first"},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"second"},
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "first\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_ignored_without_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "echo",
// 				Args:         []string{"hello"},
// 				ExpectsStdin: false,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("this should be ignored"),
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "hello\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_in_first_message_ignored_without_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "echo",
// 				Args:         []string{"output"},
// 				ExpectsStdin: false,
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("ignored input"),
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "output\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("rapid_stdin_chunks", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		for i := 0; i < 50; i++ {
// 			client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 				Input: &commandv1beta1.CommandInput{
// 					Stdin: []byte(fmt.Sprintf("%d,", i)),
// 				},
// 			})
// 		}

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)

// 		expected := ""
// 		for i := 0; i < 50; i++ {
// 			expected += fmt.Sprintf("%d,", i)
// 		}
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, expected, string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("empty_stdin_chunks", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("start"),
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte{},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: nil,
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("end"),
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "startend", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("signal_in_first_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedCommandRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"30"},
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Signal: commandv1beta1.Signal_SIGNAL_SIGTERM,
// 			},
// 		})
// 		client.CloseSend()

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedCommand(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("command did not terminate after immediate SIGTERM")
// 		}
// 	})
// }

// // runStreamedShellInputTests tests input handling for ExecuteStreamedShellCommand
// func runStreamedShellInputTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("command_stdin_eof_in_first_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("shell all in one"),
// 				Eof:   true,
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "shell all in one", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_eof_in_subsequent_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("shell final data"),
// 				Eof:   true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "shell final data", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_after_eof_returns_error", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("before eof"),
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte(" should cause error"),
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.Error(t, serviceErr)
// 		assert.Contains(t, serviceErr.Error(), "stdin data received after EOF")
// 	})

// 	t.Run("duplicate_command_ignored", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo first",
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo second",
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "first\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_ignored_without_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "echo hello",
// 				ExpectsStdin: false,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("this should be ignored"),
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "hello\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("stdin_in_first_message_ignored_without_expects_stdin", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "echo output",
// 				ExpectsStdin: false,
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("ignored input"),
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "output\n", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("rapid_stdin_chunks", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		for i := 0; i < 50; i++ {
// 			client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 				Input: &commandv1beta1.CommandInput{
// 					Stdin: []byte(fmt.Sprintf("%d,", i)),
// 				},
// 			})
// 		}

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)

// 		expected := ""
// 		for i := 0; i < 50; i++ {
// 			expected += fmt.Sprintf("%d,", i)
// 		}
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, expected, string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("empty_stdin_chunks", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command:      "cat",
// 				ExpectsStdin: true,
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("start"),
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte{},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: nil,
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("end"),
// 			},
// 		})

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Eof: true,
// 			},
// 		})
// 		client.CloseSend()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)
// 		collected := test.CollectCommandResponses(responses)
// 		assert.Equal(t, int32(0), collected.ExitCode)
// 		assert.Equal(t, "startend", string(collected.Stdout))
// 		assert.Empty(t, collected.Stderr)
// 	})

// 	t.Run("non zero exit", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "test 1 -eq 0", // fails with non-zero exit
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteStreamedShellCommand(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		collected := test.CollectCommandResponses(responses)
// 		assert.NotEqual(t, int32(0), collected.ExitCode)
// 	})

// 	t.Run("signal_in_first_message", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteStreamedShellCommandRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "sleep 30",
// 			},
// 			Input: &commandv1beta1.CommandInput{
// 				Signal: commandv1beta1.Signal_SIGNAL_SIGTERM,
// 			},
// 		})
// 		client.CloseSend()

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteStreamedShellCommand(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("command did not terminate after immediate SIGTERM")
// 		}
// 	})
// }

// func runExecuteCommandsTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("stream_multiple_commands", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// Send multiple commands
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"first"},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"second"},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"third"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 3)

// 		// Each response should have result and output
// 		for i, resp := range responses {
// 			require.NotNil(t, resp.Result, "response %d should have result", i)
// 			assert.Equal(t, int32(0), resp.Result.ExitCode)
// 			require.NotNil(t, resp.Output, "response %d should have output", i)
// 		}
// 	})

// 	t.Run("results_streamed_in_order", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// Send multiple commands - results should come back in order (sequential execution)
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"first"},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"second"},
// 			},
// 		})
// 		client.CloseSend()

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.ExecuteCommands(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for service to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 2)
// 	})

// 	t.Run("partial failure", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// Mix of successful and failing commands - streaming executes sequentially
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"first"},
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "test",
// 				Args:    []string{"1", "-eq", "0"}, // fails with non-zero exit
// 			},
// 		})
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "echo",
// 				Args:    []string{"third"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err, "stream should succeed even with failing commands")

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 3, "all commands should return results")

// 		// Verify results in order (streaming is sequential)
// 		assert.Equal(t, int32(0), responses[0].Result.ExitCode)
// 		assert.Contains(t, string(responses[0].Output.Stdout), "first")

// 		assert.NotEqual(t, int32(0), responses[1].Result.ExitCode)

// 		assert.Equal(t, int32(0), responses[2].Result.ExitCode)
// 		assert.Contains(t, string(responses[2].Output.Stdout), "third")
// 	})

// 	t.Run("client_immediate_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		assert.Empty(t, responses)
// 	})

// 	t.Run("client_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// Send a long-running command
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"30"},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteCommands(server)
// 			close(done)
// 		}()

// 		// Wait for command to start
// 		time.Sleep(100 * time.Millisecond)

// 		// Close stream (simulating client connection drop)
// 		client.Close()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly when client disconnected during busy execution")
// 		}
// 	})

// 	t.Run("context_cancellation", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx, cancel := context.WithCancel(context.Background())
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// Send a long-running command
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "sleep",
// 				Args:    []string{"30"},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteCommands(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)
// 		cancel()

// 		select {
// 		case <-done:
// 			// Success - service responded to cancellation
// 		case <-time.After(2 * time.Second):
// 			t.Fatal("service did not respond to context cancellation")
// 		}
// 	})

// 	t.Run("command_not_found", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "nonexistent_command_xyz",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		// Command not found returns an RPC error
// 		require.Error(t, err)
// 		assert.Equal(t, codes.NotFound, status.Code(err))
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		if !cfg.SupportsExecEnv {
// 			t.Skip("executor does not support env in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "printenv",
// 				Args:    []string{"MULTI_CMD_VAR"},
// 				Env:     map[string]string{"MULTI_CMD_VAR": "multi_cmd_value"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		assert.Contains(t, string(responses[0].Output.Stdout), "multi_cmd_value")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		if !cfg.SupportsExecWorkdir {
// 			t.Skip("executor does not support workdir in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		output := string(responses[0].Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse](ctx)

// 		// cat with nonexistent file writes error to stderr
// 		client.Send(&commandv1beta1.ExecuteCommandsRequest{
// 			Command: &commandv1beta1.ExecutableCommand{
// 				Command: "cat",
// 				Args:    []string{"/nonexistent_file_for_stderr_test"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		assert.NotEqual(t, int32(0), responses[0].Result.ExitCode)
// 		assert.Contains(t, string(responses[0].Output.Stderr), "nonexistent_file_for_stderr_test")
// 	})
// }

// func runExecuteBatchCommandsTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("multiple_commands", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{Command: "echo", Args: []string{"first"}},
// 				{Command: "echo", Args: []string{"second"}},
// 				{Command: "echo", Args: []string{"third"}},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 3)

// 		for i, result := range resp.Results {
// 			require.NotNil(t, result.Result, "result %d should have CommandResult", i)
// 			assert.Equal(t, int32(0), result.Result.ExitCode)
// 			require.NotNil(t, result.Output, "result %d should have output", i)
// 		}

// 		assert.Equal(t, "first\n", string(resp.Results[0].Output.Stdout))
// 		assert.Equal(t, "second\n", string(resp.Results[1].Output.Stdout))
// 		assert.Equal(t, "third\n", string(resp.Results[2].Output.Stdout))
// 	})

// 	t.Run("partial failure", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{Command: "echo", Args: []string{"first"}},
// 				{Command: "test", Args: []string{"1", "-eq", "0"}}, // fails
// 				{Command: "echo", Args: []string{"third"}},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err, "batch should succeed even with failing commands")
// 		require.Len(t, resp.Results, 3, "all commands should return results")

// 		// Verify first command succeeded
// 		assert.Equal(t, int32(0), resp.Results[0].Result.ExitCode)
// 		assert.Contains(t, string(resp.Results[0].Output.Stdout), "first")

// 		// Verify second command failed with non-zero exit
// 		assert.NotEqual(t, int32(0), resp.Results[1].Result.ExitCode)

// 		// Verify third command still ran and succeeded (failure didn't stop batch)
// 		assert.Equal(t, int32(0), resp.Results[2].Result.ExitCode)
// 		assert.Contains(t, string(resp.Results[2].Output.Stdout), "third")
// 	})

// 	t.Run("empty_batch", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		assert.Empty(t, resp.Results)
// 	})

// 	t.Run("command_not_found", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{Command: "nonexistent_command_xyz"},
// 			},
// 		}

// 		_, err := instance.ExecuteBatchCommands(ctx, req)
// 		// Command not found returns an RPC error
// 		require.Error(t, err)
// 		assert.Equal(t, codes.NotFound, status.Code(err))
// 	})

// 	t.Run("preserves_order", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{Command: "echo", Args: []string{"1"}},
// 				{Command: "echo", Args: []string{"2"}},
// 				{Command: "echo", Args: []string{"3"}},
// 				{Command: "echo", Args: []string{"4"}},
// 				{Command: "echo", Args: []string{"5"}},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 5)

// 		// Results should be in order even if executed concurrently
// 		for i, result := range resp.Results {
// 			expected := fmt.Sprintf("%d\n", i+1)
// 			assert.Equal(t, expected, string(result.Output.Stdout))
// 		}
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		if !cfg.SupportsExecEnv {
// 			t.Skip("executor does not support env in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{
// 					Command: "printenv",
// 					Args:    []string{"BATCH_CMD_VAR"},
// 					Env:     map[string]string{"BATCH_CMD_VAR": "batch_cmd_value"},
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		assert.Contains(t, string(resp.Results[0].Output.Stdout), "batch_cmd_value")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		if !cfg.SupportsExecWorkdir {
// 			t.Skip("executor does not support workdir in non-shell command execution")
// 		}
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{
// 					Command: "pwd",
// 					Workdir: "/tmp",
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		output := string(resp.Results[0].Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		// cat with nonexistent file writes error to stderr
// 		req := &commandv1beta1.ExecuteBatchCommandsRequest{
// 			Commands: []*commandv1beta1.ExecutableCommand{
// 				{
// 					Command: "cat",
// 					Args:    []string{"/nonexistent_file_for_stderr_test"},
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		assert.NotEqual(t, int32(0), resp.Results[0].Result.ExitCode)
// 		assert.Contains(t, string(resp.Results[0].Output.Stderr), "nonexistent_file_for_stderr_test")
// 	})
// }

// func runExecuteShellCommandsTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("stream_multiple_commands", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo first"},
// 		})
// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo second | tr 'a-z' 'A-Z'"},
// 		})
// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo third"},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 3)

// 		for i, resp := range responses {
// 			require.NotNil(t, resp.Result, "response %d should have result", i)
// 			assert.Equal(t, int32(0), resp.Result.ExitCode)
// 		}
// 	})

// 	t.Run("partial failure", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo first"},
// 		})
// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "test 1 -eq 0"}, // fails with non-zero exit
// 		})
// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo third"},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err, "stream should succeed even with failing commands")

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 3, "all commands should return results")

// 		// Verify results in order (streaming is sequential)
// 		assert.Equal(t, int32(0), responses[0].Result.ExitCode)
// 		assert.Contains(t, string(responses[0].Output.Stdout), "first")

// 		assert.NotEqual(t, int32(0), responses[1].Result.ExitCode)

// 		assert.Equal(t, int32(0), responses[2].Result.ExitCode)
// 		assert.Contains(t, string(responses[2].Output.Stdout), "third")
// 	})

// 	t.Run("client_immediate_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		assert.Empty(t, responses)
// 	})

// 	t.Run("client_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "sleep 30"},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteShellCommands(server)
// 			close(done)
// 		}()

// 		// Wait for command to start
// 		time.Sleep(100 * time.Millisecond)

// 		// Close stream (simulating client connection drop)
// 		client.Close()

// 		select {
// 		case <-done:
// 			// Service should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("service did not close cleanly when client disconnected during busy execution")
// 		}
// 	})

// 	t.Run("context_cancellation", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx, cancel := context.WithCancel(context.Background())
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "sleep 30"},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.ExecuteShellCommands(server)
// 			close(done)
// 		}()

// 		time.Sleep(50 * time.Millisecond)
// 		cancel()

// 		select {
// 		case <-done:
// 		case <-time.After(2 * time.Second):
// 			t.Fatal("service did not respond to context cancellation")
// 		}
// 	})

// 	t.Run("with_pipes", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{Command: "echo hello world | wc -w"},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		assert.Equal(t, int32(0), responses[0].Result.ExitCode)
// 		// wc -w output may have leading spaces
// 		assert.Contains(t, string(responses[0].Output.Stdout), "2")
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo $MULTI_SHELL_VAR",
// 				Env:     map[string]string{"MULTI_SHELL_VAR": "multi_shell_value"},
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		assert.Contains(t, string(responses[0].Output.Stdout), "multi_shell_value")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "pwd",
// 				Workdir: "/tmp",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		output := string(responses[0].Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse](ctx)

// 		client.Send(&commandv1beta1.ExecuteShellCommandsRequest{
// 			Command: &commandv1beta1.ShellCommand{
// 				Command: "echo stdout_multi_shell && echo stderr_multi_shell >&2",
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.ExecuteShellCommands(server)
// 		require.NoError(t, err)

// 		responses := client.DrainResponses()
// 		require.Len(t, responses, 1)
// 		assert.Contains(t, string(responses[0].Output.Stdout), "stdout_multi_shell")
// 		assert.Contains(t, string(responses[0].Output.Stderr), "stderr_multi_shell")
// 	})
// }

// func runExecuteBatchShellCommandsTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("multiple_commands", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{Command: "echo first"},
// 				{Command: "echo second | tr 'a-z' 'A-Z'"},
// 				{Command: "echo third"},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 3)

// 		assert.Equal(t, "first\n", string(resp.Results[0].Output.Stdout))
// 		assert.Equal(t, "SECOND\n", string(resp.Results[1].Output.Stdout))
// 		assert.Equal(t, "third\n", string(resp.Results[2].Output.Stdout))
// 	})

// 	t.Run("partial failure", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{Command: "echo first"},
// 				{Command: "test 1 -eq 0"}, // fails
// 				{Command: "echo third"},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err, "batch should succeed even with failing commands")
// 		require.Len(t, resp.Results, 3, "all commands should return results")

// 		// Verify first command succeeded
// 		assert.Equal(t, int32(0), resp.Results[0].Result.ExitCode)
// 		assert.Contains(t, string(resp.Results[0].Output.Stdout), "first")

// 		// Verify second command failed with non-zero exit
// 		assert.NotEqual(t, int32(0), resp.Results[1].Result.ExitCode)

// 		// Verify third command still ran and succeeded (failure didn't stop batch)
// 		assert.Equal(t, int32(0), resp.Results[2].Result.ExitCode)
// 		assert.Contains(t, string(resp.Results[2].Output.Stdout), "third")
// 	})

// 	t.Run("empty_batch", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		assert.Empty(t, resp.Results)
// 	})

// 	t.Run("with_pipes_and_redirects", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{Command: "echo hello | cat"},
// 				{Command: "echo world | tr 'a-z' 'A-Z'"},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 2)

// 		assert.Equal(t, "hello\n", string(resp.Results[0].Output.Stdout))
// 		assert.Equal(t, "WORLD\n", string(resp.Results[1].Output.Stdout))
// 	})

// 	t.Run("preserves_order", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{Command: "echo 1"},
// 				{Command: "echo 2"},
// 				{Command: "echo 3"},
// 				{Command: "echo 4"},
// 				{Command: "echo 5"},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 5)

// 		for i, result := range resp.Results {
// 			expected := fmt.Sprintf("%d\n", i+1)
// 			assert.Equal(t, expected, string(result.Output.Stdout))
// 		}
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{
// 					Command: "echo $BATCH_SHELL_VAR",
// 					Env:     map[string]string{"BATCH_SHELL_VAR": "batch_shell_value"},
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		assert.Contains(t, string(resp.Results[0].Output.Stdout), "batch_shell_value")
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{
// 					Command: "pwd",
// 					Workdir: "/tmp",
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		output := string(resp.Results[0].Output.Stdout)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("stderr_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		req := &commandv1beta1.ExecuteBatchShellCommandsRequest{
// 			Commands: []*commandv1beta1.ShellCommand{
// 				{
// 					Command: "echo stdout_batch_shell && echo stderr_batch_shell >&2",
// 				},
// 			},
// 		}

// 		resp, err := instance.ExecuteBatchShellCommands(ctx, req)
// 		require.NoError(t, err)
// 		require.Len(t, resp.Results, 1)
// 		assert.Contains(t, string(resp.Results[0].Output.Stdout), "stdout_batch_shell")
// 		assert.Contains(t, string(resp.Results[0].Output.Stderr), "stderr_batch_shell")
// 	})
// }

// func runTerminalSessionTests(t *testing.T, cfg ServiceTestConfig) {
// 	t.Run("basic_shell_interaction", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		// Start terminal session
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Send a command
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("echo hello\n"),
// 			},
// 		})

// 		time.Sleep(100 * time.Millisecond)

// 		// Exit the terminal
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout waiting for terminal session to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		// Verify we got output
// 		responses := client.DrainResponses()
// 		require.NotEmpty(t, responses)

// 		var totalOutput []byte
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalOutput = append(totalOutput, resp.Output.Stdout...)
// 			}
// 		}
// 		assert.Contains(t, string(totalOutput), "hello")
// 	})

// 	t.Run("pty_dimensions", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 120,
// 					Rows: 40,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Check dimensions with stty
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("stty size\n"),
// 			},
// 		})

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		var totalOutput []byte
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalOutput = append(totalOutput, resp.Output.Stdout...)
// 			}
// 		}
// 		// stty size returns "rows cols"
// 		assert.Contains(t, string(totalOutput), "40 120")
// 	})

// 	t.Run("pty_resize", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Resize the terminal
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Resize: &commandv1beta1.TerminalSessionRequest_ResizeEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 200,
// 					Rows: 50,
// 				},
// 			},
// 		})

// 		time.Sleep(50 * time.Millisecond)

// 		// Check new dimensions
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("stty size\n"),
// 			},
// 		})

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		var totalOutput []byte
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalOutput = append(totalOutput, resp.Output.Stdout...)
// 			}
// 		}
// 		assert.Contains(t, string(totalOutput), "50 200")
// 	})

// 	t.Run("send_signal_SIGINT", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Send SIGINT - for PTY sessions, this closes the session
// 		// (signals don't reach the foreground process group)
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Signal: commandv1beta1.Signal_SIGNAL_SIGINT,
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("terminal did not terminate after SIGINT")
// 		}
// 	})

// 	t.Run("send_signal_SIGTERM", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Send SIGTERM - for PTY sessions, this closes the session
// 		// (signals don't reach the foreground process group)
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Signal: commandv1beta1.Signal_SIGNAL_SIGTERM,
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("terminal did not terminate after SIGTERM")
// 		}
// 	})

// 	t.Run("client_immediate_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		// Immediately disconnect after sending start
// 		client.Close()

// 		done := make(chan struct{})
// 		go func() {
// 			instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		select {
// 		case <-done:
// 			// Terminal session should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("terminal session did not close cleanly")
// 		}
// 	})

// 	t.Run("client_disconnect", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		done := make(chan struct{})
// 		go func() {
// 			instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		// Wait for session to start, then send a long-running command
// 		time.Sleep(100 * time.Millisecond)
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("sleep 30\n"),
// 			},
// 		})

// 		// Wait a bit for the sleep command to start
// 		time.Sleep(100 * time.Millisecond)

// 		// Close stream (simulating client connection drop)
// 		client.Close()

// 		select {
// 		case <-done:
// 			// Terminal session should clean up and exit
// 		case <-time.After(3 * time.Second):
// 			t.Fatal("terminal session did not close cleanly when client disconnected during busy execution")
// 		}
// 	})

// 	t.Run("with_workdir", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Workdir: "/tmp",
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("pwd\n"),
// 			},
// 		})

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		var totalOutput []byte
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalOutput = append(totalOutput, resp.Output.Stdout...)
// 			}
// 		}
// 		// macOS resolves /tmp to /private/tmp
// 		output := string(totalOutput)
// 		assert.True(t, strings.Contains(output, "/tmp") || strings.Contains(output, "/private/tmp"),
// 			"expected /tmp in output, got: %s", output)
// 	})

// 	t.Run("with_env", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Env: map[string]string{"MY_TEST_VAR": "terminal_test_value"},
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("echo $MY_TEST_VAR\n"),
// 			},
// 		})

// 		time.Sleep(100 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(5 * time.Second):
// 			t.Fatal("timeout")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		var totalOutput []byte
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalOutput = append(totalOutput, resp.Output.Stdout...)
// 			}
// 		}
// 		assert.Contains(t, string(totalOutput), "terminal_test_value")
// 	})

// 	t.Run("no_start_event", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		// Send input without start event
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("echo hello\n"),
// 			},
// 		})
// 		client.CloseSend()

// 		err := instance.TerminalSession(server)
// 		require.Error(t, err)
// 		assert.Contains(t, err.Error(), "start")
// 	})

// 	t.Run("large_output", func(t *testing.T) {
// 		instance := cfg.NewInstance(t)
// 		defer instance.Close()

// 		ctx := context.Background()
// 		client, server := test.NewMockBidiStream[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse](ctx)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Start: &commandv1beta1.TerminalSessionRequest_StartEvent{
// 				Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
// 					Cols: 80,
// 					Rows: 24,
// 				},
// 			},
// 		})

// 		var serviceErr error
// 		done := make(chan struct{})
// 		go func() {
// 			serviceErr = instance.TerminalSession(server)
// 			close(done)
// 		}()

// 		time.Sleep(100 * time.Millisecond)

// 		// Generate large output
// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("for i in $(seq 1 500); do echo 'Line '$i' with some padding text to make it longer'; done\n"),
// 			},
// 		})

// 		time.Sleep(500 * time.Millisecond)

// 		client.Send(&commandv1beta1.TerminalSessionRequest{
// 			Input: &commandv1beta1.CommandInput{
// 				Stdin: []byte("exit\n"),
// 			},
// 		})

// 		select {
// 		case <-done:
// 		case <-time.After(10 * time.Second):
// 			t.Fatal("timeout waiting for terminal session to complete")
// 		}

// 		require.NoError(t, serviceErr)

// 		responses := client.DrainResponses()
// 		var totalBytes int
// 		for _, resp := range responses {
// 			if resp.Output != nil {
// 				totalBytes += len(resp.Output.Stdout)
// 			}
// 		}
// 		t.Logf("Received %d bytes from terminal", totalBytes)
// 		assert.Greater(t, totalBytes, 5000, "should have received significant output")
// 	})
// }

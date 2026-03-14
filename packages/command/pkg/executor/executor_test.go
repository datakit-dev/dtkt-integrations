package executor

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ExecutorTestConfig holds configuration for the shared test suite
type ExecutorTestConfig struct {
	// Name is the executor name for test output
	Name string

	// NewExecutor creates a new executor instance for testing
	NewExecutor func(t *testing.T) CommandExecutor

	// SupportsExecWorkdir indicates if ExecuteCommand supports workdir
	SupportsExecWorkdir bool

	// SupportsExecEnv indicates if ExecuteCommand supports environment variables
	SupportsExecEnv bool

	// Skip can be used to skip certain tests
	Skip func(t *testing.T)
}

// RunExecutorTestSuite runs the shared test suite against any CommandExecutor implementation
func RunExecutorTestSuite(t *testing.T, cfg ExecutorTestConfig) {
	if cfg.Skip != nil {
		cfg.Skip(t)
	}

	t.Run("ExecuteCommand", func(t *testing.T) {
		runExecuteCommandTests(t, cfg)
	})

	t.Run("ExecuteShellCommand", func(t *testing.T) {
		runExecuteShellCommandTests(t, cfg)
	})

	t.Run("TerminalSession", func(t *testing.T) {
		runTerminalSessionTests(t, cfg)
	})
}

// ExecuteCommand tests

func runExecuteCommandTests(t *testing.T, cfg ExecutorTestConfig) {
	t.Run("Success", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "echo",
			Args:    []string{"hello"},
		}

		cio := &CommandIO{
			Stdin:  nil,
			Stdout: stdout,
			Stderr: stderr,
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, reason, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}
		if reason != commandv1beta1.CommandResult_REASON_COMPLETED {
			t.Errorf("expected REASON_COMPLETED, got %v", reason)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "hello" {
			t.Errorf("expected output 'hello', got '%s'", output)
		}
	})

	t.Run("WithArgs", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "echo",
			Args:    []string{"hello", "world"},
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "hello world" {
			t.Errorf("expected output 'hello world', got '%s'", output)
		}
	})

	t.Run("WithEnv", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sh",
			Args:    []string{"-c", "echo $TEST_VAR"},
			Env:     map[string]string{"TEST_VAR": "test_value"},
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)

		if !cfg.SupportsExecEnv {
			if err == nil {
				t.Error("expected error when env is specified, got nil")
			}
			return
		}

		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "test_value" {
			t.Errorf("expected output 'test_value', got '%s'", output)
		}
	})

	t.Run("WithWorkdir", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "pwd",
			Workdir: "/tmp",
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)

		if !cfg.SupportsExecWorkdir {
			if err == nil {
				t.Error("expected error when workdir is specified, got nil")
			}
			return
		}

		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		// macOS may resolve /tmp to /private/tmp
		if output != "/tmp" && output != "/private/tmp" {
			t.Errorf("expected output '/tmp' or '/private/tmp', got '%s'", output)
		}
	})

	t.Run("NonZeroExit", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		// Use test command which exits 1 when expression is false
		cmd := &commandv1beta1.ExecutableCommand{
			Command: "test",
			Args:    []string{"1", "-eq", "0"},
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, _, _ := session.Wait()

		if exitCode == 0 {
			t.Error("expected non-zero exit code")
		}
	})

	t.Run("WithTimeout", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sleep",
			Args:    []string{"10"},
			Timeout: durationpb.New(100 * time.Millisecond),
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, reason, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != -1 {
			t.Errorf("expected exit code -1, got %d", exitCode)
		}
		if reason != commandv1beta1.CommandResult_REASON_TIMED_OUT {
			t.Errorf("expected REASON_TIMED_OUT, got %v", reason)
		}
	})

	t.Run("StdinInput", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdin := strings.NewReader("hello from stdin")
		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "cat",
		}

		cio := &CommandIO{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := stdout.String()
		if output != "hello from stdin" {
			t.Errorf("expected output 'hello from stdin', got '%s'", output)
		}
	})

	t.Run("NilStdin_CatGetsEOF", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "cat",
		}

		cio := &CommandIO{
			Stdin:  nil,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		// Should complete quickly (not hang) because stdin is nil → EOF
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - cat got EOF and exited
		case <-time.After(2 * time.Second):
			t.Fatal("cat hung - nil stdin did not provide EOF")
		}

		// cat with no input produces empty output
		if stdout.String() != "" {
			t.Errorf("expected empty output, got '%s'", stdout.String())
		}
	})

	t.Run("Signal_SIGINT", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sleep",
			Args:    []string{"10"},
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGINT
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGINT)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was interrupted)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGINT")
		}
	})

	t.Run("Signal_SIGTERM", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sleep",
			Args:    []string{"10"},
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGTERM
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGTERM)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was terminated)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGTERM")
		}
	})

	t.Run("Signal_SIGKILL", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sleep",
			Args:    []string{"10"},
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGKILL
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGKILL)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was killed)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGKILL")
		}
	})

	t.Run("Close", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ExecutableCommand{
			Command: "sleep",
			Args:    []string{"10"},
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteCommand(context.Background(), cmd, cio)
		if err != nil {
			t.Fatalf("ExecuteCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Close should terminate the process
		err = session.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Wait should complete (process was killed)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after Close")
		}
	})
}

// ExecuteShellCommand tests

func runExecuteShellCommandTests(t *testing.T, cfg ExecutorTestConfig) {
	t.Run("Success", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "echo hello",
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "hello" {
			t.Errorf("expected output 'hello', got '%s'", output)
		}
	})

	t.Run("WithPipe", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "echo hello world | tr ' ' '-'",
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "hello-world" {
			t.Errorf("expected output 'hello-world', got '%s'", output)
		}
	})

	t.Run("WithEnv", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "echo $MY_VAR",
			Env:     map[string]string{"MY_VAR": "shell_value"},
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		if output != "shell_value" {
			t.Errorf("expected output 'shell_value', got '%s'", output)
		}
	})

	t.Run("WithWorkdir", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "pwd",
			Workdir: "/tmp",
		}

		cio := &CommandIO{
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := strings.TrimSpace(stdout.String())
		// macOS may resolve /tmp to /private/tmp
		if output != "/tmp" && output != "/private/tmp" {
			t.Errorf("expected output '/tmp' or '/private/tmp', got '%s'", output)
		}
	})

	t.Run("NonZeroExit", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ShellCommand{
			Command: "exit 7",
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, _ := session.Wait()

		if exitCode != 7 {
			t.Errorf("expected exit code 7, got %d", exitCode)
		}
	})

	t.Run("StdinInput", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdin := strings.NewReader("hello from shell stdin")
		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "cat",
		}

		cio := &CommandIO{
			Stdin:  stdin,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := stdout.String()
		if output != "hello from shell stdin" {
			t.Errorf("expected output 'hello from shell stdin', got '%s'", output)
		}
	})

	t.Run("NilStdin_CatGetsEOF", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		cmd := &commandv1beta1.ShellCommand{
			Command: "cat",
		}

		cio := &CommandIO{
			Stdin:  nil,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		// Should complete quickly (not hang) because stdin is nil → EOF
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - cat got EOF and exited
		case <-time.After(2 * time.Second):
			t.Fatal("cat hung - nil stdin did not provide EOF")
		}

		// cat with no input produces empty output
		if stdout.String() != "" {
			t.Errorf("expected empty output, got '%s'", stdout.String())
		}
	})

	t.Run("Signal_SIGINT", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ShellCommand{
			Command: "sleep 10",
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGINT
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGINT)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was interrupted)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGINT")
		}
	})

	t.Run("Signal_SIGTERM", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ShellCommand{
			Command: "sleep 10",
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGTERM
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGTERM)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was terminated)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGTERM")
		}
	})

	t.Run("Signal_SIGKILL", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ShellCommand{
			Command: "sleep 10",
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Send SIGKILL
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGKILL)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		// Wait should complete (process was killed)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after SIGKILL")
		}
	})

	t.Run("Close", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		cmd := &commandv1beta1.ShellCommand{
			Command: "sleep 10",
		}

		cio := &CommandIO{
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.ExecuteShellCommand(context.Background(), cmd, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("ExecuteShellCommand failed: %v", err)
		}

		// Give the process time to start
		time.Sleep(100 * time.Millisecond)

		// Close should terminate the process
		err = session.Close()
		if err != nil {
			t.Fatalf("Close failed: %v", err)
		}

		// Wait should complete (process was killed)
		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - process terminated
		case <-time.After(2 * time.Second):
			t.Error("process did not terminate after Close")
		}
	})
}

// TerminalSession tests

func runTerminalSessionTests(t *testing.T, cfg ExecutorTestConfig) {
	t.Run("BasicShell", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		// Send commands with delays like a real user would
		time.Sleep(100 * time.Millisecond)
		//nolint:errcheck
		stdinW.Write([]byte("echo hello\n"))
		time.Sleep(100 * time.Millisecond)
		//nolint:errcheck
		stdinW.Write([]byte("echo world\n"))
		time.Sleep(100 * time.Millisecond)
		//nolint:errcheck
		stdinW.Write([]byte("exit\n"))
		//nolint:errcheck
		//nolint:errcheck
		stdinW.Close()

		exitCode, _, err := session.Wait()
		if err != nil {
			t.Fatalf("Wait failed: %v", err)
		}

		if exitCode != 0 {
			t.Errorf("expected exit code 0, got %d", exitCode)
		}

		output := stdout.String()
		if !strings.Contains(output, "hello") {
			t.Errorf("expected output to contain 'hello', got '%s'", output)
		}
		if !strings.Contains(output, "world") {
			t.Errorf("expected output to contain 'world', got '%s'", output)
		}
	})

	t.Run("WithWorkdir", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{
			Workdir: "/tmp",
		}

		cio := &CommandIO{
			Stdin:  strings.NewReader("pwd\nexit\n"),
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		//nolint:errcheck
		//nolint:errcheck
		session.Wait()

		output := stdout.String()
		if !strings.Contains(output, "/tmp") && !strings.Contains(output, "/private/tmp") {
			t.Errorf("expected output to contain '/tmp', got '%s'", output)
		}
	})

	t.Run("WithEnv", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{
			Env: map[string]string{"MY_TEST_VAR": "terminal_value"},
		}

		cio := &CommandIO{
			Stdin:  strings.NewReader("echo $MY_TEST_VAR\nexit\n"),
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		//nolint:errcheck
		session.Wait()

		output := stdout.String()
		if !strings.Contains(output, "terminal_value") {
			t.Errorf("expected output to contain 'terminal_value', got '%s'", output)
		}
	})

	t.Run("NilStdin_ReturnsError", func(t *testing.T) {
		// Terminal sessions require stdin - it's interactive by definition
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		cio := &CommandIO{
			Stdin:  nil,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		_, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err == nil {
			t.Error("expected error for nil stdin, got nil")
		}
	})

	t.Run("PTY_Dimensions", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{
			Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
				Rows: 24,
				Cols: 80,
			},
		}

		cio := &CommandIO{
			Stdin:  strings.NewReader("stty size\nexit\n"),
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		//nolint:errcheck
		//nolint:errcheck
		session.Wait()

		output := stdout.String()
		// With PTY, "stty size" should return "24 80"
		if !strings.Contains(output, "24 80") {
			t.Errorf("expected dimensions '24 80', got '%s'", output)
		}
	})

	t.Run("PTY_Resize", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{
			Dimensions: &commandv1beta1.TerminalSessionRequest_Dimensions{
				Rows: 24,
				Cols: 80,
			},
		}

		stdinR, stdinW := io.Pipe()
		//nolint:errcheck
		//nolint:errcheck
		defer stdinW.Close()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}
		//nolint:errcheck
		defer session.Close()

		// Resize should work with PTY (rows, cols)
		err = session.Resize(40, 120)
		if err != nil {
			t.Errorf("Resize failed: %v", err)
		}
	})

	t.Run("PTY_RawMode", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		cio := &CommandIO{
			Stdin:  strings.NewReader("tty\nexit\n"),
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		//nolint:errcheck
		session.Wait()

		output := stdout.String()
		// With PTY, tty command should return a pty device path
		// Without PTY, it returns "not a tty"
		if strings.Contains(output, "not a tty") {
			t.Errorf("expected PTY to be allocated, but got 'not a tty': %s", output)
		}
		if !strings.Contains(output, "/dev/") {
			t.Errorf("expected PTY device path like /dev/ttys* or /dev/pts/*, got: %s", output)
		}
	})

	t.Run("PTY_ControlCharacters_CtrlC", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		// Start sleep command
		//nolint:errcheck
		stdinW.Write([]byte("sleep 30\n"))
		time.Sleep(200 * time.Millisecond)

		// Send Ctrl+C (ASCII 0x03)
		//nolint:errcheck
		stdinW.Write([]byte{0x03})
		time.Sleep(100 * time.Millisecond)

		// Check exit code - SIGINT (2) should give 128+2=130
		//nolint:errcheck
		stdinW.Write([]byte("echo $?\n"))
		time.Sleep(100 * time.Millisecond)

		// Send exit to terminate the shell
		//nolint:errcheck
		stdinW.Write([]byte("exit\n"))
		//nolint:errcheck
		//nolint:errcheck
		stdinW.Close()

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// With PTY, Ctrl+C should have interrupted sleep with exit code 130
			output := stdout.String()
			if !strings.Contains(output, "130") {
				t.Errorf("expected exit code 130 (SIGINT) in output, got: %s", output)
			}
		case <-time.After(2 * time.Second):
			//nolint:errcheck
			session.Close()
			t.Error("Ctrl+C did not interrupt the process - PTY not handling control characters")
		}
	})

	t.Run("PTY_ControlCharacters_CtrlD", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		// Wait for shell to start
		time.Sleep(200 * time.Millisecond)

		// Send Ctrl+D (ASCII 0x04) - EOF, should exit shell cleanly
		//nolint:errcheck
		stdinW.Write([]byte{0x04})
		//nolint:errcheck
		stdinW.Close()

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Ctrl+D should have exited the shell cleanly
		case <-time.After(2 * time.Second):
			//nolint:errcheck
			session.Close()
			t.Error("Ctrl+D did not exit the shell - PTY not handling EOF")
		}
	})

	t.Run("PTY_ControlCharacters_CtrlBackslash", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		stdout := &bytes.Buffer{}

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: stdout,
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		// Start sleep command
		//nolint:errcheck
		stdinW.Write([]byte("sleep 30\n"))
		time.Sleep(200 * time.Millisecond)

		// Send Ctrl+\ (ASCII 0x1C) - SIGQUIT
		//nolint:errcheck
		stdinW.Write([]byte{0x1C})
		time.Sleep(100 * time.Millisecond)

		// Check exit code - SIGQUIT (3) should give 128+3=131
		//nolint:errcheck
		stdinW.Write([]byte("echo $?\n"))
		time.Sleep(100 * time.Millisecond)

		// Send exit to terminate the shell
		//nolint:errcheck
		stdinW.Write([]byte("exit\n"))
		//nolint:errcheck
		stdinW.Close()

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// With PTY, Ctrl+\ should have quit sleep with exit code 131
			output := stdout.String()
			if !strings.Contains(output, "131") {
				t.Errorf("expected exit code 131 (SIGQUIT) in output, got: %s", output)
			}
		case <-time.After(2 * time.Second):
			//nolint:errcheck
			session.Close()
			t.Error("Ctrl+\\ did not quit the process - PTY not handling control characters")
		}
	})

	t.Run("Signal_SIGINT", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()
		//nolint:errcheck
		defer stdinW.Close()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// For PTY sessions, Signal() closes the session (signals can't reach foreground process group)
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGINT)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - session was terminated
		case <-time.After(2 * time.Second):
			t.Error("terminal session didn't terminate after Signal")
		}
	})

	t.Run("Signal_SIGTERM", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()
		//nolint:errcheck
		defer stdinW.Close()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// For PTY sessions, Signal() closes the session (signals can't reach foreground process group)
		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGTERM)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - session was terminated
		case <-time.After(2 * time.Second):
			t.Error("terminal session didn't terminate after Signal")
		}
	})

	t.Run("Signal_SIGKILL", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()
		//nolint:errcheck
		defer stdinW.Close()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		err = session.Signal(commandv1beta1.Signal_SIGNAL_SIGKILL)
		if err != nil {
			t.Logf("Signal failed (may be expected on some implementations): %v", err)
		}

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			//nolint:errcheck
			session.Close()
			t.Error("terminal session didn't terminate after SIGKILL")
		}
	})

	t.Run("Close", func(t *testing.T) {
		exec := cfg.NewExecutor(t)
		//nolint:errcheck
		defer exec.Close()

		startEvent := &commandv1beta1.TerminalSessionRequest_StartEvent{}

		stdinR, stdinW := io.Pipe()
		//nolint:errcheck
		defer stdinW.Close()

		cio := &CommandIO{
			Stdin:  stdinR,
			Stdout: &bytes.Buffer{},
			Stderr: &bytes.Buffer{},
		}

		session, err := exec.TerminalSession(context.Background(), startEvent, "/bin/sh", cio)
		if err != nil {
			t.Fatalf("TerminalSession failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Close the session - should terminate the terminal
		err = session.Close()
		if err != nil {
			t.Logf("Close returned error (may be expected): %v", err)
		}

		done := make(chan struct{})
		go func() {
			//nolint:errcheck
			session.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success - terminal session was closed
		case <-time.After(2 * time.Second):
			t.Error("terminal session didn't terminate after Close")
		}
	})
}

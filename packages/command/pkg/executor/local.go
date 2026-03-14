package executor

import (
	"context"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/creack/pty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// LocalExecutor runs commands locally.
type LocalExecutor struct{}

var _ CommandExecutor = (*LocalExecutor)(nil)

func NewLocalExecutor() *LocalExecutor {
	return &LocalExecutor{}
}

func (e *LocalExecutor) ExecuteCommand(
	ctx context.Context,
	command *commandv1beta1.ExecutableCommand,
	cio *CommandIO,
) (Session, error) {
	var cancel context.CancelFunc = func() {}
	if command.GetTimeout() != nil {
		ctx, cancel = context.WithTimeout(ctx, command.GetTimeout().AsDuration())
	}

	cmd := exec.CommandContext(ctx, command.Command, command.Args...)
	if command.Workdir != "" {
		cmd.Dir = command.Workdir
	}
	cmd.Env = mergeEnv(command.Env)

	cmd.Stdin = cio.Stdin
	cmd.Stdout = cio.Stdout
	cmd.Stderr = cio.Stderr

	err := cmd.Start()
	if err != nil {
		cancel()
		return nil, status.Errorf(codes.Internal, "failed to start command: %v", err)
	}

	return &localSession{ctx: ctx, cmd: cmd, cancel: cancel}, nil
}

func (e *LocalExecutor) ExecuteShellCommand(
	ctx context.Context,
	command *commandv1beta1.ShellCommand,
	shellCmd string,
	cio *CommandIO,
) (Session, error) {
	var cancel context.CancelFunc = func() {}
	if command.GetTimeout() != nil {
		ctx, cancel = context.WithTimeout(ctx, command.GetTimeout().AsDuration())
	}

	cmd := exec.CommandContext(ctx, shellCmd, "-c", command.Command)
	if command.Workdir != "" {
		cmd.Dir = command.Workdir
	}
	cmd.Env = mergeEnv(command.Env)

	cmd.Stdin = cio.Stdin
	cmd.Stdout = cio.Stdout
	cmd.Stderr = cio.Stderr

	err := cmd.Start()
	if err != nil {
		cancel()
		return nil, status.Errorf(codes.Internal, "failed to start command: %v", err)
	}

	return &localSession{ctx: ctx, cmd: cmd, cancel: cancel}, nil
}

func (e *LocalExecutor) TerminalSession(
	ctx context.Context,
	startEvent *commandv1beta1.TerminalSessionRequest_StartEvent,
	shellCmd string,
	cio *CommandIO,
) (Session, error) {
	if cio.Stdin == nil {
		return nil, status.Error(codes.InvalidArgument, "stdin is required for terminal sessions")
	}

	cmd := exec.CommandContext(ctx, shellCmd)
	if startEvent.Workdir != "" {
		cmd.Dir = startEvent.Workdir
	}
	cmd.Env = mergeEnv(startEvent.Env)
	cmd.Env = append(cmd.Env, "TERM=xterm-256color", "COLORTERM=truecolor")

	// Set initial window size from dimensions
	winSize := &pty.Winsize{
		Rows: 24,
		Cols: 80,
	}
	if startEvent.Dimensions != nil {
		winSize.Rows = uint16(startEvent.Dimensions.Rows)
		winSize.Cols = uint16(startEvent.Dimensions.Cols)
	}

	// Start command with PTY
	ptmx, err := pty.StartWithSize(cmd, winSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start PTY: %v", err)
	}

	// Copy stdin to PTY (in background)
	go func() {
		_, err := io.Copy(ptmx, cio.Stdin)
		if err != nil {
			log.Error(ctx, "copy stdin to PTY", log.Err(err))
		}
	}()

	// Copy PTY output to stdout
	go func() {
		_, err := io.Copy(cio.Stdout, ptmx)
		if err != nil {
			log.Error(ctx, "copy PTY to stdout", log.Err(err))
		}
	}()

	return &localSession{ctx: ctx, cmd: cmd, ptyFile: ptmx, isPTY: true}, nil
}

func (e *LocalExecutor) Close() error {
	return nil
}

func mergeEnv(overrides map[string]string) []string {
	env := os.Environ()
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	return env
}

type localSession struct {
	ctx     context.Context
	cmd     *exec.Cmd
	cancel  context.CancelFunc
	ptyFile *os.File // PTY master file descriptor (nil for non-PTY sessions)
	isPTY   bool     // Whether this is a PTY session
}

func (s *localSession) Wait() (exitCode int, reason commandv1beta1.CommandResult_Reason, err error) {
	// If context has a deadline (timeout), we need to close stdin when context is done
	// to unblock cmd.Wait() which waits for stdin copying to complete.
	if _, hasDeadline := s.ctx.Deadline(); hasDeadline && s.cmd.Stdin != nil {
		go func() {
			<-s.ctx.Done()
			// Close stdin pipe reader to unblock cmd.Wait()
			if closer, ok := s.cmd.Stdin.(io.Closer); ok {
				//nolint:errcheck
				closer.Close()
			}
		}()
	}

	err = s.cmd.Wait()

	// Close PTY if present
	if s.ptyFile != nil {
		//nolint:errcheck
		s.ptyFile.Close()
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if s.ctx.Err() == context.DeadlineExceeded {
				return -1, commandv1beta1.CommandResult_REASON_TIMED_OUT, nil
			}

			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				return -1, commandv1beta1.CommandResult_REASON_KILLED, nil
			}

			return exitErr.ExitCode(), commandv1beta1.CommandResult_REASON_UNSPECIFIED, nil
		}
		return 0, 0, status.Errorf(codes.Internal, "failed to wait for command: %v", err)
	}

	return 0, commandv1beta1.CommandResult_REASON_COMPLETED, nil
}

func (s *localSession) Resize(rows, cols int) error {
	if s.ptyFile == nil {
		return status.Errorf(codes.FailedPrecondition, "resize not supported for non-PTY sessions")
	}
	return pty.Setsize(s.ptyFile, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (s *localSession) Signal(signal commandv1beta1.Signal) error {
	// For PTY sessions, signals don't reach the foreground process group,
	// so we just close the session instead (same behavior as ExecuteShellCommand timeout).
	if s.isPTY {
		return s.Close()
	}

	var sig syscall.Signal

	switch signal {
	case commandv1beta1.Signal_SIGNAL_SIGINT:
		sig = syscall.SIGINT
	case commandv1beta1.Signal_SIGNAL_SIGTERM:
		sig = syscall.SIGTERM
	case commandv1beta1.Signal_SIGNAL_SIGKILL:
		sig = syscall.SIGKILL
	default:
		return status.Errorf(codes.InvalidArgument, "unknown signal: %v", signal)
	}

	return s.cmd.Process.Signal(sig)
}

func (s *localSession) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	if s.ptyFile != nil {
		//nolint:errcheck
		s.ptyFile.Close()
	}

	if s.cmd.Process != nil && (s.cmd.ProcessState == nil || !s.cmd.ProcessState.Exited()) {
		if err := s.cmd.Process.Kill(); err != nil {
			return status.Errorf(codes.Internal, "failed to close session (kill process): %v", err)
		}
	}
	return nil
}

package executor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// RemoteExecutor runs commands via an SSH client.
type RemoteExecutor struct {
	sshConfig       *commandv1beta1.SSHConfig
	sshClientConfig *ssh.ClientConfig
	sshClient       *ssh.Client
}

var _ CommandExecutor = (*RemoteExecutor)(nil)

func NewRemoteExecutor(sshConfig *commandv1beta1.SSHConfig) (e *RemoteExecutor, err error) {
	e = &RemoteExecutor{
		sshConfig: sshConfig,
	}

	// Build SSH client config
	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		return nil, err
	}
	e.sshClientConfig = clientConfig

	// Establish initial SSH connection
	client, err := dialSSH(context.Background(), e.sshConfig, e.sshClientConfig)
	if err != nil {
		return nil, err
	}
	e.sshClient = client

	return e, nil
}

func (e *RemoteExecutor) ExecuteCommand(
	ctx context.Context,
	command *commandv1beta1.ExecutableCommand,
	cio *CommandIO,
) (Session, error) {
	var cancel context.CancelFunc = func() {}
	if command.GetTimeout() != nil {
		ctx, cancel = context.WithTimeout(ctx, command.GetTimeout().AsDuration())
	}

	session, err := e.newSession(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	if command.Workdir != "" {
		cancel()
		//nolint:errcheck
		session.Close()
		return nil, status.Errorf(codes.InvalidArgument, "workdir is not supported in non-shell remote command execution")
	}

	if len(command.Env) > 0 {
		cancel()
		//nolint:errcheck
		session.Close()
		return nil, status.Errorf(codes.InvalidArgument, "env is not supported in non-shell remote command execution")
	}

	session.Stdin = cio.Stdin
	session.Stdout = cio.Stdout
	session.Stderr = cio.Stderr

	// Build command with properly quoted arguments
	parts := []string{command.Command}
	for _, arg := range command.Args {
		parts = append(parts, strconv.Quote(arg))
	}
	fullCmd := strings.Join(parts, " ")

	err = session.Start(fullCmd)
	if err != nil {
		cancel()
		//nolint:errcheck
		session.Close()
		return nil, status.Errorf(codes.Internal, "failed to start remote command: %v", err)
	}

	return &remoteSession{ctx: ctx, cancel: cancel, session: session}, nil
}

func (e *RemoteExecutor) ExecuteShellCommand(
	ctx context.Context,
	command *commandv1beta1.ShellCommand,
	shellCmd string,
	cio *CommandIO,
) (Session, error) {
	var cancel context.CancelFunc = func() {}
	if command.GetTimeout() != nil {
		ctx, cancel = context.WithTimeout(ctx, command.GetTimeout().AsDuration())
	}

	session, err := e.newSession(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	// Build command with env vars and workdir prefix
	// SSH session.Start() runs through the remote shell, so we build a single
	// command string that sets env vars, changes dir, then runs the user command
	var fullCmd strings.Builder
	for k, v := range command.Env {
		fmt.Fprintf(&fullCmd, "export %s=%s; ", k, strconv.Quote(v))
	}
	if command.Workdir != "" {
		fmt.Fprintf(&fullCmd, "cd %s && ", strconv.Quote(command.Workdir))
	}

	fullCmd.WriteString(command.Command)

	session.Stdin = cio.Stdin
	session.Stdout = cio.Stdout
	session.Stderr = cio.Stderr

	err = session.Start(fullCmd.String())
	if err != nil {
		cancel()
		//nolint:errcheck
		session.Close()
		return nil, status.Errorf(codes.Internal, "failed to start remote command: %v", err)
	}

	return &remoteSession{ctx: ctx, cancel: cancel, session: session}, nil
}

func (e *RemoteExecutor) TerminalSession(
	ctx context.Context,
	startEvent *commandv1beta1.TerminalSessionRequest_StartEvent,
	shellCmd string,
	cio *CommandIO,
) (Session, error) {
	if cio.Stdin == nil {
		return nil, status.Error(codes.InvalidArgument, "stdin is required for terminal sessions")
	}

	session, err := e.newSession(ctx)
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// Default dimensions if not specified
	rows := 24
	cols := 80
	if d := startEvent.GetDimensions(); d != nil {
		rows = int(d.Rows)
		cols = int(d.Cols)
	}

	err = session.RequestPty("xterm-256color", rows, cols, modes)
	if err != nil {
		//nolint:errcheck
		session.Close()
		return nil, err
	}

	fullCmd := ""
	for k, v := range startEvent.Env {
		fullCmd += fmt.Sprintf("export %s=%q; ", k, v)
	}
	if startEvent.Workdir != "" {
		fullCmd += fmt.Sprintf("cd %s; ", startEvent.Workdir)
	}

	fullCmd += shellCmd

	session.Stdin = cio.Stdin
	session.Stdout = cio.Stdout
	session.Stderr = cio.Stderr

	err = session.Start(fullCmd)
	if err != nil {
		//nolint:errcheck
		session.Close()
		return nil, err
	}

	return &remoteSession{ctx: ctx, session: session, isPTY: true}, nil
}

func (e *RemoteExecutor) Close() error {
	if e.sshClient == nil {
		return nil
	}
	return e.sshClient.Close()
}

// newSession attempts to create a new SSH session, reconnecting if necessary.
func (e *RemoteExecutor) newSession(ctx context.Context) (*ssh.Session, error) {
	session, err := e.sshClient.NewSession()
	if err != nil {
		// Connection failed, try to reconnect
		if client := e.sshClient; client != nil {
			//nolint:errcheck
			client.Close()
		}

		client, err := dialSSH(ctx, e.sshConfig, e.sshClientConfig)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to re-dial SSH client: %v", err)
		}
		e.sshClient = client

		// Retry NewSession after reconnect
		session, err = e.sshClient.NewSession()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create session after reconnect: %v", err)
		}
	}

	return session, nil
}

type remoteSession struct {
	ctx     context.Context
	cancel  context.CancelFunc
	session *ssh.Session
	isPTY   bool // Whether this is a PTY session
}

func (s *remoteSession) Wait() (exitCode int, reason commandv1beta1.CommandResult_Reason, err error) {
	// Watch for context cancellation and kill session if triggered
	done := make(chan struct{})
	go func() {
		select {
		case <-s.ctx.Done():
			_ = s.session.Signal(ssh.SIGKILL)
			_ = s.session.Close()
		case <-done:
			// Wait completed normally, stop watching
		}
	}()

	err = s.session.Wait()
	close(done) // Stop the watcher

	if err != nil {
		if s.ctx.Err() == context.DeadlineExceeded {
			return -1, commandv1beta1.CommandResult_REASON_TIMED_OUT, nil
		}

		switch e := err.(type) {
		case *ssh.ExitError:
			// Killed by signal (SIGINT, SIGTERM, SIGKILL, etc.)
			if sig := e.Signal(); sig != "" {
				return -1, commandv1beta1.CommandResult_REASON_KILLED, nil
			}

			// Normal exit with status
			if status := e.ExitStatus(); status >= 0 {
				return status, commandv1beta1.CommandResult_REASON_UNSPECIFIED, nil
			}

			// Defensive fallback
			return -1, commandv1beta1.CommandResult_REASON_KILLED, nil

		case *ssh.ExitMissingError:
			return -1, commandv1beta1.CommandResult_REASON_UNSPECIFIED, status.Errorf(
				codes.Internal,
				"remote server did not send exit-status or exit-signal",
			)

		default:
			return 0, 0, status.Errorf(
				codes.Internal,
				"failed to wait for session: %v",
				err,
			)
		}
	}

	return 0, commandv1beta1.CommandResult_REASON_COMPLETED, nil
}
func (s *remoteSession) Resize(rows, cols int) error {
	return s.session.WindowChange(rows, cols)
}

func (s *remoteSession) Signal(signal commandv1beta1.Signal) error {
	// For PTY sessions, signals don't reach the foreground process group,
	// so we just close the session instead.
	if s.isPTY {
		return s.Close()
	}

	var sig ssh.Signal

	switch signal {
	case commandv1beta1.Signal_SIGNAL_SIGINT:
		sig = ssh.SIGINT
	case commandv1beta1.Signal_SIGNAL_SIGTERM:
		sig = ssh.SIGTERM
	case commandv1beta1.Signal_SIGNAL_SIGKILL:
		sig = ssh.SIGKILL
	default:
		return status.Errorf(codes.InvalidArgument, "unknown signal: %v", signal)
	}

	return s.session.Signal(sig)
}

func (s *remoteSession) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	// Best-effort cleanup - session may already be closed by timeout watcher
	_ = s.session.Signal(ssh.SIGKILL)
	_ = s.session.Close()
	return nil
}

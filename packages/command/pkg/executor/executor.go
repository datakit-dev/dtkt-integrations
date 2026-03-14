package executor

import (
	"context"
	"io"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// CommandExecutor defines a standard interface for running commands locally or remotely.
type CommandExecutor interface {
	// ExecuteCommand executes a single non-interactive command safely (no shell features).
	ExecuteCommand(
		ctx context.Context,
		command *commandv1beta1.ExecutableCommand,
		cio *CommandIO,
	) (Session, error)

	// ExecuteShellCommand executes a single non-interactive shell command (allows pipes/redirection).
	ExecuteShellCommand(
		ctx context.Context,
		command *commandv1beta1.ShellCommand,
		shellCmd string,
		cio *CommandIO,
	) (Session, error)

	// TerminalSession starts a fully interactive session (PTY) for commands or shell.
	TerminalSession(
		ctx context.Context,
		startEvent *commandv1beta1.TerminalSessionRequest_StartEvent,
		shellCmd string,
		cio *CommandIO,
	) (Session, error)

	// Close terminates the command executor.
	Close() error
}

// CommandIO holds the standard input/output streams and other options for command execution.
type CommandIO struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Session interface {
	// Wait waits for the session to complete and returns the exit code.
	Wait() (exitCode int, reason commandv1beta1.CommandResult_Reason, err error)

	// Resize resizes the terminal (rows, cols follows POSIX convention).
	Resize(rows, cols int) error

	// Signal sends a signal to the session.
	Signal(signal commandv1beta1.Signal) error

	// Close ends the session.
	Close() error
}

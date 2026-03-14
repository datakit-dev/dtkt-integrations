package pkg

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/datakit-dev/dtkt-integrations/command/pkg/executor"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/util"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

type CommandService struct {
	commandv1beta1.UnimplementedCommandServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewCommandService(mux v1beta1.InstanceMux[*Instance]) *CommandService {
	return &CommandService{
		mux: mux,
	}
}

func (s *CommandService) ExecuteCommand(ctx context.Context, req *commandv1beta1.ExecuteCommandRequest) (*commandv1beta1.ExecuteCommandResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	cmd := req.GetCommand()
	_, exists := inst.commands[cmd.Command]
	if !exists {
		log.Error(ctx, "Command not found", slog.String("command", cmd.Command))
		return nil, status.Errorf(codes.NotFound, "command not found: %s", cmd.Command)
	}

	var b []byte
	if req.GetInput() != nil && req.GetInput().GetStdin() != nil {
		b = req.GetInput().GetStdin()
	}

	stdin := bytes.NewReader(b)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cio := &executor.CommandIO{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	log.Debug(ctx, "Starting command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args))

	start := time.Now()
	session, err := inst.exec.ExecuteCommand(ctx, cmd, cio)
	if err != nil {
		// For validation errors (InvalidArgument), return the error directly
		// For actual start failures, return REASON_FAILED_TO_START
		if status.Code(err) == codes.InvalidArgument {
			return nil, err
		}
		log.Warn(ctx, "Failed to start command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args), log.Err(err))
		return &commandv1beta1.ExecuteCommandResponse{
			Result: &commandv1beta1.CommandResult{
				ExitCode: -1,
				Reason:   commandv1beta1.CommandResult_REASON_FAILED_TO_START,
			},
		}, nil
	}

	log.Debug(ctx, "Waiting for command to finish")
	exitCode, reason, err := session.Wait()
	duration := time.Since(start)
	log.Debug(ctx, "Command finished", slog.Int("exit_code", exitCode), slog.String("reason", reason.String()))
	if err != nil {
		log.Error(ctx, "Command execution failed", log.Err(err))
		return nil, err
	}

	err = session.Close()
	if err != nil {
		log.Error(ctx, "Failed to close session", log.Err(err))
	}

	return &commandv1beta1.ExecuteCommandResponse{
		Result: &commandv1beta1.CommandResult{
			ExitCode: int32(exitCode),
			Reason:   reason,
			Duration: durationpb.New(duration),
		},
		Output: &commandv1beta1.CommandOutput{
			Stdout: stdout.Bytes(),
			Stderr: stderr.Bytes(),
		},
	}, nil
}

func (s *CommandService) ExecuteStreamedCommand(stream grpc.BidiStreamingServer[commandv1beta1.ExecuteStreamedCommandRequest, commandv1beta1.ExecuteStreamedCommandResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	var (
		start        time.Time
		duration     time.Duration
		exitCode     int
		reason       commandv1beta1.CommandResult_Reason
		session      executor.Session
		expectsStdin bool
	)

	var stdinReader *io.PipeReader
	var stdinWriter *io.PipeWriter
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var stdinCh chan []byte
	stdoutCh := make(chan []byte, numChunks)
	stderrCh := make(chan []byte, numChunks)

	stdinClosed := false
	closeStdinOnce := sync.Once{}
	closeStdin := func() {
		closeStdinOnce.Do(func() {
			stdinClosed = true
			if stdinCh != nil {
				close(stdinCh)
			}
		})
	}

	reqCh := make(chan *commandv1beta1.ExecuteStreamedCommandRequest, 1)
	errCh := make(chan error, 10)
	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})

	stdoutPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, chunkBytes)
			return &buf
		},
	}
	stderrPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, chunkBytes)
			return &buf
		},
	}

	wg := sync.WaitGroup{}
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "readPipeToCh"), slog.String("pipe", "stdout"))
		readPipeToCh(logger, stdoutReader, stdoutCh, stdoutPool, errCh)
	})
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "readPipeToCh"), slog.String("pipe", "stderr"))
		readPipeToCh(logger, stderrReader, stderrCh, stderrPool, errCh)
	})

	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleChannelRequests"))
		defer logger.Debug("Goroutine exiting")
		defer closeStdin()

		for {
			select {
			case <-doneCh:
				logger.Debug("Command finished, stopping request handling")
				return
			case <-cancelCh:
				logger.Debug("Command cancelled, stopping request handling")
				return
			case req := <-reqCh:
				if req.GetCommand() != nil {
					command := req.GetCommand()
					logger.Warn("Multiple command start requests received; ignoring", slog.String("command", command.Command), slog.Any("args", command.Args))
				}

				if req.GetInput() != nil {
					if session == nil {
						logger.Error("Command input received before command start")
						errCh <- status.Errorf(codes.FailedPrecondition, "command must be started before sending input")
						return
					}

					input := req.GetInput()
					if input == nil {
						logger.Error("Empty command input received")
						errCh <- status.Errorf(codes.InvalidArgument, "input is required")
						return
					}

					if input.GetStdin() != nil {
						if !expectsStdin {
							logger.Warn("Stdin data received but expects_stdin=false, ignoring", slog.Int("bytes", len(input.GetStdin())))
						} else if stdinClosed {
							logger.Error("Stdin data received after EOF")
							errCh <- status.Errorf(codes.FailedPrecondition, "stdin data received after EOF")
							return
						} else {
							logger.Debug("Received stdin input", slog.Int("bytes", len(input.GetStdin())))
							stdinCh <- input.GetStdin()
						}
					}

					if input.Eof {
						if !expectsStdin {
							logger.Warn("EOF received but expects_stdin=false, ignoring")
						} else if stdinClosed {
							logger.Error("Duplicate EOF received")
							errCh <- status.Errorf(codes.FailedPrecondition, "duplicate EOF received")
							return
						} else {
							logger.Debug("Received stdin eof")
							closeStdin()
						}
					}

					if input.GetSignal() != commandv1beta1.Signal_SIGNAL_UNSPECIFIED {
						logger.Debug("Received signal input", slog.String("signal", input.GetSignal().String()))
						if err := session.Signal(input.GetSignal()); err != nil {
							logger.Error("Failed to send signal", slog.String("signal", input.GetSignal().String()))
							errCh <- status.Errorf(codes.Internal, "failed to send signal: %v", err)
							return
						}
					}
				}
			}
		}
	})

	// Flush stdout/stderr and send output to response stream
	wg.Go(func() {
		handler := &OutputHandler{
			Logger:     log.FromCtx(ctx).With(slog.String("goroutine", "handleResponseOutput")),
			StdoutCh:   stdoutCh,
			StderrCh:   stderrCh,
			StdoutPool: stdoutPool,
			StderrPool: stderrPool,
			ErrCh:      errCh,
			Flush: func(stdout, stderr []byte) error {
				return stream.Send(&commandv1beta1.ExecuteStreamedCommandResponse{
					Output: &commandv1beta1.CommandOutput{
						Stdout: stdout,
						Stderr: stderr,
					},
				})
			},
		}
		handler.Run()
	})

	// Explicitly handle the start command
	req, err := stream.Recv()
	if err != nil {
		log.Error(ctx, "Failed to receive initial request", log.Err(err))
		return err
	}

	if req.GetCommand() != nil {
		cmd := req.GetCommand()
		_, exists := inst.commands[cmd.Command]
		if !exists {
			log.Error(ctx, "Command not found", slog.String("command", cmd.Command))
			return status.Errorf(codes.NotFound, "command not found: %s", cmd.Command)
		}

		expectsStdin = cmd.GetExpectsStdin()
		if expectsStdin {
			stdinReader, stdinWriter = io.Pipe()
			stdinCh = make(chan []byte, numChunks)
			wg.Go(func() {
				logger := log.FromCtx(ctx).With(slog.String("goroutine", "writeChToPipe"), slog.String("pipe", "stdin"))
				writeChToPipe(logger, stdinWriter, stdinCh, errCh)
			})
		}

		cio := &executor.CommandIO{
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
		}
		if stdinReader != nil {
			cio.Stdin = stdinReader
		}

		log.Debug(ctx, "Starting command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args), slog.Bool("expects_stdin", expectsStdin))

		start = time.Now()
		session, err = inst.exec.ExecuteCommand(ctx, cmd, cio)
		if err != nil {
			log.Error(ctx, "Failed to start command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args), log.Err(err))
			return err
		}

		// Send first message input through reqCh to be processed by handleChannelRequests
		if req.GetInput() != nil {
			reqCh <- &commandv1beta1.ExecuteStreamedCommandRequest{Input: req.GetInput()}
		}
	} else {
		log.Error(ctx, "Invalid request: no command to start")
		return status.Errorf(codes.InvalidArgument, "invalid request")
	}

	// The first request was handled (start command or invalid); handle new stream requests and push them to reqCh
	// This is not added to the wait group as it blocks waiting on stream.Recv which might only exit on stream close
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleStreamRequests"))
		logger.Debug("Starting handleStreamRequests")
		defer logger.Debug("Goroutine exiting")

		for {
			req, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Debug("Stream closed by client")
					return
				}

				logger.Error("Failed to receive request", log.Err(err))
				return
			}

			logger.Debug("Received request", slog.Any("req", req))
			reqCh <- req
		}
	}()

	// Wait for the command finish
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "waitForCommand"))

		logger.Debug("Waiting for command to finish")
		var err error
		exitCode, reason, err = session.Wait()
		duration = time.Since(start)
		logger.Debug("Command finished", slog.Int("exit_code", exitCode), slog.String("reason", reason.String()))

		if err != nil {
			logger.Error("Command execution failed", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "command execution failed: %v", err)
		}

		if err := stdoutWriter.Close(); err != nil {
			logger.Error("Failed to close stdout writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err)
		}

		if err := stderrWriter.Close(); err != nil {
			logger.Error("Failed to close stderr writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close stderr writer: %v", err)
		}

		close(doneCh)
	})

	// Handle context done
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleCtxDone"))
		logger.Debug("Starting handleCtxDone")

		<-ctx.Done()
		logger.Debug("Context Done")

		err = session.Close()
		if err != nil {
			logger.Error("Failed to close session", log.Err(err))
		}

		if err := stdoutWriter.Close(); err != nil {
			logger.Error("Failed to close stdout writer", log.Err(err))
			select {
			case errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err):
			default:
			}
		}

		if err := stderrWriter.Close(); err != nil {
			logger.Error("Failed to close stderr writer", log.Err(err))
			select {
			case errCh <- status.Errorf(codes.Internal, "failed to close stderr writer: %v", err):
			default:
			}
		}
	}()

	wgErr := sync.WaitGroup{}
	wgErr.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleErrors"))
		logger.Debug("Starting handleErrors")

		var collected []error
		cancelled := false

		for err := range errCh {
			logger.Error("Command error", log.Err(err))
			collected = append(collected, err)

			if !cancelled {
				cancelled = true
				close(cancelCh)
			}
		}

		if len(collected) > 0 {
			err = errors.Join(collected...)
		}
	})

	log.Debug(ctx, "Waiting for all goroutines to exit")
	wg.Wait()

	// Close errCh and wait for handleErrors goroutine to exit
	close(errCh)
	wgErr.Wait()

	if err != nil {
		log.Error(ctx, "Command execution failed", log.Err(err))
		return err
	}

	log.Debug(ctx, "All goroutines finished, sending final result", slog.Int("exit_code", exitCode), slog.String("duration", duration.String()))
	if err = stream.Send(&commandv1beta1.ExecuteStreamedCommandResponse{
		Result: &commandv1beta1.CommandResult{
			ExitCode: int32(exitCode),
			Reason:   reason,
			Duration: durationpb.New(duration),
		},
	}); err != nil {
		log.Error(ctx, "Failed to send final result", log.Err(err))
		return err
	}

	return nil
}

func (s *CommandService) ExecuteCommands(stream grpc.BidiStreamingServer[commandv1beta1.ExecuteCommandsRequest, commandv1beta1.ExecuteCommandsResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	for {
		// Check context before receiving
		select {
		case <-ctx.Done():
			log.Debug(ctx, "Context cancelled")
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug(ctx, "Stream closed by client")
				return nil
			}
			log.Error(ctx, "Failed to receive request", log.Err(err))
			return err
		}

		cmd := req.GetCommand()
		_, exists := inst.commands[cmd.Command]
		if !exists {
			log.Error(ctx, "Command not found", slog.String("command", cmd.Command))
			return status.Errorf(codes.NotFound, "command not found: %s", cmd.Command)
		}

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cio := &executor.CommandIO{
			Stdin:  bytes.NewReader(nil),
			Stdout: stdout,
			Stderr: stderr,
		}

		log.Debug(ctx, "Starting command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args))

		start := time.Now()
		session, err := inst.exec.ExecuteCommand(ctx, cmd, cio)
		if err != nil {
			log.Error(ctx, "Failed to start command", slog.String("command", cmd.Command), log.Err(err))
			return err
		}

		// Handle context cancellation while waiting
		waitDone := make(chan struct{})
		var exitCode int
		var waitErr error

		go func() {
			exitCode, _, waitErr = session.Wait()
			close(waitDone)
		}()

		select {
		case <-ctx.Done():
			log.Debug(ctx, "Context cancelled, closing session")
			//nolint:errcheck
			session.Close()
			return ctx.Err()
		case <-waitDone:
		}

		duration := time.Since(start)
		log.Debug(ctx, "Command finished", slog.Int("exit_code", exitCode))

		if waitErr != nil {
			log.Error(ctx, "Command execution failed", log.Err(waitErr))
			return waitErr
		}

		if err := session.Close(); err != nil {
			log.Error(ctx, "Failed to close session", log.Err(err))
		}

		if err := stream.Send(&commandv1beta1.ExecuteCommandsResponse{
			Result: &commandv1beta1.CommandResult{
				ExitCode: int32(exitCode),
				Duration: durationpb.New(duration),
			},
			Output: &commandv1beta1.CommandOutput{
				Stdout: stdout.Bytes(),
				Stderr: stderr.Bytes(),
			},
		}); err != nil {
			log.Error(ctx, "Failed to send response", log.Err(err))
			return err
		}
	}
}

func (s *CommandService) ExecuteBatchCommands(ctx context.Context, req *commandv1beta1.ExecuteBatchCommandsRequest) (*commandv1beta1.ExecuteBatchCommandsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	commands := req.GetCommands()
	if len(commands) == 0 {
		return &commandv1beta1.ExecuteBatchCommandsResponse{}, nil
	}

	// Validate all commands exist before executing
	for _, cmd := range commands {
		if _, exists := inst.commands[cmd.Command]; !exists {
			log.Error(ctx, "Command not found", slog.String("command", cmd.Command))
			return nil, status.Errorf(codes.NotFound, "command not found: %s", cmd.Command)
		}
	}

	results, errs := util.MapParallelErrs(commands, func(cmd *commandv1beta1.ExecutableCommand) (*commandv1beta1.BatchResult, error) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cio := &executor.CommandIO{
			Stdin:  bytes.NewReader(nil),
			Stdout: stdout,
			Stderr: stderr,
		}

		log.Debug(ctx, "Starting batch command", slog.String("command", cmd.Command), slog.Any("args", cmd.Args))

		start := time.Now()
		session, err := inst.exec.ExecuteCommand(ctx, cmd, cio)
		if err != nil {
			log.Error(ctx, "Failed to start batch command", slog.String("command", cmd.Command), log.Err(err))
			return nil, err
		}

		exitCode, _, err := session.Wait()
		duration := time.Since(start)
		log.Debug(ctx, "Batch command finished", slog.String("command", cmd.Command), slog.Int("exit_code", exitCode))

		if err != nil {
			log.Error(ctx, "Batch command execution failed", slog.String("command", cmd.Command), log.Err(err))
			return nil, err
		}

		if err := session.Close(); err != nil {
			log.Error(ctx, "Failed to close session", slog.String("command", cmd.Command), log.Err(err))
		}

		return &commandv1beta1.BatchResult{
			Output: &commandv1beta1.CommandOutput{
				Stdout: stdout.Bytes(),
				Stderr: stderr.Bytes(),
			},
			Result: &commandv1beta1.CommandResult{
				ExitCode: int32(exitCode),
				Duration: durationpb.New(duration),
			},
		}, nil
	})

	// Log any errors but still return results
	for idx, err := range errs {
		if err != nil {
			log.Error(ctx, "Batch command error", slog.Int("index", idx), log.Err(err))
		}
	}

	return &commandv1beta1.ExecuteBatchCommandsResponse{
		Results: results,
	}, nil
}

func (s *CommandService) ExecuteShellCommand(ctx context.Context, req *commandv1beta1.ExecuteShellCommandRequest) (*commandv1beta1.ExecuteShellCommandResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if !inst.config.GetAllowShell() {
		log.Error(ctx, "Shell command execution not allowed")
		return nil, status.Errorf(codes.PermissionDenied, "shell command execution not allowed")
	}

	cmd := req.GetCommand()

	stdin := bytes.NewReader(nil)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cio := &executor.CommandIO{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}

	log.Debug(ctx, "Starting shell command", slog.String("command", cmd.Command))

	start := time.Now()
	session, err := inst.exec.ExecuteShellCommand(ctx, cmd, inst.config.GetShellCommand(), cio)
	if err != nil {
		// For validation errors (InvalidArgument), return the error directly
		// For actual start failures, return REASON_FAILED_TO_START
		if status.Code(err) == codes.InvalidArgument {
			return nil, err
		}
		log.Warn(ctx, "Failed to start shell command", slog.String("command", cmd.Command), log.Err(err))
		return &commandv1beta1.ExecuteShellCommandResponse{
			Result: &commandv1beta1.CommandResult{
				ExitCode: -1,
				Reason:   commandv1beta1.CommandResult_REASON_FAILED_TO_START,
			},
		}, nil
	}

	log.Debug(ctx, "Waiting for shell command to finish")
	exitCode, reason, err := session.Wait()
	duration := time.Since(start)
	log.Debug(ctx, "Shell command finished", slog.Int("exit_code", exitCode), slog.String("reason", reason.String()))
	if err != nil {
		log.Error(ctx, "Shell command execution failed", log.Err(err))
		return nil, err
	}

	err = session.Close()
	if err != nil {
		log.Error(ctx, "Failed to close session", log.Err(err))
	}

	return &commandv1beta1.ExecuteShellCommandResponse{
		Result: &commandv1beta1.CommandResult{
			ExitCode: int32(exitCode),
			Reason:   reason,
			Duration: durationpb.New(duration),
		},
		Output: &commandv1beta1.CommandOutput{
			Stdout: stdout.Bytes(),
			Stderr: stderr.Bytes(),
		},
	}, nil
}

func (s *CommandService) ExecuteStreamedShellCommand(stream grpc.BidiStreamingServer[commandv1beta1.ExecuteStreamedShellCommandRequest, commandv1beta1.ExecuteStreamedShellCommandResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if !inst.config.GetAllowShell() {
		log.Error(ctx, "Shell command execution not allowed")
		return status.Errorf(codes.PermissionDenied, "shell command execution not allowed")
	}

	var (
		start        time.Time
		duration     time.Duration
		exitCode     int
		reason       commandv1beta1.CommandResult_Reason
		session      executor.Session
		expectsStdin bool
	)

	var stdinReader *io.PipeReader
	var stdinWriter *io.PipeWriter
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var stdinCh chan []byte
	stdoutCh := make(chan []byte, numChunks)
	stderrCh := make(chan []byte, numChunks)

	stdinClosed := false
	closeStdinOnce := sync.Once{}
	closeStdin := func() {
		closeStdinOnce.Do(func() {
			stdinClosed = true
			if stdinCh != nil {
				close(stdinCh)
			}
		})
	}

	reqCh := make(chan *commandv1beta1.ExecuteStreamedShellCommandRequest, 1)
	errCh := make(chan error, 10)
	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})

	stdoutPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, chunkBytes)
			return &buf
		},
	}
	stderrPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, chunkBytes)
			return &buf
		},
	}

	wg := sync.WaitGroup{}
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "readPipeToCh"), slog.String("pipe", "stdout"))
		readPipeToCh(logger, stdoutReader, stdoutCh, stdoutPool, errCh)
	})
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "readPipeToCh"), slog.String("pipe", "stderr"))
		readPipeToCh(logger, stderrReader, stderrCh, stderrPool, errCh)
	})

	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleChannelRequests"))
		defer logger.Debug("Goroutine exiting")
		defer closeStdin()

		for {
			select {
			case <-doneCh:
				logger.Debug("Command finished, stopping request handling")
				return
			case <-cancelCh:
				logger.Debug("Command cancelled, stopping request handling")
				return
			case req := <-reqCh:
				if req.GetCommand() != nil {
					command := req.GetCommand()
					logger.Warn("Multiple command start requests received; ignoring", slog.String("command", command.Command))
				}

				if req.GetInput() != nil {
					if session == nil {
						logger.Error("Command input received before command start")
						errCh <- status.Errorf(codes.FailedPrecondition, "command must be started before sending input")
						return
					}

					input := req.GetInput()
					if input == nil {
						logger.Error("Empty command input received")
						errCh <- status.Errorf(codes.InvalidArgument, "input is required")
						return
					}

					if input.GetStdin() != nil {
						if !expectsStdin {
							logger.Warn("Stdin data received but expects_stdin=false, ignoring", slog.Int("bytes", len(input.GetStdin())))
						} else if stdinClosed {
							logger.Error("Stdin data received after EOF")
							errCh <- status.Errorf(codes.FailedPrecondition, "stdin data received after EOF")
							return
						} else {
							logger.Debug("Received stdin input", slog.Int("bytes", len(input.GetStdin())))
							stdinCh <- input.GetStdin()
						}
					}

					if input.Eof {
						if !expectsStdin {
							logger.Warn("EOF received but expects_stdin=false, ignoring")
						} else if stdinClosed {
							logger.Error("Duplicate EOF received")
							errCh <- status.Errorf(codes.FailedPrecondition, "duplicate EOF received")
							return
						} else {
							logger.Debug("Received stdin eof")
							closeStdin()
						}
					}

					if input.GetSignal() != commandv1beta1.Signal_SIGNAL_UNSPECIFIED {
						logger.Debug("Received signal input", slog.String("signal", input.GetSignal().String()))
						if err := session.Signal(input.GetSignal()); err != nil {
							logger.Error("Failed to send signal", slog.String("signal", input.GetSignal().String()))
							errCh <- status.Errorf(codes.Internal, "failed to send signal: %v", err)
							return
						}
					}
				}
			}
		}
	})

	// Flush stdout/stderr and send output to response stream
	wg.Go(func() {
		handler := &OutputHandler{
			Logger:     log.FromCtx(ctx).With(slog.String("goroutine", "handleResponseOutput")),
			StdoutCh:   stdoutCh,
			StderrCh:   stderrCh,
			StdoutPool: stdoutPool,
			StderrPool: stderrPool,
			ErrCh:      errCh,
			Flush: func(stdout, stderr []byte) error {
				return stream.Send(&commandv1beta1.ExecuteStreamedShellCommandResponse{
					Output: &commandv1beta1.CommandOutput{
						Stdout: stdout,
						Stderr: stderr,
					},
				})
			},
		}
		handler.Run()
	})

	// Explicitly handle the start command
	req, err := stream.Recv()
	if err != nil {
		log.Error(ctx, "Failed to receive initial request", log.Err(err))
		return err
	}

	if req.GetCommand() != nil {
		cmd := req.GetCommand()

		expectsStdin = cmd.GetExpectsStdin()
		if expectsStdin {
			stdinReader, stdinWriter = io.Pipe()
			stdinCh = make(chan []byte, numChunks)
			wg.Go(func() {
				logger := log.FromCtx(ctx).With(slog.String("goroutine", "writeChToPipe"), slog.String("pipe", "stdin"))
				writeChToPipe(logger, stdinWriter, stdinCh, errCh)
			})
		}

		cio := &executor.CommandIO{
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
		}
		if stdinReader != nil {
			cio.Stdin = stdinReader
		}

		log.Debug(ctx, "Starting shell command", slog.String("command", cmd.Command), slog.Bool("expects_stdin", expectsStdin))

		start = time.Now()
		session, err = inst.exec.ExecuteShellCommand(ctx, cmd, inst.config.GetShellCommand(), cio)
		if err != nil {
			log.Error(ctx, "Failed to start shell command", slog.String("command", cmd.Command), log.Err(err))
			return err
		}

		// Send first message input through reqCh to be processed by handleChannelRequests
		if req.GetInput() != nil {
			reqCh <- &commandv1beta1.ExecuteStreamedShellCommandRequest{Input: req.GetInput()}
		}
	} else {
		log.Error(ctx, "Invalid request: no command to start")
		return status.Errorf(codes.InvalidArgument, "invalid request")
	}

	// The first request was handled (start command or invalid); handle new stream requests and push them to reqCh
	// This is not added to the wait group as it blocks waiting on stream.Recv which might only exit on stream close
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleStreamRequests"))
		logger.Debug("Starting handleStreamRequests")
		defer logger.Debug("Goroutine exiting")

		for {
			req, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Debug("Stream closed by client")
					return
				}

				logger.Error("Failed to receive request", log.Err(err))
				return
			}

			logger.Debug("Received request", slog.Any("req", req))
			reqCh <- req
		}
	}()

	// Wait for the command finish
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "waitForCommand"))

		logger.Debug("Waiting for shell command to finish")
		var err error
		exitCode, reason, err = session.Wait()
		duration = time.Since(start)
		logger.Debug("Shell command finished", slog.Int("exit_code", exitCode), slog.String("reason", reason.String()))

		if err != nil {
			logger.Error("Shell command execution failed", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "shell command execution failed: %v", err)
		}

		if err := stdoutWriter.Close(); err != nil {
			logger.Error("Failed to close stdout writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err)
		}

		if err := stderrWriter.Close(); err != nil {
			logger.Error("Failed to close stderr writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close stderr writer: %v", err)
		}

		close(doneCh)
	})

	// Handle context done
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleCtxDone"))
		logger.Debug("Starting handleCtxDone")

		<-ctx.Done()
		logger.Debug("Context Done")

		err = session.Close()
		if err != nil {
			logger.Error("Failed to close session", log.Err(err))
		}

		if err := stdoutWriter.Close(); err != nil {
			logger.Error("Failed to close stdout writer", log.Err(err))
			select {
			case errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err):
			default:
			}
		}

		if err := stderrWriter.Close(); err != nil {
			logger.Error("Failed to close stderr writer", log.Err(err))
			select {
			case errCh <- status.Errorf(codes.Internal, "failed to close stderr writer: %v", err):
			default:
			}
		}
	}()

	wgErr := sync.WaitGroup{}
	wgErr.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleErrors"))
		logger.Debug("Starting handleErrors")

		var collected []error
		cancelled := false

		for err := range errCh {
			logger.Error("Shell command error", log.Err(err))
			collected = append(collected, err)

			if !cancelled {
				cancelled = true
				close(cancelCh)
			}
		}

		if len(collected) > 0 {
			err = errors.Join(collected...)
		}
	})

	log.Debug(ctx, "Waiting for all goroutines to exit")
	wg.Wait()

	// Close errCh and wait for handleErrors goroutine to exit
	close(errCh)
	wgErr.Wait()

	if err != nil {
		log.Error(ctx, "Shell command execution failed", log.Err(err))
		return err
	}

	log.Debug(ctx, "All goroutines finished, sending final result", slog.Int("exit_code", exitCode), slog.String("duration", duration.String()))
	if err = stream.Send(&commandv1beta1.ExecuteStreamedShellCommandResponse{
		Result: &commandv1beta1.CommandResult{
			ExitCode: int32(exitCode),
			Reason:   reason,
			Duration: durationpb.New(duration),
		},
	}); err != nil {
		log.Error(ctx, "Failed to send final result", log.Err(err))
		return err
	}

	return nil
}

func (s *CommandService) ExecuteShellCommands(stream grpc.BidiStreamingServer[commandv1beta1.ExecuteShellCommandsRequest, commandv1beta1.ExecuteShellCommandsResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if !inst.config.GetAllowShell() {
		log.Error(ctx, "Shell command execution not allowed")
		return status.Errorf(codes.PermissionDenied, "shell command execution not allowed")
	}

	for {
		// Check context before receiving
		select {
		case <-ctx.Done():
			log.Debug(ctx, "Context cancelled")
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug(ctx, "Stream closed by client")
				return nil
			}
			log.Error(ctx, "Failed to receive request", log.Err(err))
			return err
		}

		cmd := req.GetCommand()

		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cio := &executor.CommandIO{
			Stdin:  bytes.NewReader(nil),
			Stdout: stdout,
			Stderr: stderr,
		}

		log.Debug(ctx, "Starting shell command", slog.String("command", cmd.Command))

		start := time.Now()
		session, err := inst.exec.ExecuteShellCommand(ctx, cmd, inst.config.GetShellCommand(), cio)
		if err != nil {
			log.Error(ctx, "Failed to start shell command", slog.String("command", cmd.Command), log.Err(err))
			return err
		}

		// Handle context cancellation while waiting
		waitDone := make(chan struct{})
		var exitCode int
		var waitErr error

		go func() {
			exitCode, _, waitErr = session.Wait()
			close(waitDone)
		}()

		select {
		case <-ctx.Done():
			log.Debug(ctx, "Context cancelled, closing session")
			return errors.Join(ctx.Err(), session.Close())
		case <-waitDone:
		}

		duration := time.Since(start)
		log.Debug(ctx, "Shell command finished", slog.Int("exit_code", exitCode))

		if waitErr != nil {
			log.Error(ctx, "Shell command execution failed", log.Err(waitErr))
			return waitErr
		}

		if err := session.Close(); err != nil {
			log.Error(ctx, "Failed to close session", log.Err(err))
		}

		if err := stream.Send(&commandv1beta1.ExecuteShellCommandsResponse{
			Result: &commandv1beta1.CommandResult{
				ExitCode: int32(exitCode),
				Duration: durationpb.New(duration),
			},
			Output: &commandv1beta1.CommandOutput{
				Stdout: stdout.Bytes(),
				Stderr: stderr.Bytes(),
			},
		}); err != nil {
			log.Error(ctx, "Failed to send response", log.Err(err))
			return err
		}
	}
}

func (s *CommandService) ExecuteBatchShellCommands(ctx context.Context, req *commandv1beta1.ExecuteBatchShellCommandsRequest) (*commandv1beta1.ExecuteBatchShellCommandsResponse, error) {
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}

	if !inst.config.GetAllowShell() {
		log.Error(ctx, "Shell command execution not allowed")
		return nil, status.Errorf(codes.PermissionDenied, "shell command execution not allowed")
	}

	commands := req.GetCommands()
	if len(commands) == 0 {
		return &commandv1beta1.ExecuteBatchShellCommandsResponse{}, nil
	}

	results, errs := util.MapParallelErrs(commands, func(cmd *commandv1beta1.ShellCommand) (*commandv1beta1.BatchResult, error) {
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}

		cio := &executor.CommandIO{
			Stdin:  bytes.NewReader(nil),
			Stdout: stdout,
			Stderr: stderr,
		}

		log.Debug(ctx, "Starting batch shell command", slog.String("command", cmd.Command))

		start := time.Now()
		session, err := inst.exec.ExecuteShellCommand(ctx, cmd, inst.config.GetShellCommand(), cio)
		if err != nil {
			log.Error(ctx, "Failed to start batch shell command", slog.String("command", cmd.Command), log.Err(err))
			return nil, err
		}

		exitCode, _, err := session.Wait()
		duration := time.Since(start)
		log.Debug(ctx, "Batch shell command finished", slog.String("command", cmd.Command), slog.Int("exit_code", exitCode))

		if err != nil {
			log.Error(ctx, "Batch shell command execution failed", slog.String("command", cmd.Command), log.Err(err))
			return nil, err
		}

		if err := session.Close(); err != nil {
			log.Error(ctx, "Failed to close session", slog.String("command", cmd.Command), log.Err(err))
		}

		return &commandv1beta1.BatchResult{
			Output: &commandv1beta1.CommandOutput{
				Stdout: stdout.Bytes(),
				Stderr: stderr.Bytes(),
			},
			Result: &commandv1beta1.CommandResult{
				ExitCode: int32(exitCode),
				Duration: durationpb.New(duration),
			},
		}, nil
	})

	// Log any errors but still return results
	for idx, err := range errs {
		if err != nil {
			log.Error(ctx, "Batch shell command error", slog.Int("index", idx), log.Err(err))
		}
	}

	return &commandv1beta1.ExecuteBatchShellCommandsResponse{
		Results: results,
	}, nil
}

func (s *CommandService) TerminalSession(stream grpc.BidiStreamingServer[commandv1beta1.TerminalSessionRequest, commandv1beta1.TerminalSessionResponse]) error {
	ctx := stream.Context()
	inst, err := s.mux.GetInstance(ctx)
	if err != nil {
		return status.Error(codes.FailedPrecondition, err.Error())
	}

	if !inst.config.GetAllowShell() {
		log.Error(ctx, "Terminal session not allowed (requires shell)")
		return status.Errorf(codes.PermissionDenied, "terminal session not allowed")
	}

	var session executor.Session

	var stdinReader *io.PipeReader
	var stdinWriter *io.PipeWriter
	stdoutReader, stdoutWriter := io.Pipe()

	var stdinCh chan []byte
	stdoutCh := make(chan []byte, numChunks)

	stdinClosed := false
	closeStdinOnce := sync.Once{}
	closeStdin := func() {
		closeStdinOnce.Do(func() {
			stdinClosed = true
			if stdinCh != nil {
				close(stdinCh)
			}
		})
	}

	reqCh := make(chan *commandv1beta1.TerminalSessionRequest, 1)
	errCh := make(chan error, 10)
	cancelCh := make(chan struct{})
	doneCh := make(chan struct{})

	stdoutPool := &sync.Pool{
		New: func() any {
			buf := make([]byte, chunkBytes)
			return &buf
		},
	}

	// Buffer for accumulating PTY data including any remainder from previous reads
	ptyBuf := make([]byte, chunkBytes)
	ptyPendingSize := 0

	wg := sync.WaitGroup{}
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "readPtyToCh"), slog.String("pipe", "stdout"))
		readPtyToCh(logger, stdoutReader, stdoutCh, stdoutPool, errCh, ptyBuf, &ptyPendingSize)
	})

	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleChannelRequests"))
		defer logger.Debug("Goroutine exiting")
		defer closeStdin()

		for {
			select {
			case <-doneCh:
				logger.Debug("Terminal session finished, stopping request handling")
				return
			case <-cancelCh:
				logger.Debug("Terminal session cancelled, stopping request handling")
				return
			case req := <-reqCh:
				if req.GetStart() != nil {
					logger.Warn("Multiple start requests received, ignoring")
				}

				if req.GetResize() != nil {
					resize := req.GetResize()
					dims := resize.GetDimensions()
					if dims != nil && session != nil {
						logger.Debug("Resizing terminal", slog.Int("rows", int(dims.Rows)), slog.Int("cols", int(dims.Cols)))
						if err := session.Resize(int(dims.Rows), int(dims.Cols)); err != nil {
							logger.Error("Failed to resize terminal", log.Err(err))
						}
					}
				}

				if req.GetInput() != nil {
					if session == nil {
						logger.Error("Input received before terminal session start")
						errCh <- status.Errorf(codes.FailedPrecondition, "terminal session must be started before sending input")
						return
					}

					input := req.GetInput()

					if input.GetStdin() != nil {
						if stdinClosed {
							logger.Error("Stdin data received after EOF")
							errCh <- status.Errorf(codes.FailedPrecondition, "stdin data received after EOF")
							return
						}
						logger.Debug("Received stdin input", slog.Int("bytes", len(input.GetStdin())))
						stdinCh <- input.GetStdin()
					}

					if input.Eof {
						if stdinClosed {
							logger.Error("Duplicate EOF received")
							errCh <- status.Errorf(codes.FailedPrecondition, "duplicate EOF received")
							return
						}
						logger.Debug("Received stdin eof")
						closeStdin()
					}

					if input.GetSignal() != commandv1beta1.Signal_SIGNAL_UNSPECIFIED {
						logger.Debug("Received signal input", slog.String("signal", input.GetSignal().String()))
						if err := session.Signal(input.GetSignal()); err != nil {
							logger.Error("Failed to send signal", slog.String("signal", input.GetSignal().String()))
							errCh <- status.Errorf(codes.Internal, "failed to send signal: %v", err)
							return
						}
					}
				}
			}
		}
	})

	// Flush stdout and send output to response stream
	wg.Go(func() {
		handler := &OutputHandler{
			Logger:     log.FromCtx(ctx).With(slog.String("goroutine", "handleResponseOutput")),
			StdoutCh:   stdoutCh,
			StderrCh:   nil, // PTY merges stdout/stderr
			StdoutPool: stdoutPool,
			StderrPool: nil,
			ErrCh:      errCh,
			Flush: func(stdout, _ []byte) error {
				return stream.Send(&commandv1beta1.TerminalSessionResponse{
					Output: &commandv1beta1.CommandOutput{
						Stdout: stdout,
					},
				})
			},
		}
		handler.Run()
	})

	// Explicitly handle the start request
	req, err := stream.Recv()
	if err != nil {
		log.Error(ctx, "Failed to receive initial request", log.Err(err))
		return err
	}

	if req.GetStart() != nil {
		startEvent := req.GetStart()

		// Terminal sessions always expect stdin
		stdinReader, stdinWriter = io.Pipe()
		stdinCh = make(chan []byte, numChunks)
		wg.Go(func() {
			logger := log.FromCtx(ctx).With(slog.String("goroutine", "writeChToPipe"), slog.String("pipe", "stdin"))
			writeChToPipe(logger, stdinWriter, stdinCh, errCh)
		})

		cio := &executor.CommandIO{
			Stdin:  stdinReader,
			Stdout: stdoutWriter,
			Stderr: stdoutWriter, // PTY merges stdout/stderr
		}

		log.Debug(ctx, "Starting terminal session",
			slog.String("workdir", startEvent.GetWorkdir()),
			slog.Int("cols", int(startEvent.GetDimensions().GetCols())),
			slog.Int("rows", int(startEvent.GetDimensions().GetRows())))

		session, err = inst.exec.TerminalSession(ctx, startEvent, inst.config.GetShellCommand(), cio)
		if err != nil {
			log.Error(ctx, "Failed to start terminal session", log.Err(err))
			return err
		}

		// Send first message input through reqCh to be processed by handleChannelRequests
		if req.GetInput() != nil {
			reqCh <- &commandv1beta1.TerminalSessionRequest{Input: req.GetInput()}
		}
	} else {
		log.Error(ctx, "Invalid request: no start event to initialize terminal session")
		return status.Errorf(codes.InvalidArgument, "invalid request: start event required")
	}

	// Handle new stream requests and push them to reqCh
	// This is not added to the wait group as it blocks waiting on stream.Recv which might only exit on stream close
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleStreamRequests"))
		logger.Debug("Starting handleStreamRequests")
		defer logger.Debug("Goroutine exiting")

		for {
			req, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					logger.Debug("Stream closed by client")
					return
				}

				logger.Error("Failed to receive request", log.Err(err))
				return
			}

			logger.Debug("Received request", slog.Any("req", req))
			reqCh <- req
		}
	}()

	// Wait for the terminal session to finish
	wg.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "waitForSession"))

		logger.Debug("Waiting for terminal session to finish")
		exitCode, _, err := session.Wait()
		logger.Debug("Terminal session finished", slog.Int("exit_code", exitCode))

		if err != nil {
			logger.Error("Terminal session failed", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "terminal session failed: %v", err)
		}

		if err := stdoutWriter.Close(); err != nil {
			logger.Error("Failed to close stdout writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err)
		}

		close(doneCh)
	})

	// Handle context done - only close session if context is cancelled before session finishes
	go func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleCtxDone"))
		logger.Debug("Starting handleCtxDone")

		select {
		case <-ctx.Done():
			logger.Debug("Context Done")

			err = session.Close()
			if err != nil {
				logger.Error("Failed to close session", log.Err(err))
			}

			if err := stdoutWriter.Close(); err != nil {
				logger.Error("Failed to close stdout writer", log.Err(err))
				select {
				case errCh <- status.Errorf(codes.Internal, "failed to close stdout writer: %v", err):
				default:
				}
			}
		case <-doneCh:
			// Session finished gracefully, no need to force close
			logger.Debug("Session finished before context done")
		}
	}()

	wgErr := sync.WaitGroup{}
	wgErr.Go(func() {
		logger := log.FromCtx(ctx).With(slog.String("goroutine", "handleErrors"))
		logger.Debug("Starting handleErrors")

		var collected []error
		cancelled := false

		for err := range errCh {
			logger.Error("Terminal session error", log.Err(err))
			collected = append(collected, err)

			if !cancelled {
				cancelled = true
				close(cancelCh)
			}
		}

		if len(collected) > 0 {
			err = errors.Join(collected...)
		}
	})

	log.Debug(ctx, "Waiting for all goroutines to exit")
	wg.Wait()

	// Close errCh and wait for handleErrors goroutine to exit
	close(errCh)
	wgErr.Wait()

	if err != nil {
		log.Error(ctx, "Terminal session failed", log.Err(err))
		return err
	}

	log.Debug(ctx, "Terminal session completed")
	return nil
}

package pkg

import (
	"errors"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/charmbracelet/x/ansi/parser"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/lib/log"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const chunkBytes = 32 * 1024 // 32 KiB
const numChunks = 8          // 8 x 32 KiB chunks; 256 KiB total

const flushThresholdBytes = 1024 // 1 KiB

const coalesceDelay = 10 * time.Millisecond // delay for coalescing output

// findTerminalSafeBoundary scans the buffer and returns the number of bytes
// that can be safely flushed without splitting a terminal sequence or UTF-8 rune.
//
// Uses charmbracelet/x/ansi/parser to track state transitions. A boundary is safe
// when the parser returns to GroundState with a non-None action (meaning a complete
// unit was processed: PrintAction for ASCII, ExecuteAction for UTF-8/control,
// DispatchAction for escape sequences).
//
// Returns the index of the last safe boundary (may be 0 if entire buffer is
// part of an incomplete sequence).
func findTerminalSafeBoundary(buf []byte) int {
	if len(buf) == 0 {
		return 0
	}

	state := parser.GroundState
	lastSafeBoundary := 0

	for i := 0; i < len(buf); i++ {
		var action parser.Action
		state, action = parser.Table.Transition(state, buf[i])

		// Safe boundary when:
		// 1. Parser is in GroundState (not mid-sequence)
		// 2. Action is not NoneAction (something was actually completed)
		if state == parser.GroundState && action != parser.NoneAction {
			lastSafeBoundary = i + 1
		}
	}

	return lastSafeBoundary
}

// Drain the ch into the buffer and return the chunks to the pool.
// An initial chunk has already been read from the channel.
func drainChannel(ch <-chan []byte, buf *[]byte, chunk []byte, pool *sync.Pool) {
	*buf = append(*buf, chunk...)
	pool.Put(&chunk)

	for {
		select {
		case chunk, ok := <-ch:
			if !ok {
				return
			}
			*buf = append(*buf, chunk...)
			pool.Put(&chunk)
		default:
			return
		}
	}
}

// Read from a pipe reader and send output to a channel. Terminates when the pipe has closed upstream or an error occurs.
func readPipeToCh(logger *slog.Logger, r *io.PipeReader, outputCh chan<- []byte, pool *sync.Pool, errCh chan<- error) {
	logger.Debug("Starting readPipeToCh")
	defer logger.Debug("Goroutine exiting")
	defer close(outputCh)

	for {
		// logger.Debug("Reading from pipe")

		tmp := pool.Get().(*[]byte)
		n, err := r.Read(*tmp)
		if n > 0 {
			chunk := (*tmp)[:n] // slice of correct length
			outputCh <- chunk   // send pointer to the channel
		} else {
			// no data read, return buffer
			pool.Put(tmp)
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug("Pipe closed")
				return
			}
			logger.Error("Failed to read from pipe", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to read from pipe: %v", err)
			return
		}

		// logger.Debug("Read from pipe", slog.Int("bytes", n))
	}
}

// readPtyToCh reads from a PTY pipe and sends terminal-safe chunks to a channel.
//
// TerminalSession streams PTY output, which is a terminal protocol (not plain text).
// Message boundaries must never split UTF-8 runes or ANSI escape sequences.
// This function buffers output and only emits chunks at terminal-safe boundaries.
//
// Terminates when the pipe has closed upstream or an error occurs.
func readPtyToCh(logger *slog.Logger, r *io.PipeReader, outputCh chan<- []byte, pool *sync.Pool, errCh chan<- error, ptyBuf []byte, ptyPendingSize *int) {
	logger.Debug("Starting readPtyToCh")
	defer logger.Debug("Goroutine exiting")
	defer close(outputCh)

	for {
		n, err := r.Read(ptyBuf[*ptyPendingSize:])
		if n > 0 {
			total := *ptyPendingSize + n
			data := ptyBuf[:total]

			lastSafe := findTerminalSafeBoundary(data)
			if lastSafe > 0 {
				tmp := pool.Get().(*[]byte)
				chunk := (*tmp)[:lastSafe]
				copy(chunk, data[:lastSafe])

				outputCh <- chunk
			}

			remainder := total - lastSafe
			copy(ptyBuf, data[lastSafe:total])
			*ptyPendingSize = remainder
		}

		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug("Pipe closed")
				return
			}
			logger.Error("Failed to read from pipe", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to read from pipe: %v", err)
			return
		}
	}
}

// Write from inputCh to a pipe writer; terminates when inputCh is closed or an error occurs.
func writeChToPipe(logger *slog.Logger, w *io.PipeWriter, inputCh <-chan []byte, errCh chan<- error) {
	logger.Debug("Starting writeChToPipe")
	defer logger.Debug("Goroutine exiting")

	defer func() {
		if err := w.Close(); err != nil {
			logger.Error("Failed to close pipe writer", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to close pipe writer: %v", err)
		}
	}()

	for chunk := range inputCh {
		logger.Debug("Writing to pipe", slog.Int("bytes", len(chunk)))
		if _, err := w.Write(chunk); err != nil {
			logger.Error("Failed to write to pipe", log.Err(err))
			errCh <- status.Errorf(codes.Internal, "failed to write to pipe: %v", err)
			return
		}
	}
}

// OutputHandler handles coalescing and flushing of stdout/stderr output.
// It buffers small chunks and sends them together after a delay or threshold.
type OutputHandler struct {
	Logger     *slog.Logger
	StdoutCh   <-chan []byte
	StderrCh   <-chan []byte // nil for terminal sessions (PTY merges stdout/stderr)
	StdoutPool *sync.Pool
	StderrPool *sync.Pool // nil for terminal sessions
	Flush      func(stdout, stderr []byte) error
	ErrCh      chan<- error
}

// Run processes output from channels and flushes to the stream.
// It coalesces small chunks and sends them together after a delay or threshold.
func (h *OutputHandler) Run() {
	defer h.Logger.Debug("Goroutine exiting")

	stdoutBuf := make([]byte, 0, chunkBytes*numChunks)
	stderrBuf := make([]byte, 0, chunkBytes*numChunks)

	var timer *time.Timer
	var timerC <-chan time.Time

	startOrResetTimer := func() {
		if timer == nil {
			timer = time.NewTimer(coalesceDelay)
		} else {
			timer.Reset(coalesceDelay)
		}
		timerC = timer.C
	}

	stopTimer := func() {
		if timer != nil {
			timer.Stop()
			timer = nil
			timerC = nil
		}
	}

	flush := func() bool {
		if len(stdoutBuf) == 0 && len(stderrBuf) == 0 {
			return true
		}

		h.Logger.Debug("Flushing output", slog.Int("stdout_bytes", len(stdoutBuf)), slog.Int("stderr_bytes", len(stderrBuf)))
		if err := h.Flush(stdoutBuf, stderrBuf); err != nil {
			h.Logger.Error("Failed to send output", log.Err(err))
			h.ErrCh <- err
			return false
		}

		stdoutBuf = stdoutBuf[:0]
		stderrBuf = stderrBuf[:0]
		return true
	}

	stdoutCh := h.StdoutCh
	stderrCh := h.StderrCh

	for stdoutCh != nil || stderrCh != nil {
		h.Logger.Debug("Waiting for output or timer")
		timerFired := false

		select {
		case chunk, ok := <-stdoutCh:
			if !ok {
				stdoutCh = nil
				continue
			}
			// h.Logger.Debug("Received stdout chunk", slog.Int("bytes", len(chunk)))
			drainChannel(stdoutCh, &stdoutBuf, chunk, h.StdoutPool)

		case chunk, ok := <-stderrCh:
			if !ok {
				stderrCh = nil
				continue
			}
			// h.Logger.Debug("Received stderr chunk", slog.Int("bytes", len(chunk)))
			drainChannel(stderrCh, &stderrBuf, chunk, h.StderrPool)

		case <-timerC:
			h.Logger.Debug("Coalesce timer fired")
			timerFired = true
		}

		if len(stdoutBuf) >= flushThresholdBytes || len(stderrBuf) >= flushThresholdBytes || timerFired {
			if ok := flush(); !ok {
				return
			}
			stopTimer()
		} else if len(stdoutBuf)+len(stderrBuf) > 0 {
			h.Logger.Debug("Starting or resetting coalesce timer", slog.Int("stdout_bytes", len(stdoutBuf)), slog.Int("stderr_bytes", len(stderrBuf)))
			startOrResetTimer()
		}
	}

	// Final flush
	flush()
}

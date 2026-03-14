package test

import (
	"context"
	"io"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// Compile-time interface assertions
var (
	_ grpc.ServerStreamingServer[any] = (*MockServerStreamServer[any])(nil)
	_ grpc.ServerStreamingClient[any] = (*MockServerStreamClient[any])(nil)
)

// mockServerStreamState holds the shared state between client and server views.
type mockServerStreamState[Res any] struct {
	ctx     context.Context
	cancel  context.CancelFunc
	resCh   chan *Res // server sends responses, client receives
	mu      sync.Mutex
	closed  bool
	header  metadata.MD
	trailer metadata.MD
}

// MockServerStreamClient is the client-side view of a mock server stream.
// It implements grpc.ServerStreamingClient and provides control methods for testing.
type MockServerStreamClient[Res any] struct {
	state   *mockServerStreamState[Res]
	recvErr error
}

// MockServerStreamServer is the server-side view of a mock server stream.
// It implements grpc.ServerStreamingServer and provides control methods for testing.
type MockServerStreamServer[Res any] struct {
	state   *mockServerStreamState[Res]
	sendErr error
}

// NewMockServerStream creates a new mock server stream and returns both views.
func NewMockServerStream[Res any](ctx context.Context) (*MockServerStreamClient[Res], *MockServerStreamServer[Res]) {
	ctx, cancel := context.WithCancel(ctx)
	state := &mockServerStreamState[Res]{
		ctx:     ctx,
		cancel:  cancel,
		resCh:   make(chan *Res, 100),
		header:  metadata.MD{},
		trailer: metadata.MD{},
	}
	client := &MockServerStreamClient[Res]{state: state}
	server := &MockServerStreamServer[Res]{state: state}
	return client, server
}

// Client-side implementation

// Recv implements grpc.ServerStreamingClient.Recv
func (c *MockServerStreamClient[Res]) Recv() (*Res, error) {
	c.state.mu.Lock()
	err := c.recvErr
	c.state.mu.Unlock()

	if err != nil {
		return nil, err
	}

	select {
	case msg, ok := <-c.state.resCh:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-c.state.ctx.Done():
		return nil, c.state.ctx.Err()
	}
}

// Header implements grpc.ServerStreamingClient.Header
func (c *MockServerStreamClient[Res]) Header() (metadata.MD, error) {
	return c.state.header, nil
}

// Trailer implements grpc.ServerStreamingClient.Trailer
func (c *MockServerStreamClient[Res]) Trailer() metadata.MD {
	return c.state.trailer
}

// CloseSend implements grpc.ServerStreamingClient.CloseSend
func (c *MockServerStreamClient[Res]) CloseSend() error {
	// No-op for server streaming - client doesn't send
	return nil
}

// Context implements grpc.ServerStreamingClient.Context
func (c *MockServerStreamClient[Res]) Context() context.Context {
	return c.state.ctx
}

// SendMsg implements grpc.ClientStream.SendMsg
func (c *MockServerStreamClient[Res]) SendMsg(msg any) error {
	// No-op for server streaming - client doesn't send
	return nil
}

// RecvMsg implements grpc.ClientStream.RecvMsg
func (c *MockServerStreamClient[Res]) RecvMsg(msg any) error {
	received, err := c.Recv()
	if err != nil {
		return err
	}
	*msg.(*Res) = *received
	return nil
}

// Control methods for MockServerStreamClient

// SetRecvError injects an error for the next Recv call.
func (c *MockServerStreamClient[Res]) SetRecvError(err error) {
	c.state.mu.Lock()
	c.recvErr = err
	c.state.mu.Unlock()
}

// DrainResponses returns all buffered responses without blocking.
func (c *MockServerStreamClient[Res]) DrainResponses() []*Res {
	var responses []*Res
	for {
		select {
		case resp := <-c.state.resCh:
			responses = append(responses, resp)
		default:
			return responses
		}
	}
}

// Close simulates the client connection being dropped or the RPC being cancelled.
func (c *MockServerStreamClient[Res]) Close() {
	c.state.mu.Lock()
	c.state.closed = true
	c.state.mu.Unlock()
	c.state.cancel()
}

// Server-side implementation

// Send implements grpc.ServerStreamingServer.Send
func (s *MockServerStreamServer[Res]) Send(msg *Res) error {
	s.state.mu.Lock()
	err := s.sendErr
	closed := s.state.closed
	s.state.mu.Unlock()

	if err != nil {
		return err
	}
	if closed {
		return io.EOF
	}

	// Deep copy the message to simulate wire serialization.
	msgCopy := proto.Clone(any(msg).(proto.Message))

	select {
	case s.state.resCh <- any(msgCopy).(*Res):
		return nil
	case <-s.state.ctx.Done():
		return s.state.ctx.Err()
	}
}

// Context implements grpc.ServerStreamingServer.Context
func (s *MockServerStreamServer[Res]) Context() context.Context {
	return s.state.ctx
}

// SetHeader implements grpc.ServerStreamingServer.SetHeader
func (s *MockServerStreamServer[Res]) SetHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SendHeader implements grpc.ServerStreamingServer.SendHeader
func (s *MockServerStreamServer[Res]) SendHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SetTrailer implements grpc.ServerStreamingServer.SetTrailer
func (s *MockServerStreamServer[Res]) SetTrailer(md metadata.MD) {
	s.state.trailer = metadata.Join(s.state.trailer, md)
}

// SendMsg implements grpc.ServerStreamingServer.SendMsg
func (s *MockServerStreamServer[Res]) SendMsg(msg any) error {
	return s.Send(msg.(*Res))
}

// RecvMsg implements grpc.ServerStreamingServer.RecvMsg
func (s *MockServerStreamServer[Res]) RecvMsg(msg any) error {
	// No-op for server streaming - server doesn't receive
	return io.EOF
}

// Control methods for MockServerStreamServer

// SetSendError injects an error for the next Send call.
func (s *MockServerStreamServer[Res]) SetSendError(err error) {
	s.state.mu.Lock()
	s.sendErr = err
	s.state.mu.Unlock()
}

// Header returns the header metadata.
func (s *MockServerStreamServer[Res]) Header() metadata.MD {
	return s.state.header
}

// Trailer returns the trailer metadata.
func (s *MockServerStreamServer[Res]) Trailer() metadata.MD {
	return s.state.trailer
}

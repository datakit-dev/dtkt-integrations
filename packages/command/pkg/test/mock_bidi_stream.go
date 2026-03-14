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
	_ grpc.BidiStreamingServer[any, any] = (*MockBidiServer[any, any])(nil)
	_ grpc.BidiStreamingClient[any, any] = (*MockBidiClient[any, any])(nil)
)

// mockBidiState holds the shared state between client and server views.
type mockBidiState[Req, Res any] struct {
	ctx     context.Context
	cancel  context.CancelFunc
	reqCh   chan *Req // client sends requests, server receives
	resCh   chan *Res // server sends responses, client receives
	mu      sync.Mutex
	closed  bool
	header  metadata.MD
	trailer metadata.MD
}

// MockBidiClient is the client-side view of a mock bidirectional stream.
// It implements grpc.BidiStreamingClient and provides control methods for testing.
type MockBidiClient[Req, Res any] struct {
	state   *mockBidiState[Req, Res]
	sendErr error
	recvErr error
}

// MockBidiServer is the server-side view of a mock bidirectional stream.
// It implements grpc.BidiStreamingServer and provides control methods for testing.
type MockBidiServer[Req, Res any] struct {
	state   *mockBidiState[Req, Res]
	sendErr error
	recvErr error
}

// NewMockBidiStream creates a new mock bidirectional stream and returns both views.
func NewMockBidiStream[Req, Res any](ctx context.Context) (*MockBidiClient[Req, Res], *MockBidiServer[Req, Res]) {
	ctx, cancel := context.WithCancel(ctx)
	state := &mockBidiState[Req, Res]{
		ctx:     ctx,
		cancel:  cancel,
		reqCh:   make(chan *Req, 100),
		resCh:   make(chan *Res, 100),
		header:  metadata.MD{},
		trailer: metadata.MD{},
	}
	client := &MockBidiClient[Req, Res]{state: state}
	server := &MockBidiServer[Req, Res]{state: state}
	return client, server
}

// Client-side implementation

// Send implements grpc.BidiStreamingClient.Send
func (c *MockBidiClient[Req, Res]) Send(msg *Req) error {
	c.state.mu.Lock()
	err := c.sendErr
	closed := c.state.closed
	c.state.mu.Unlock()

	if err != nil {
		return err
	}
	if closed {
		return io.EOF
	}

	select {
	case c.state.reqCh <- msg:
		return nil
	case <-c.state.ctx.Done():
		return c.state.ctx.Err()
	}
}

// Recv implements grpc.BidiStreamingClient.Recv
func (c *MockBidiClient[Req, Res]) Recv() (*Res, error) {
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

// CloseSend implements grpc.BidiStreamingClient.CloseSend
func (c *MockBidiClient[Req, Res]) CloseSend() error {
	close(c.state.reqCh)
	return nil
}

// Header implements grpc.BidiStreamingClient.Header
func (c *MockBidiClient[Req, Res]) Header() (metadata.MD, error) {
	return c.state.header, nil
}

// Trailer implements grpc.BidiStreamingClient.Trailer
func (c *MockBidiClient[Req, Res]) Trailer() metadata.MD {
	return c.state.trailer
}

// Context implements grpc.BidiStreamingClient.Context
func (c *MockBidiClient[Req, Res]) Context() context.Context {
	return c.state.ctx
}

// SendMsg implements grpc.ClientStream.SendMsg
func (c *MockBidiClient[Req, Res]) SendMsg(msg any) error {
	return c.Send(msg.(*Req))
}

// RecvMsg implements grpc.ClientStream.RecvMsg
func (c *MockBidiClient[Req, Res]) RecvMsg(msg any) error {
	received, err := c.Recv()
	if err != nil {
		return err
	}
	*msg.(*Res) = *received
	return nil
}

// Control methods for MockBidiClient

// SetSendError injects an error for the next Send call.
func (c *MockBidiClient[Req, Res]) SetSendError(err error) {
	c.state.mu.Lock()
	c.sendErr = err
	c.state.mu.Unlock()
}

// SetRecvError injects an error for the next Recv call.
func (c *MockBidiClient[Req, Res]) SetRecvError(err error) {
	c.state.mu.Lock()
	c.recvErr = err
	c.state.mu.Unlock()
}

// DrainResponses returns all buffered responses without blocking.
func (c *MockBidiClient[Req, Res]) DrainResponses() []*Res {
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
func (c *MockBidiClient[Req, Res]) Close() {
	c.state.mu.Lock()
	c.state.closed = true
	c.state.mu.Unlock()
	c.state.cancel()
}

// Server-side implementation

// Send implements grpc.BidiStreamingServer.Send
func (s *MockBidiServer[Req, Res]) Send(msg *Res) error {
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

// Recv implements grpc.BidiStreamingServer.Recv
func (s *MockBidiServer[Req, Res]) Recv() (*Req, error) {
	s.state.mu.Lock()
	err := s.recvErr
	s.state.mu.Unlock()

	if err != nil {
		return nil, err
	}

	select {
	case msg, ok := <-s.state.reqCh:
		if !ok {
			return nil, io.EOF
		}
		return msg, nil
	case <-s.state.ctx.Done():
		return nil, s.state.ctx.Err()
	}
}

// Context implements grpc.BidiStreamingServer.Context
func (s *MockBidiServer[Req, Res]) Context() context.Context {
	return s.state.ctx
}

// SetHeader implements grpc.BidiStreamingServer.SetHeader
func (s *MockBidiServer[Req, Res]) SetHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SendHeader implements grpc.BidiStreamingServer.SendHeader
func (s *MockBidiServer[Req, Res]) SendHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SetTrailer implements grpc.BidiStreamingServer.SetTrailer
func (s *MockBidiServer[Req, Res]) SetTrailer(md metadata.MD) {
	s.state.trailer = metadata.Join(s.state.trailer, md)
}

// SendMsg implements grpc.BidiStreamingServer.SendMsg
func (s *MockBidiServer[Req, Res]) SendMsg(msg any) error {
	return s.Send(msg.(*Res))
}

// RecvMsg implements grpc.BidiStreamingServer.RecvMsg
func (s *MockBidiServer[Req, Res]) RecvMsg(msg any) error {
	received, err := s.Recv()
	if err != nil {
		return err
	}
	*msg.(*Req) = *received
	return nil
}

// Control methods for MockBidiServer

// SetSendError injects an error for the next Send call.
func (s *MockBidiServer[Req, Res]) SetSendError(err error) {
	s.state.mu.Lock()
	s.sendErr = err
	s.state.mu.Unlock()
}

// SetRecvError injects an error for the next Recv call.
func (s *MockBidiServer[Req, Res]) SetRecvError(err error) {
	s.state.mu.Lock()
	s.recvErr = err
	s.state.mu.Unlock()
}

// Header returns the header metadata.
func (s *MockBidiServer[Req, Res]) Header() metadata.MD {
	return s.state.header
}

// Trailer returns the trailer metadata.
func (s *MockBidiServer[Req, Res]) Trailer() metadata.MD {
	return s.state.trailer
}

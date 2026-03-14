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
	_ grpc.ClientStreamingServer[any, any] = (*MockClientStreamServer[any, any])(nil)
	_ grpc.ClientStreamingClient[any, any] = (*MockClientStreamClient[any, any])(nil)
)

// mockClientStreamState holds the shared state between client and server views.
type mockClientStreamState[Req, Res any] struct {
	ctx     context.Context
	cancel  context.CancelFunc
	reqCh   chan *Req // client sends requests, server receives
	mu      sync.Mutex
	closed  bool
	resp    *Res
	respSet bool
	header  metadata.MD
	trailer metadata.MD
}

// MockClientStreamClient is the client-side view of a mock client stream.
// It implements grpc.ClientStreamingClient and provides control methods for testing.
type MockClientStreamClient[Req, Res any] struct {
	state   *mockClientStreamState[Req, Res]
	sendErr error
}

// MockClientStreamServer is the server-side view of a mock client stream.
// It implements grpc.ClientStreamingServer and provides control methods for testing.
type MockClientStreamServer[Req, Res any] struct {
	state   *mockClientStreamState[Req, Res]
	recvErr error
}

// NewMockClientStream creates a new mock client stream and returns both views.
func NewMockClientStream[Req, Res any](ctx context.Context) (*MockClientStreamClient[Req, Res], *MockClientStreamServer[Req, Res]) {
	ctx, cancel := context.WithCancel(ctx)
	state := &mockClientStreamState[Req, Res]{
		ctx:     ctx,
		cancel:  cancel,
		reqCh:   make(chan *Req, 100),
		header:  metadata.MD{},
		trailer: metadata.MD{},
	}
	client := &MockClientStreamClient[Req, Res]{state: state}
	server := &MockClientStreamServer[Req, Res]{state: state}
	return client, server
}

// Client-side implementation

// Send implements grpc.ClientStreamingClient.Send
func (c *MockClientStreamClient[Req, Res]) Send(msg *Req) error {
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

// CloseAndRecv implements grpc.ClientStreamingClient.CloseAndRecv
func (c *MockClientStreamClient[Req, Res]) CloseAndRecv() (*Res, error) {
	close(c.state.reqCh)

	// Wait for server to set the response
	c.state.mu.Lock()
	resp := c.state.resp
	c.state.mu.Unlock()

	if resp == nil {
		return nil, io.EOF
	}
	return resp, nil
}

// Header implements grpc.ClientStreamingClient.Header
func (c *MockClientStreamClient[Req, Res]) Header() (metadata.MD, error) {
	return c.state.header, nil
}

// Trailer implements grpc.ClientStreamingClient.Trailer
func (c *MockClientStreamClient[Req, Res]) Trailer() metadata.MD {
	return c.state.trailer
}

// CloseSend implements grpc.ClientStreamingClient.CloseSend
func (c *MockClientStreamClient[Req, Res]) CloseSend() error {
	close(c.state.reqCh)
	return nil
}

// Context implements grpc.ClientStreamingClient.Context
func (c *MockClientStreamClient[Req, Res]) Context() context.Context {
	return c.state.ctx
}

// SendMsg implements grpc.ClientStream.SendMsg
func (c *MockClientStreamClient[Req, Res]) SendMsg(msg any) error {
	return c.Send(msg.(*Req))
}

// RecvMsg implements grpc.ClientStream.RecvMsg
func (c *MockClientStreamClient[Req, Res]) RecvMsg(msg any) error {
	c.state.mu.Lock()
	resp := c.state.resp
	c.state.mu.Unlock()

	if resp == nil {
		return io.EOF
	}
	*msg.(*Res) = *resp
	return nil
}

// Control methods for MockClientStreamClient

// SetSendError injects an error for the next Send call.
func (c *MockClientStreamClient[Req, Res]) SetSendError(err error) {
	c.state.mu.Lock()
	c.sendErr = err
	c.state.mu.Unlock()
}

// Response returns the response set by the server via SendAndClose.
func (c *MockClientStreamClient[Req, Res]) Response() *Res {
	c.state.mu.Lock()
	defer c.state.mu.Unlock()
	return c.state.resp
}

// Close simulates the client connection being dropped or the RPC being cancelled.
func (c *MockClientStreamClient[Req, Res]) Close() {
	c.state.mu.Lock()
	c.state.closed = true
	c.state.mu.Unlock()
	c.state.cancel()
}

// Server-side implementation

// Recv implements grpc.ClientStreamingServer.Recv
func (s *MockClientStreamServer[Req, Res]) Recv() (*Req, error) {
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

// SendAndClose implements grpc.ClientStreamingServer.SendAndClose
func (s *MockClientStreamServer[Req, Res]) SendAndClose(msg *Res) error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if s.state.closed {
		return io.EOF
	}

	// Deep copy the message to simulate wire serialization.
	msgCopy := proto.Clone(any(msg).(proto.Message))
	s.state.resp = any(msgCopy).(*Res)
	s.state.respSet = true
	return nil
}

// Context implements grpc.ClientStreamingServer.Context
func (s *MockClientStreamServer[Req, Res]) Context() context.Context {
	return s.state.ctx
}

// SetHeader implements grpc.ClientStreamingServer.SetHeader
func (s *MockClientStreamServer[Req, Res]) SetHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SendHeader implements grpc.ClientStreamingServer.SendHeader
func (s *MockClientStreamServer[Req, Res]) SendHeader(md metadata.MD) error {
	s.state.header = metadata.Join(s.state.header, md)
	return nil
}

// SetTrailer implements grpc.ClientStreamingServer.SetTrailer
func (s *MockClientStreamServer[Req, Res]) SetTrailer(md metadata.MD) {
	s.state.trailer = metadata.Join(s.state.trailer, md)
}

// SendMsg implements grpc.ClientStreamingServer.SendMsg
func (s *MockClientStreamServer[Req, Res]) SendMsg(msg any) error {
	return s.SendAndClose(msg.(*Res))
}

// RecvMsg implements grpc.ClientStreamingServer.RecvMsg
func (s *MockClientStreamServer[Req, Res]) RecvMsg(msg any) error {
	received, err := s.Recv()
	if err != nil {
		return err
	}
	*msg.(*Req) = *received
	return nil
}

// Control methods for MockClientStreamServer

// SetRecvError injects an error for the next Recv call.
func (s *MockClientStreamServer[Req, Res]) SetRecvError(err error) {
	s.state.mu.Lock()
	s.recvErr = err
	s.state.mu.Unlock()
}

// Header returns the header metadata.
func (s *MockClientStreamServer[Req, Res]) Header() metadata.MD {
	return s.state.header
}

// Trailer returns the trailer metadata.
func (s *MockClientStreamServer[Req, Res]) Trailer() metadata.MD {
	return s.state.trailer
}

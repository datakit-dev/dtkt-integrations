package pkg

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/integrationsdk/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	testbinv1beta "github.com/datakit-dev/dtkt-integrations/testbin/pkg/proto/integration/testbin/v1beta"
)

type EchoService struct {
	testbinv1beta.UnimplementedEchoServiceServer
	mux v1beta1.InstanceMux[*Instance]
}

func NewEchoService(mux v1beta1.InstanceMux[*Instance]) *EchoService {
	return &EchoService{mux: mux}
}

func (s *EchoService) EchoUnary(ctx context.Context, req *testbinv1beta.EchoUnaryRequest) (*testbinv1beta.EchoUnaryResponse, error) {
	if err := sleepCtx(ctx, req.GetDelay().AsDuration()); err != nil {
		return nil, err
	}
	if errStatus := req.GetError(); errStatus != nil && codes.Code(errStatus.GetCode()) != codes.OK {
		return nil, status.ErrorProto(errStatus)
	}
	return &testbinv1beta.EchoUnaryResponse{
		Payload: req.GetPayload(),
		At:      timestamppb.Now(),
	}, nil
}

func (s *EchoService) EchoServerStream(req *testbinv1beta.EchoServerStreamRequest, stream grpc.ServerStreamingServer[testbinv1beta.EchoServerStreamResponse]) error {
	ctx := stream.Context()
	if err := sleepCtx(ctx, req.GetDelay().AsDuration()); err != nil {
		return err
	}

	errStatus := req.GetError()
	hasError := errStatus != nil && codes.Code(errStatus.GetCode()) != codes.OK

	successes := req.GetCount()
	if hasError && req.ErrorOn != nil {
		successes = req.GetErrorOn()
	}

	interval := req.GetInterval().AsDuration()
	for i := int32(0); i < successes; i++ {
		if i > 0 {
			if err := sleepCtx(ctx, interval); err != nil {
				return err
			}
		}
		if err := stream.Send(&testbinv1beta.EchoServerStreamResponse{
			Payload:  req.GetPayload(),
			Sequence: int64(i),
			At:       timestamppb.Now(),
		}); err != nil {
			return err
		}
	}

	if hasError {
		return status.ErrorProto(errStatus)
	}
	return nil
}

func (s *EchoService) EchoClientStream(stream grpc.ClientStreamingServer[testbinv1beta.EchoClientStreamRequest, testbinv1beta.EchoClientStreamResponse]) error {
	ctx := stream.Context()
	var (
		received int64
		last     *structpb.Value
	)
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(&testbinv1beta.EchoClientStreamResponse{
				Received:    received,
				LastPayload: last,
				At:          timestamppb.Now(),
			})
		}
		if err != nil {
			return err
		}
		if err := sleepCtx(ctx, msg.GetDelay().AsDuration()); err != nil {
			return err
		}
		if errStatus := msg.GetError(); errStatus != nil && codes.Code(errStatus.GetCode()) != codes.OK {
			return status.ErrorProto(errStatus)
		}
		received++
		last = msg.GetPayload()
	}
}

func (s *EchoService) EchoBidiStream(stream grpc.BidiStreamingServer[testbinv1beta.EchoBidiStreamRequest, testbinv1beta.EchoBidiStreamResponse]) error {
	ctx := stream.Context()
	var sequence int64
	for {
		msg, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		errStatus := msg.GetError()
		hasError := errStatus != nil && codes.Code(errStatus.GetCode()) != codes.OK

		successes := msg.GetCount()
		if hasError && msg.ErrorOn != nil {
			successes = msg.GetErrorOn()
		}

		delay := msg.GetDelay().AsDuration()
		for i := int32(0); i < successes; i++ {
			if i > 0 && delay > 0 {
				if err := sleepCtx(ctx, delay); err != nil {
					return err
				}
			}
			if err := stream.Send(&testbinv1beta.EchoBidiStreamResponse{
				Payload:  msg.GetPayload(),
				Sequence: sequence,
				At:       timestamppb.Now(),
			}); err != nil {
				return err
			}
			sequence++
		}

		if hasError {
			return status.ErrorProto(errStatus)
		}
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

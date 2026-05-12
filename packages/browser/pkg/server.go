package pkg

import (
	"context"
	"log/slog"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/network"
	"google.golang.org/grpc"
)

type RegisterServiceFunc func(context.Context, grpc.ServiceRegistrar)

func StartServer(ctx context.Context, log *slog.Logger, addr network.Address, regService RegisterServiceFunc) (err error) {
	conn, err := network.NewConnector(addr)
	if err != nil {
		return err
	}

	lis, err := conn.Bind(ctx)
	if err != nil {
		return err
	}
	defer lis.Close()

	srv := grpc.NewServer()
	regService(ctx, srv)

	context.AfterFunc(ctx, srv.GracefulStop)

	return srv.Serve(lis)
}

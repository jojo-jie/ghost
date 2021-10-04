package service

import (
	"context"
	v1 "ghost/api/helloworld/v1"
	"ghost/pkg/track"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestRpcClient(t *testing.T) {
	md := metadata.Pairs(
		"orders", "client_test",
		"orders", "v1.1.0",
		"orders", "sql",
	)
	newCtx:= metadata.NewOutgoingContext(context.Background(), md)
	_, err := track.New("http://localhost:14268/api/traces", "stocks")
	if err != nil {
		return
	}

	opts := make([]grpc.ClientOption, 0, 5)
	opts = append(opts, grpc.WithEndpoint(":6677"),
		grpc.WithMiddleware(middleware.Chain(
			tracing.Client(),
		)))
	conn, err := grpc.DialInsecure(newCtx, opts...)

	if err != nil {
		t.Errorf("%+v", err)
	}
	client := v1.NewGreeterClient(conn)
	reply, err := client.SayHello(newCtx, &v1.HelloRequest{
		UserId: "479870",
	})
	if err != nil {
		t.Error(err)
	}
	defer track.End(newCtx)
	t.Log(reply.UserInfo.Price)
}

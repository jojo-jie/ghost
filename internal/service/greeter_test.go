package service

import (
	"context"
	v1 "ghost/api/helloworld/v1"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

func TestRpcClient(t *testing.T) {
	md := metadata.Pairs(
		"orders", "client_test",
		"ghost", "v1.1.0",
		"orders", "sql",
	)

	newCtx := metadata.NewOutgoingContext(context.Background(), md)
	opts := make([]grpc.ClientOption, 0, 5)
	opts = append(opts, grpc.WithEndpoint(":6677"), grpc.WithOptions(ggrpc.WithInsecure(), ggrpc.WithBlock()))
	conn, err := grpc.Dial(newCtx, opts...)
	if err != nil {
		t.Errorf("%+v", err)
	}
	reply, err := v1.NewGreeterClient(conn).SayHello(newCtx, &v1.HelloRequest{
		UserId: "479870",
	})
	if err != nil {
		t.Error(err)
	}
	t.Log(reply.UserInfo.Price)
}
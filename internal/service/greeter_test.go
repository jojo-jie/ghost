package service

import (
	"context"
	v1 "ghost/api/helloworld/v1"
	"ghost/pkg/jwt"
	"ghost/pkg/track"
	"github.com/go-kratos/etcd/registry"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"io"
	nhttp "net/http"

	jwtv4 "github.com/golang-jwt/jwt/v4"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"testing"
	"time"
)

func TestRpcClient(t *testing.T) {
	md := metadata.Pairs(
		"orders", "client_test",
		"orders", "v1.1.0",
		"orders", "sql",
	)
	newCtx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := track.New("http://localhost:14268/api/traces", "stocks")
	if err != nil {
		return
	}
	dis, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second,
		DialOptions: []grpc2.DialOption{grpc2.WithBlock()},
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer dis.Close()
	endpoint := "discovery://microservices/orders"
	opts := make([]grpc.ClientOption, 0, 3)
	opts = append(opts, grpc.WithEndpoint(endpoint), grpc.WithDiscovery(registry.New(dis)),
		grpc.WithMiddleware(middleware.Chain(
			tracing.Client(),
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return []byte("testKey"), nil
			}),
		)))
	conn, err := grpc.DialInsecure(newCtx, opts...)

	if err != nil {
		t.Errorf("%+v", err)
	}
	client := v1.NewGreeterClient(conn)
	reply, err := client.SayHello(newCtx, &v1.HelloRequest{
		UserId: "479870",
	})
	defer track.End(newCtx)
	if err != nil {
		t.Log(v1.IsContentMissing(err))
		t.Error(err)
		return
	}

	t.Log(reply.UserInfo.Price)
}

func TestGg(t *testing.T) {
	t.Log(os.Hostname())
}

func TestHttpClient(t *testing.T) {
	conn, err := http.NewClient(
		context.Background(),
		http.WithEndpoint("127.0.0.1:4466"),
		http.WithMiddleware(
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return []byte("testKey"), nil
			}),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	req, err := nhttp.NewRequest(nhttp.MethodGet, "http://127.0.0.1:4466/helloworld/479870", nil)
	if err != nil {
		t.Fatal(err)
	}
	do, err := conn.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer do.Body.Close()
	body, err := io.ReadAll(do.Body)
	t.Log(string(body))
}

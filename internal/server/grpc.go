package server

import (
	"context"
	v1 "ghost/api/helloworld/v1"
	"ghost/internal/conf"
	"ghost/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *grpc.Server {
	var opts = []grpc.ServerOption{
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(
					// 设置中间件打印日志
					recovery.WithLogger(log.DefaultLogger),
					// 设置服务异常时可以使用自定义的 handler 进行处理，例如投递异常信息到 sentry。
					recovery.WithHandler(func(ctx context.Context, req, err interface{}) error {
						return nil
					}),
				),
				tracing.Server(),
				logging.Server(log.DefaultLogger),
			),
		),
	}
	middleware.Chain()
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	v1.RegisterGreeterServer(srv, greeter)
	return srv
}

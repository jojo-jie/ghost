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
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new a HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			middleware.Chain(
				recovery.Recovery(
					recovery.WithLogger(log.DefaultLogger),
					recovery.WithHandler(func(ctx context.Context, req, err interface{}) error {
						return nil
					}),
				),
				tracing.Server(),
				logging.Server(log.DefaultLogger),
			),
			/*selector.Server(jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
				return []byte(c.GetJwtKey()), nil
			})).Prefix("/helloworld").Build(),*/
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	return srv
}

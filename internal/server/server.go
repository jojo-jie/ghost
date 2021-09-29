package server

import (
	"ghost/internal/conf"
	"ghost/internal/service"
	"ghost/pkg/track"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(InitGlobalTracer, NewHTTPServer, NewGRPCServer)

// InitGlobalTracer set trace provider
func InitGlobalTracer(c *conf.Server, greeter *service.GreeterService, logger log.Logger, name string) (*tracesdk.TracerProvider, error) {
	return track.New(c.GetJaegerUrl(), name)
}

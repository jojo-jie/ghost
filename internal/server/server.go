package server

import (
	"ghost/internal/conf"
	"ghost/internal/service"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(InitGlobalTracer, NewHTTPServer, NewGRPCServer)

// InitGlobalTracer set trace provider
func InitGlobalTracer(c *conf.Server, greeter *service.GreeterService, logger log.Logger, name string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.GetJaegerUrl())))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String("v0.11"),
			semconv.HostNameKey.String("hostname1"),
			attribute.String("environment", "dev"),
			attribute.Int64("ID", 1),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

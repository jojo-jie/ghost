package track

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"net/http"
	"strings"
)

// New jaeger 初始化
func New(endpoint, name string) (*tracesdk.TracerProvider, error) {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		return nil, err
	}
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate based on the parent span to 100%
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		// Always be sure to batch in production.
		// time range 5000 msec tracesdk.WithBatchTimeout()
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(name),

			semconv.ServiceVersionKey.String("v0.11"),
			semconv.HostNameKey.String("hostname1"),
			semconv.NetHostIPKey.String("192.168.1.34"),
			attribute.String("environment", "dev"),
			attribute.String("exporter", "jaeger"),
		)),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

type SpanOption func(span trace.Span)

func InjectHttp(ctx context.Context, req *http.Request) SpanOption {
	return func(span trace.Span) {
		uberTraceId := make([]string, 4, 4)
		uberTraceId[0] = span.SpanContext().TraceID().String()
		uberTraceId[1] = span.SpanContext().SpanID().String()
		uberTraceId[2] = trace.SpanContextFromContext(ctx).SpanID().String()
		uberTraceId[3] = "1"
		//https://www.jaegertracing.io/docs/1.18/client-libraries/#tracespan-identity
		//跨应用http uber-trace-id
		req.Header.Set("uber-trace-id", strings.Join(uberTraceId, ":"))
	}
}

func SetAttributes(args string) SpanOption {
	return func(span trace.Span) {
		span.SetAttributes(semconv.ServiceNameKey.String(args))
	}
}

func Start(ctx context.Context, tracer trace.Tracer, spanName string) (newCtx context.Context, finish func(...SpanOption)) {
	if ctx == nil {
		ctx = context.Background()
	}
	newCtx, span := tracer.Start(ctx, spanName)
	finish = func(option ...SpanOption) {
		for _, o := range option {
			o(span)
		}
		span.End()
	}
	return newCtx, finish
}

// End client push end
func End(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*tracesdk.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return errors.New("track fail")
}

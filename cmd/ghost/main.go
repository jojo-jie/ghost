package main

import (
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/sdk/trace"
	grpc2 "google.golang.org/grpc"
	"os"
	"time"

	"ghost/internal/conf"
	"github.com/go-kratos/etcd/registry"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
	// Client is third service client
	Client interface{}

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "/Users/kirito/workspace/ghost/configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server, provider *trace.TracerProvider, client *clientv3.Client) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
		kratos.Registrar(registry.New(client)),
	)
}

func main() {
	flag.Parse()
	fmt.Println("====", flagconf)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
		/*config.WithDecoder(func(value *config.KeyValue, m map[string]interface{}) error {
			return nil
		}),
		config.WithResolver(func(m map[string]interface{}) error {
			return nil
		}),*/
	)
	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	c.Watch("server.jwt_key", func(key string, value config.Value) {
		s, _ := value.String()
		bc.Server.JwtKey = s
	})

	// etcd 初始化
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{bc.GetServer().GetEtcdUrl()},
		DialTimeout: time.Second,
		DialOptions: []grpc2.DialOption{grpc2.WithBlock()},
	})
	if err != nil {
		panic(err)
	}
	Client = client
	defer client.Close()

	Name = bc.GetServer().GetName()
	Version = bc.GetServer().GetVersion()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace_id", tracing.TraceID(),
		"span_id", tracing.SpanID(),
	)

	app, cleanup, err := initApp(bc.Server, bc.Data, logger, Name, client)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

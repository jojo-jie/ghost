package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"go.opentelemetry.io/otel/sdk/trace"
	grpc2 "google.golang.org/grpc"
	"os"
	"time"

	"ghost/internal/conf"
	etcdConf "github.com/go-kratos/etcd/config"
	"github.com/go-kratos/etcd/registry"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/encoding"
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

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "/Users/kirito/workspace/ghost/configs/config.json", "config path, eg: -conf config.yaml")
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

	// etcd 初始化
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{os.Getenv("ETCD_URL")},
		DialTimeout: time.Second,
		DialOptions: []grpc2.DialOption{grpc2.WithBlock()},
	})
	if err != nil {
		panic(err)
	}
	defer client.Close()
	configKey := "ghost"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	v, err := client.Get(ctx, configKey)
	if err != nil {
		if err == context.Canceled {
			panic(err)
		}
	}

	if v.Count == 0 {
		all, err := os.ReadFile(flagconf)
		if err != nil {
			panic(err)
		}
		_, err = client.Put(ctx, configKey, string(all))
		if err != nil {
			panic(err)
		}
	}
	source, err := etcdConf.New(client, etcdConf.Path(configKey), etcdConf.Context(ctx))
	if err != nil {
		return
	}

	var bc conf.Bootstrap
	var app *kratos.App
	var cleanup func()
	c := config.New(
		config.WithSource(
			source,
		),
		config.WithDecoder(func(src *config.KeyValue, target map[string]interface{}) error {
			src.Format = "json"
			if codec := encoding.GetCodec(src.Format); codec != nil {
				return codec.Unmarshal(src.Value, &target)
			}
			return fmt.Errorf("unsupported key: %s format: %s", src.Key, src.Format)
		}),
		config.WithResolver(func(m map[string]interface{}) error {
			if codec := encoding.GetCodec("json"); codec != nil {
				marshal, err := codec.Marshal(m)
				if err != nil {
					return err
				}
				return codec.Unmarshal(marshal, &bc)
			}
			return nil
		}),
	)
	if err := c.Load(); err != nil {
		panic(err)
	}

	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

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

	app, cleanup, err = initApp(bc.GetServer(), bc.GetData(), logger, bc.GetServer().GetName(), client)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

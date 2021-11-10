// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"ghost/internal/biz"
	"ghost/internal/conf"
	"ghost/internal/data"
	"ghost/internal/server"
	"ghost/internal/service"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// initApp init kratos application.
// ...interface{} init external service example etcd
func initApp(*conf.Server, *conf.Data, log.Logger, string, *clientv3.Client) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}

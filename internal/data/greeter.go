package data

import (
	"context"
	"ghost/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

type greeterRepo struct {
	data *Data
	log  *log.Helper
}

// NewGreeterRepo .
func NewGreeterRepo(data *Data, logger log.Logger) biz.GreeterRepo {
	return &greeterRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *greeterRepo) CreateGreeter(ctx context.Context, g *biz.Greeter) error {
	return nil
}

func (r *greeterRepo) UpdateGreeter(ctx context.Context, g *biz.Greeter) error {
	return nil
}

func (r *greeterRepo) ShowGreeter(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	var order biz.Greeter
	result := r.data.Db.Model(g).Where("user_id = ?", g.UserId).First(&order)
	return &order, result.Error
}

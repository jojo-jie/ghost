package data

import (
	"context"
	"ghost/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"time"
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
	location, _ := time.LoadLocation("Asia/Shanghai")
	nowTime := time.Now().In(location).Format("2006-01-02 15:04:05")
	r.data.Db.Model(g).WithContext(ctx).Update("operating_time", nowTime)
	return nil
}

func (r *greeterRepo) ShowGreeter(ctx context.Context, g *biz.Greeter) (*biz.Greeter, error) {
	var order biz.Greeter
	result := r.data.Db.WithContext(ctx).Model(g).Where("user_id = ?", g.UserId).First(&order)
	return &order, result.Error
}

package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
)

type Greeter struct {
	UserId   int      `gorm:"primary_key" json:"user_id"`
	Nickname string   `json:"third_shop_id"`
	Account  string   `json:"third_orderid"`
	UserInfo string `json:"full_order_json"`
}

type UserInfo struct {
	Cid     int64    `json:"cid"`
	Num     int64    `json:"num"`
	Oid     string `json:"oid"`
	Price   string `json:"price"`
	Title   string `json:"title"`
	EndTime string `json:"end_time"`
}

type GreeterRepo interface {
	CreateGreeter(context.Context, *Greeter) error
	UpdateGreeter(context.Context, *Greeter) error
	ShowGreeter(context.Context, *Greeter) (*Greeter, error)
}

type GreeterUsecase struct {
	repo GreeterRepo
	log  *log.Helper
}

func NewGreeterUsecase(repo GreeterRepo, logger log.Logger) *GreeterUsecase {
	return &GreeterUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *GreeterUsecase) Create(ctx context.Context, g *Greeter) error {
	return uc.repo.CreateGreeter(ctx, g)
}

func (uc *GreeterUsecase) Update(ctx context.Context, g *Greeter) error {
	return uc.repo.UpdateGreeter(ctx, g)
}

func (uc *GreeterUsecase) Show(ctx context.Context, g *Greeter) (*Greeter, error) {
	return uc.repo.ShowGreeter(ctx, g)
}

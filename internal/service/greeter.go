package service

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strconv"

	v1 "ghost/api/helloworld/v1"
	"ghost/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer

	uc  *biz.GreeterUsecase
	log *log.Helper
}

// NewGreeterService new a greeter service.
func NewGreeterService(uc *biz.GreeterUsecase, logger log.Logger) *GreeterService {
	return &GreeterService{uc: uc, log: log.NewHelper(logger)}
}

// SayHello implements helloworld.GreeterServer
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloReply, error) {
	s.log.WithContext(ctx).Infof("SayHello Received: %v", in.GetUserId())
	userId, err := strconv.Atoi(in.GetUserId())
	if err != nil {
		return nil, v1.ErrorContentMissing("搞啥呢 %s", err)
	}
	data, err := s.uc.Show(ctx, &biz.Greeter{UserId: userId})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, v1.ErrorContentMissing("啥也没有 %s", err)
	}

	var d biz.UserInfo
	err = json.Unmarshal([]byte(data.UserInfo), &d)
	if err != nil {
		return nil, v1.ErrorContentMissing("失败 %s", err)
	}
	return &v1.HelloReply{
		UserId:   int32(data.UserId),
		Nickname: data.Nickname,
		Account:  data.Account,
		UserInfo: &v1.UserInfo{
			Cid:     d.Cid,
			Num:     d.Num,
			Oid:     d.Oid,
			Price:   d.Price,
			Title:   d.Title,
			EndTime: d.EndTime,
		},
	}, nil
}

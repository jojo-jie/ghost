package service

import (
	"context"
	"encoding/json"
	"errors"
	v1 "ghost/api/helloworld/v1"
	"ghost/internal/biz"
	"ghost/pkg/bapi"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
	"strconv"
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
	md, _ := metadata.FromIncomingContext(ctx)
	strings := md["orders"]
	s.log.WithContext(ctx).Infof("SayHello Received Md: %v", strings)
	userId, err := strconv.Atoi(in.GetUserId())
	if err != nil {
		return nil, v1.ErrorContentMissing("搞啥呢 %s", err)
	}


	data, err := s.uc.Show(ctx, &biz.Greeter{UserId: userId})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, v1.ErrorContentMissing("啥也没有 %s", err)
	}

	api := bapi.NewApi("http://127.0.0.1:4433")
	body, err := api.GetTagList(ctx)

	err = s.uc.Update(ctx, &biz.Greeter{UserId: userId})
	if err != nil {
		return nil, err
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
			TagList: body,
		},
	}, nil
}

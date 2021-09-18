package service

import (
	"context"
	"encoding/json"
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	"strconv"
	"time"

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
	db(ctx)
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

func db(ctx context.Context)  {
	tracer := otel.Tracer("mysql")
	kind := trace.SpanKindServer
	duration, _ := time.ParseDuration("100ns")
	ctx, span := tracer.Start(ctx,
		"sql",
		trace.WithAttributes(
			attribute.String("event", "eventName"),
			attribute.String("command", "commandName"),
			attribute.String("query", "select * from orders"),
			attribute.Int64("queryId", 123),
			attribute.String("ms", duration.String()),
		),
		trace.WithSpanKind(kind),
	)

	span.SetAttributes(attribute.Bool("error", true))
	span.SetStatus(500, "有毒了!!!")
	span.End()
}

package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "ghost/api/helloworld/v1"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	APP_KEY    = "eddycjy"
	APP_SECRET = "go-programming-tour-book"
)

type AccessToken struct {
	Token string
}

func (a Api) getAccessToken(ctx context.Context) (string, error) {
	path := fmt.Sprintf("%s?app_key=%s&app_secret=%s", "auth", APP_KEY, APP_SECRET)
	body, err := a.httpGet(ctx, path)
	if err != nil {
		return "", err
	}
	var accessToken AccessToken
	err = json.Unmarshal(body, &accessToken)
	if err != nil {
		return "", err
	}
	return accessToken.Token, nil
}

type Api struct {
	Url string
}

func NewApi(url string) *Api {
	return &Api{
		Url: url,
	}
}

func (a *Api) httpGet(ctx context.Context, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/%s", a.Url, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	//https://www.jaegertracing.io/docs/1.18/client-libraries/#tracespan-identity
	//跨应用http uber-trace-id
	tracer := otel.Tracer("")
	newCtx,span:=tracer.Start(ctx, "HTTP GET: "+a.Url)
	span.SetAttributes(semconv.ServiceNameKey.String(path))
	defer span.End()
	req = req.WithContext(newCtx)
	uberTraceId:=make([]string, 4,4)
	uberTraceId[0] = span.SpanContext().TraceID().String()
	uberTraceId[1] = span.SpanContext().SpanID().String()
	uberTraceId[2] = trace.SpanContextFromContext(ctx).SpanID().String()
	uberTraceId[3] = "1"

	req.Header.Set("uber-trace-id", strings.Join(uberTraceId, ":"))
	client := &http.Client{Timeout: time.Millisecond * 300}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

type Tags struct {
	List []*v1.Tag `json:"list"`
}

func (a Api) GetTagList(ctx context.Context) ([]*v1.Tag, error) {
	token, err := a.getAccessToken(ctx)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s?token=%s", "api/v1/tags", token)
	body, err := a.httpGet(ctx, path)
	if err != nil {
		return nil, err
	}
	var tags Tags
	err = json.Unmarshal(body, &tags)
	if err != nil {
		return nil, err
	}
	return tags.List, nil
}

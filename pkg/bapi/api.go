package bapi

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "ghost/api/helloworld/v1"
	"ghost/pkg/track"
	"go.opentelemetry.io/otel"
	"io"
	"net/http"
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
	newCtx,finish:=track.Start(ctx, otel.Tracer(""), "HTTP GET: "+a.Url)
	finish(track.SetAttributes(path), track.InjectHttp(ctx, req))
	req = req.WithContext(newCtx)
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

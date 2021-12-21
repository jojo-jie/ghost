package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	v1 "ghost/api/helloworld/v1"
	"ghost/pkg/track"
	"github.com/go-kratos/etcd/registry"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/selector/filter"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"io"
	nhttp "net/http"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"unsafe"

	jwtv4 "github.com/golang-jwt/jwt/v4"
	clientv3 "go.etcd.io/etcd/client/v3"
	grpc2 "google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
	"os"
	"testing"
	"time"
)

func TestRpcClient(t *testing.T) {
	md := metadata.Pairs(
		"orders", "client_test",
		"orders", "v1.1.0",
		"orders", "sql",
	)
	newCtx := metadata.NewOutgoingContext(context.Background(), md)
	_, err := track.New("http://localhost:14268/api/traces", "stocks")
	if err != nil {
		return
	}
	dis, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second,
		DialOptions: []grpc2.DialOption{grpc2.WithBlock()},
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer dis.Close()
	wrr.WithFilter(filter.Version("2.1.1"))
	endpoint := "discovery://microservices/orders"
	opts := make([]grpc.ClientOption, 0, 3)
	opts = append(opts, grpc.WithEndpoint(endpoint), grpc.WithDiscovery(registry.New(dis)),
		grpc.WithMiddleware(middleware.Chain(
			tracing.Client(),
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return []byte("testKey"), nil
			}),
		)), grpc.WithBalancerName(wrr.Name))
	conn, err := grpc.DialInsecure(newCtx, opts...)

	if err != nil {
		t.Errorf("%+v", err)
	}
	client := v1.NewGreeterClient(conn)
	reply, err := client.SayHello(newCtx, &v1.HelloRequest{
		UserId: "479870",
	})
	defer track.End(newCtx)
	if err != nil {
		t.Log(v1.IsContentMissing(err))
		t.Error(err)
		return
	}

	t.Log(reply.UserInfo.Price)
}

func TestGg(t *testing.T) {
	t.Log(os.Hostname())
}

func TestHttpClient(t *testing.T) {
	conn, err := http.NewClient(
		context.Background(),
		http.WithEndpoint("127.0.0.1:4466"),
		http.WithMiddleware(
			jwt.Client(func(token *jwtv4.Token) (interface{}, error) {
				return []byte("testKey"), nil
			}),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	req, err := nhttp.NewRequest(nhttp.MethodGet, "http://127.0.0.1:4466/helloworld/479870", nil)
	if err != nil {
		t.Fatal(err)
	}
	do, err := conn.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer do.Body.Close()
	body, err := io.ReadAll(do.Body)
	t.Log(string(body))
}

func TestFormatInt(t *testing.T) {
	t.Log(strconv.FormatInt(3002604296, 35))
}

func TestHash(t *testing.T) {
	s := `{"shop_id": 123, "code": 1, "success": 1, "extra": "shop_id 123 is authorized successfully", "data": {"more_info": "more info"}, "timestamp": 1470198856}`
	h := hmac.New(sha256.New, []byte("lllll"))
	h.Write([]byte(s))
	t.Log(h.Sum(nil))
	t.Logf("%s", string(h.Sum(nil)))
	t.Logf("%x", h.Sum(nil))

	cc := "123asdLi乐?&{[]#3@"
	t.Log([]byte(cc))
	t.Log([]rune(cc))
}

func TestDR(t *testing.T) {
	t.Log("returns", demo())
}

func demo() (ii int) {
	defer func() {
		ii++
		fmt.Println("defer1", ii)
	}()

	defer func() {
		ii++
		fmt.Println("defer2", ii)
	}()
	return
}

type W struct {
	a, b int
}

func TestUintptr(t *testing.T) {
	var w *W = new(W)
	t.Log(w.a, w.b)

	b := unsafe.Pointer(uintptr(unsafe.Pointer(w)) + unsafe.Offsetof(w.b))
	*((*int)(b)) = 10
	t.Log(w.a, w.b)
}

// https://mp.weixin.qq.com/s/kQLAnh-frOALCDNU924zxQ
func TestFanIn(t *testing.T) {
	// create two sample message and stop channels
	mc1, sc1 := generate("message from generator 1", 200*time.Millisecond)
	mc2, sc2 := generate("message from generator 2", 200*time.Millisecond)

	// multiplex message channels
	mmc, wg1 := multiplex(mc1, mc2)

	// create errs channel for graceful shutdown
	errs := make(chan error)

	// wait for interrupt or terminate signal
	go func() {
		sc := make(chan os.Signal, 1)
		signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s signal received", <-sc)
	}()

	// wait for multiplexed messages
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	go func() {
		defer wg2.Done()
		for m := range mmc {
			t.Log(m)
		}
	}()

	// wait for errors
	if err := <-errs; err != nil {
		t.Log(err.Error())
	}

	stopGenerating(mc1, sc1)
	stopGenerating(mc2, sc2)
	wg1.Wait()

	// close multiplexed messages channel
	close(mmc)
	wg2.Wait()
}

func generate(message string, interval time.Duration) (chan string, chan struct{}) {
	mc := make(chan string)
	sc := make(chan struct{})

	go func() {
		defer func() {
			close(sc)
		}()
		for {
			select {
			case <-sc:
				return
			default:
				time.Sleep(interval)
				mc <- message
			}
		}
	}()
	return mc, sc
}

func stopGenerating(mc chan string, sc chan struct{}) {
	sc <- struct{}{}
	close(mc)
}

// 多路复用函数
func multiplex(mcs ...chan string) (chan string, *sync.WaitGroup) {
	mmc := make(chan string)
	wg := &sync.WaitGroup{}

	for _, mc := range mcs {
		wg.Add(1)
		go func(mc chan string, wg *sync.WaitGroup) {
			defer wg.Done()
			for m := range mc {
				mmc <- m
			}
		}(mc, wg)
	}
	return mmc, wg
}

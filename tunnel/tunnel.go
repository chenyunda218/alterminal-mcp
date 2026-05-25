package tunnel

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/advbet/sseclient"
)

type Tunnel struct {
	BaseUrl   string
	TunnelId  string
	SecretKey string
}

func New(baseUrl, tunnelId, secretKey string) *Tunnel {
	return &Tunnel{
		BaseUrl:   baseUrl,
		TunnelId:  tunnelId,
		SecretKey: secretKey,
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (t *Tunnel) Connect() context.CancelFunc {
	customTransport := &http.Transport{}
	customClient := &http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("mcp_secret_key", t.SecretKey)
			return customTransport.RoundTrip(req)
		}),
	}

	c := sseclient.New(t.BaseUrl+"/mcps/"+t.TunnelId+"/sse", "")
	c.HTTPClient = customClient
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	eventHandler := func(event *sseclient.Event) error {
		// 打印接收到的事件名、ID和数据
		log.Printf("收到事件 [%s] id=%s data=%s", event.Event, event.ID, string(event.Data))
		return nil
	}

	// 4. 定义错误处理器（返回 false 停止重连，true 继续重连）
	errorHandler := func(err error) error {
		log.Printf("发生错误: %v", err)
		return nil
	}

	// 5. 启动连接，开始接收事件
	log.Println("开始连接 SSE 服务器...")
	c.Start(ctx, eventHandler, errorHandler)
	return cancel
}

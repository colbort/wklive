package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type Resp map[string]any

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 测试阶段直接放开
		return true
	},
}

func TestSocket(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/app/market/ws/market", tickWsHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	conn, _, err := websocket.DefaultDialer.Dial(
		"ws"+strings.TrimPrefix(server.URL, "http")+"/app/market/ws/market",
		nil,
	)
	if err != nil {
		t.Fatalf("dial market websocket: %v", err)
	}
	defer conn.Close()

	var connected Resp
	if err := conn.ReadJSON(&connected); err != nil {
		t.Fatalf("read connected message: %v", err)
	}
	if connected["type"] != "connected" {
		t.Fatalf("unexpected connected message: %#v", connected)
	}

	if err := conn.WriteJSON(Resp{"type": "subscribe", "symbol": "BTCUSDT"}); err != nil {
		t.Fatalf("write subscribe message: %v", err)
	}
	for _, expectedType := range []string{"subscribed", "kline"} {
		var response Resp
		if err := conn.ReadJSON(&response); err != nil {
			t.Fatalf("read %s message: %v", expectedType, err)
		}
		if response["type"] != expectedType {
			t.Fatalf("expected %s message, got %#v", expectedType, response)
		}
	}
}

func tickWsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("incoming request: method=%s host=%s uri=%s origin=%s ua=%s",
		r.Method,
		r.Host,
		r.RequestURI,
		r.Header.Get("Origin"),
		r.UserAgent(),
	)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade failed: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer conn.Close()

	log.Printf("upgrade success: remote=%s\n", r.RemoteAddr)

	// 连接成功先发一条 connected
	_ = conn.WriteJSON(Resp{
		"type":     "connected",
		"serverTs": time.Now().UnixMilli(),
	})

	for {
		msgType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read closed: %v\n", err)
			return
		}

		log.Printf("recv msgType=%d data=%s\n", msgType, string(data))

		// 尝试按 json 解析
		var req map[string]any
		if err := json.Unmarshal(data, &req); err == nil {
			// 如果是 subscribe，返回订阅成功
			if t, ok := req["type"].(string); ok && t == "subscribe" {
				_ = conn.WriteJSON(Resp{
					"type":     "subscribed",
					"serverTs": time.Now().UnixMilli(),
					"echo":     req,
				})

				// 模拟推一条行情数据
				_ = conn.WriteJSON(Resp{
					"type":     "kline",
					"serverTs": time.Now().UnixMilli(),
					"symbol":   "BTCUSDT",
					"interval": "1m",
					"price":    "68432.12",
				})
				continue
			}
		}

		// 其他消息原样回显
		err = conn.WriteMessage(msgType, []byte(`{"type":"echo","data":`+jsonString(string(data))+`}`))
		if err != nil {
			log.Printf("write failed: %v\n", err)
			return
		}
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

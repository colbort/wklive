package main

import (
	"encoding/json"
	"log"
	"net/http"
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
	http.HandleFunc("/app/itick/ws/itick", tickWsHandler)

	addr := ":7777"
	log.Printf("websocket test server listening on %s\n", addr)
	log.Printf("ws url: ws://127.0.0.1%s/app/itick/ws/itick\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
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

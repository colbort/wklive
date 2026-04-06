// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_public

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/itick"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

type TickWsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTickWsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TickWsLogic {
	return &TickWsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// 客户端约定
// 首帧订阅
// {
//   "type": "subscribe",
//   "topics": [
//     {
//       "topic": "depth",    // 参数类型，（depth：盘口、quote：报价、tick：成交、kline：K线）
//       "categoryCode": "crypto",
//       "symbol": "BTCUSDT",
//       "market": "BA",
//       "interval": "1m"    //  kline 订阅必传。（1m，5m，15m，30m，1h，1d，1w，1mo）
//     }
//   ]
// }
// 心跳 ping
// {
//   "type": "ping",
//   "clientTs": 1711888888888
// }
// 服务端 pong
// {
//   "type": "pong",
//   "clientTs": 1711888888888,
//   "serverTs": 1711888889999
// }

func (l *TickWsLogic) TickWs(conn *websocket.Conn, r *http.Request) error {
	defer conn.Close()

	const (
		heartbeatTimeout = 70 * time.Second // 超过这个时间没收到任何客户端消息，直接断开
		maxPingInterval  = 40 * time.Second // 两次 ping 最大允许间隔
	)

	nowMs := func() int64 {
		return time.Now().UnixMilli()
	}

	if err := conn.SetReadDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
		return err
	}

	var firstMsg types.WsMessage
	if err := conn.ReadJSON(&firstMsg); err != nil {
		return err
	}

	if err := conn.SetReadDeadline(time.Now().Add(heartbeatTimeout)); err != nil {
		return err
	}

	if firstMsg.Type != "subscribe" || len(firstMsg.Topics) == 0 {
		_ = conn.WriteJSON(map[string]any{
			"type":     "error",
			"code":     400,
			"message":  "invalid first ws message",
			"serverTs": nowMs(),
		})
		return fmt.Errorf("invalid first ws message")
	}

	ctx, cancel := context.WithCancel(l.ctx)
	defer cancel()

	topics := make([]*itick.SubscribeTopic, 0, len(firstMsg.Topics))
	for _, topic := range firstMsg.Topics {
		topics = append(topics, &itick.SubscribeTopic{
			Topic:        topic.Topic,
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		})
	}

	stream, err := l.svcCtx.ItickCli.SubscribeStream(ctx, &itick.SubscribeRequest{
		Topics: topics,
	})
	if err != nil {
		_ = conn.WriteJSON(map[string]any{
			"type":     "error",
			"code":     500,
			"message":  err.Error(),
			"serverTs": nowMs(),
		})
		return err
	}

	replyCh := make(chan *itick.PushReply, 16)
	writeCh := make(chan any, 16)
	errCh := make(chan error, 2)

	var lastPingAt int64 // UnixMilli，0 表示还没收到过 ping

	// 读 gRPC stream
	go func() {
		for {
			reply, err := stream.Recv()
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			select {
			case replyCh <- reply:
			case <-ctx.Done():
				return
			}
		}
	}()

	// 读 ws 客户端消息，处理心跳
	go func() {
		for {
			var msg types.WsMessage
			if err := conn.ReadJSON(&msg); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			recvAt := time.Now()

			// 任何客户端消息到达，都刷新读超时
			if err := conn.SetReadDeadline(recvAt.Add(heartbeatTimeout)); err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}

			switch msg.Type {
			case "ping":
				currentPingAt := recvAt.UnixMilli()

				// 校验两次 ping 间隔
				if lastPingAt > 0 {
					interval := time.Duration(currentPingAt-lastPingAt) * time.Millisecond
					if interval > maxPingInterval {
						select {
						case writeCh <- map[string]any{
							"type":      "error",
							"code":      4001,
							"message":   "ping interval exceeded",
							"clientTs":  msg.ClientTs,
							"serverTs":  nowMs(),
							"maxMillis": maxPingInterval.Milliseconds(),
							"actualMs":  interval.Milliseconds(),
						}:
						default:
						}

						select {
						case errCh <- fmt.Errorf("ping interval exceeded: %dms", interval.Milliseconds()):
						default:
						}

					}
				}

				lastPingAt = currentPingAt

				// 回复 pong：带回客户端时间戳 + 当前服务端时间戳
				select {
				case writeCh <- map[string]any{
					"type":     "pong",
					"clientTs": msg.ClientTs,
					"serverTs": nowMs(),
				}:
				case <-ctx.Done():
					return
				}

			case "subscribe":
				// 这里先忽略后续 subscribe；后面你要动态订阅再扩展
			default:
				// 未识别消息可以忽略，也可以返回错误
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-errCh:
			cancel()
			return err

		case reply := <-replyCh:
			if err := conn.WriteJSON(map[string]any{
				"type":         "push",
				"topic":        reply.Topic,
				"categoryCode": reply.CategoryCode,
				"symbol":       reply.Symbol,
				"market":       reply.Market,
				"interval":     reply.Interval,
				"payload":      json.RawMessage(reply.Payload),
				"serverTs":     nowMs(),
			}); err != nil {
				cancel()
				return err
			}

		case out := <-writeCh:
			if err := conn.WriteJSON(out); err != nil {
				cancel()
				return err
			}
		}
	}
}

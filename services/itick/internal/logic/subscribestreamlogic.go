package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"wklive/proto/itick"
	"wklive/services/itick/internal/socket/client"
	"wklive/services/itick/internal/socket/server"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeStreamLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubscribeStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeStreamLogic {
	return &SubscribeStreamLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅数据流
func (l *SubscribeStreamLogic) SubscribeStream(in *itick.SubscribeRequest, stream itick.ItickApp_SubscribeStreamServer) error {
	sub := l.svcCtx.Hub.NewSubscriber(256)

	msgs := make([]server.ClientMessage, 0, len(in.Topics))
	for _, topic := range in.Topics {
		msgs = append(msgs, server.ClientMessage{
			Topic:        server.Topic(topic.Topic),
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		})
	}

	subscribed := make([]server.ClientMessage, 0, len(msgs))

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.svcCtx.ItickManager.RemoveGlobalSubscriptions(ctx, subscribed); err != nil {
			logx.Errorf("remove global subscriptions failed, err=%v", err)
		}
	}()
	defer l.svcCtx.Hub.RemoveSubscriber(sub)

	// 1. 先注册全局订阅
	if err := l.svcCtx.ItickManager.AddGlobalSubscriptions(stream.Context(), msgs); err != nil {
		return err
	}

	for _, msg := range msgs {
		// 2. 再注册本地 Hub 订阅
		if err := l.svcCtx.Hub.Subscribe(sub, msg); err != nil {
			_ = l.svcCtx.ItickManager.RemoveGlobalSubscriptions(stream.Context(), msgs)
			return err
		}

		subscribed = append(subscribed, msg)
	}

	// Redis lease may already exist and therefore not publish an add event.
	// After the local Hub subscriptions are registered, explicitly ensure the
	// current leader has subscribed upstream.
	if err := l.svcCtx.ItickManager.EnsureUpstreamSubscriptions(stream.Context(), msgs); err != nil {
		logx.Errorf("ensure upstream subscriptions failed, err=%v", err)
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil

		case pushMsg, ok := <-sub.C():
			if !ok {
				return nil
			}

			payloadBytes, err := json.Marshal(pushMsg.Payload)
			if err != nil {
				continue
			}

			switch pushMsg.Topic {
			case server.TopicKline:
				go func(msg server.ServerMessage) {
					payload := msg.Payload
					kline, ok := payload.(*client.KlinePayload)
					if !ok || kline == nil {
						return
					}

					data := &models.CoinKline{
						CategoryCode: msg.CategoryCode,
						Market:       msg.Market,
						Symbol:       msg.Symbol,
						Interval:     kline.Interval,
						Ts:           kline.Ts,
						Open:         kline.Open,
						High:         kline.High,
						Low:          kline.Low,
						Close:        kline.Close,
						Volume:       kline.Volume,
						Turnover:     kline.Turnover,
					}

					if err := l.svcCtx.Writer.Enqueue(data); err != nil {
						logx.Errorf("enqueue kline error: categoryCode=%s symbol=%s interval=%s ts=%d err=%v",
							data.CategoryCode, data.Symbol, data.Interval, data.Ts, err)
					}
				}(pushMsg)
			case server.TopicQuote:
				key := fmt.Sprintf("itick:quote:%s:%s:%s", pushMsg.CategoryCode, pushMsg.Market, pushMsg.Symbol)
				l.svcCtx.Cache.SetCtx(l.ctx, key, pushMsg.Payload)
			}

			if err := stream.Send(&itick.PushReply{
				Topic:        string(pushMsg.Topic),
				CategoryCode: pushMsg.CategoryCode,
				Market:       pushMsg.Market,
				Symbol:       pushMsg.Symbol,
				Interval:     pushMsg.Interval,
				Payload:      payloadBytes,
			}); err != nil {
				return err
			}
		}
	}
}

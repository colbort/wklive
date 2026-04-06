package logic

import (
	"context"
	"encoding/json"

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
	defer l.svcCtx.Hub.RemoveSubscriber(sub)

	for _, topic := range in.Topics {
		msg := server.ClientMessage{
			Topic:        server.Topic(topic.Topic),
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		}
		if err := l.svcCtx.Hub.Subscribe(sub, msg); err != nil {
			return err
		}
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
			if pushMsg.Topic == server.TopicKline {
				go func(msg server.ServerMessage) {
					// ⚠️ 先把 payload 拿出来（避免闭包踩坑）
					payload := msg.Payload
					kline, ok := payload.(*client.KlinePayload)
					if !ok || kline == nil {
						return
					}

					// 转换成 Mongo 模型
					data := &models.CoinKline{
						CategoryCode: msg.CategoryCode, // 你自己的字段
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

					// 入库（建议用 factory）
					if err := l.svcCtx.Writer.Enqueue(data); err != nil {
						logx.Errorf("enqueue kline error: categoryCode=%s symbol=%s interval=%s ts=%d err=%v", data.CategoryCode, data.Symbol, data.Interval, data.Ts, err)
					}
				}(pushMsg)
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

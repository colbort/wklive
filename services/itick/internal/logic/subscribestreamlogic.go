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

	msgs := normalizeUniqueClientMessages(in)
	if len(msgs) == 0 {
		l.svcCtx.Hub.RemoveSubscriber(sub)
		<-stream.Context().Done()
		return nil
	}

	globalAdded := make([]server.ClientMessage, 0, len(msgs))
	localAdded := make([]server.ClientMessage, 0, len(msgs))

	defer func() {
		l.svcCtx.Hub.RemoveSubscriber(sub)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if len(globalAdded) > 0 {
			if err := l.svcCtx.ItickManager.RemoveGlobalSubscriptions(ctx, globalAdded); err != nil {
				logx.Errorf("remove global subscriptions failed, err=%v", err)
			}
		}
	}()

	if err := l.svcCtx.ItickManager.AddGlobalSubscriptions(stream.Context(), msgs); err != nil {
		return err
	}
	globalAdded = append(globalAdded, msgs...)

	for _, msg := range msgs {
		if err := l.svcCtx.Hub.Subscribe(sub, msg); err != nil {
			if len(localAdded) > 0 {
				for _, added := range localAdded {
					_ = l.svcCtx.Hub.Unsubscribe(sub, added)
				}
			}
			return err
		}
		localAdded = append(localAdded, msg)
	}

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

func normalizeUniqueClientMessages(in *itick.SubscribeRequest) []server.ClientMessage {
	if in == nil || len(in.Topics) == 0 {
		return nil
	}
	uniq := make(map[string]server.ClientMessage, len(in.Topics))
	for _, topic := range in.Topics {
		if topic == nil {
			continue
		}
		msg := server.NormalizeClientMessage(server.ClientMessage{
			Topic:        server.Topic(topic.Topic),
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		})
		if msg.Topic == "" || msg.CategoryCode == "" || msg.Symbol == "" || msg.Market == "" {
			continue
		}
		if msg.Topic == server.TopicKline && msg.Interval == "" {
			continue
		}
		uniq[server.BuildTopicKey(msg)] = msg
	}
	out := make([]server.ClientMessage, 0, len(uniq))
	for _, msg := range uniq {
		out = append(out, msg)
	}
	return out
}

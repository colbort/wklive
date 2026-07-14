package logic

import (
	"context"
	"encoding/json"
	"time"

	"wklive/proto/itick"
	"wklive/services/itick/internal/market/cache"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/svc"

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
	msgs := normalizeUniqueClientMessages(in)
	if len(msgs) == 0 {
		<-stream.Context().Done()
		return nil
	}

	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case <-timer.C:
			if err := l.pushCachedMarketData(stream.Context(), msgs, stream); err != nil {
				return err
			}
			timer.Reset(5 * time.Second)
		}
	}
}

func (l *SubscribeStreamLogic) pushCachedMarketData(
	ctx context.Context,
	msgs []types.ClientMessage,
	stream itick.ItickApp_SubscribeStreamServer,
) error {
	items, err := l.svcCtx.MarketDataCache.ReadMany(ctx, msgs)
	if err != nil {
		return err
	}
	for _, item := range items {
		payload, err := json.Marshal(item.Payload)
		if err != nil {
			continue
		}
		if err := stream.Send(&itick.PushReply{
			Topic:        string(item.Message.Topic),
			CategoryCode: item.Message.CategoryCode,
			Market:       item.Message.Market,
			Symbol:       item.Message.Symbol,
			Interval:     item.Message.Interval,
			Payload:      payload,
		}); err != nil {
			return err
		}
	}
	return nil
}

func normalizeUniqueClientMessages(in *itick.SubscribeRequest) []types.ClientMessage {
	if in == nil || len(in.Topics) == 0 {
		return nil
	}
	uniq := make(map[string]types.ClientMessage, len(in.Topics))
	for _, topic := range in.Topics {
		if topic == nil {
			continue
		}
		msg := cache.NormalizeClientMessage(types.ClientMessage{
			Topic:        types.Topic(topic.Topic),
			CategoryCode: topic.CategoryCode,
			Symbol:       topic.Symbol,
			Market:       topic.Market,
			Interval:     topic.Interval,
		})
		if msg.Topic == "" || msg.CategoryCode == "" || msg.Symbol == "" || msg.Market == "" {
			continue
		}
		if msg.Topic == types.TopicKline && msg.Interval == "" {
			continue
		}
		uniq[cache.BuildTopicKey(msg)] = msg
	}
	out := make([]types.ClientMessage, 0, len(uniq))
	for _, msg := range uniq {
		out = append(out, msg)
	}
	return out
}

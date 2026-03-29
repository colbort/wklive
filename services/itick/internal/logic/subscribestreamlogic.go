package logic

import (
	"context"
	"encoding/json"

	"wklive/proto/itick"
	"wklive/services/itick/internal/socket/server"
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
	sub := l.svcCtx.Hub.NewSubscriber(256)
	defer l.svcCtx.Hub.RemoveSubscriber(sub)

	for _, topic := range in.Topics {
		msg := server.ClientMessage{
			Topic:    server.Topic(topic.Topic),
			Market:   topic.Market,
			Symbol:   topic.Symbol,
			Region:   topic.Region,
			Interval: topic.Interval,
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

			if err := stream.Send(&itick.PushReply{
				Topic:    string(pushMsg.Topic),
				Market:   pushMsg.Market,
				Symbol:   pushMsg.Symbol,
				Region:   pushMsg.Region,
				Interval: pushMsg.Interval,
				Payload:  payloadBytes,
			}); err != nil {
				return err
			}
		}
	}
}

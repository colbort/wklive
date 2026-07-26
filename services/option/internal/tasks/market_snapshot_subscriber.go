package tasks

import (
	"context"

	market "wklive/common/market"
	mq "wklive/common/mq/kafka"
	optionlogic "wklive/services/option/internal/logic/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartMarketSnapshotSubscriber(ctx context.Context, svcCtx *svc.ServiceContext) {
	go func() {
		err := svcCtx.MarketSnapshotSubscriber.Subscribe(ctx, market.AuthoritativeSnapshotTopic, func(ctx context.Context, msg mq.Message) error {
			var event market.AuthoritativeSnapshotEvent
			if err := mq.Decode(msg, &event); err != nil {
				return err
			}
			return optionlogic.NewSyncMarketQuoteLogic(ctx, svcCtx).SyncAuthoritativeSnapshot(event)
		})
		if err != nil && ctx.Err() == nil {
			logx.Errorf("option market snapshot subscriber stopped: %v", err)
		}
	}()
}

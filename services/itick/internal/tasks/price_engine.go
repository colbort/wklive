package tasks

import (
	"context"
	"time"

	"wklive/services/itick/internal/priceengine"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartPriceEngine(ctx context.Context, engine *priceengine.Engine) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			if err := engine.RunOnce(ctx, time.Now().UnixMilli()); err != nil && ctx.Err() == nil {
				logx.Errorf("price engine evaluation failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

package tasks

import (
	"context"
	"errors"
	"time"

	"wklive/services/itick/internal/priceengine"

	"github.com/zeromicro/go-zero/core/logx"
)

func StartPriceEngine(ctx context.Context, engine *priceengine.Engine) {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		var lastUnavailableLog time.Time
		inputUnavailable := false
		for {
			err := engine.RunOnce(ctx, time.Now().UnixMilli())
			if err != nil && ctx.Err() == nil {
				if errors.Is(err, priceengine.ErrInputUnavailable) {
					now := time.Now()
					if lastUnavailableLog.IsZero() || now.Sub(lastUnavailableLog) >= 30*time.Second {
						logx.Infof("price engine waiting for formula input: %v", err)
						lastUnavailableLog = now
					}
					inputUnavailable = true
				} else {
					logx.Errorf("price engine evaluation failed: %v", err)
				}
			} else if err == nil && inputUnavailable {
				logx.Infof("price engine formula input recovered")
				inputUnavailable = false
				lastUnavailableLog = time.Time{}
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

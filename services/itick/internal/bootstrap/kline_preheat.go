package bootstrap

import (
	"context"
	"fmt"
	"time"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type KlinePreheatItem struct {
	CategoryCode string
	Interval     string
}

var defaultKlineCategories = []string{
	"forex",   // 外汇
	"crypto",  // 加密货币
	"stock",   // 股票
	"future",  // 期货
	"indices", // 指数
	"fund",    // 基金
}

var defaultKlineIntervals = []string{
	"1m",
	"5m",
	"15m",
	"30m",
	"1h",
	"1d",
	"1w",
	"1mo",
}

func buildDefaultKlinePreheatItems() []KlinePreheatItem {
	items := make([]KlinePreheatItem, 0, len(defaultKlineCategories)*len(defaultKlineIntervals))
	for _, category := range defaultKlineCategories {
		for _, interval := range defaultKlineIntervals {
			items = append(items, KlinePreheatItem{
				CategoryCode: category,
				Interval:     interval,
			})
		}
	}
	return items
}

func PreheatCoinKlineModels(factory *models.CoinKlineModelFactory) error {
	if factory == nil {
		return fmt.Errorf("coin kline model factory is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	items := buildDefaultKlinePreheatItems()

	for _, item := range items {
		if err := factory.WarmupAndEnsureIndexes(ctx, item.CategoryCode, item.Interval); err != nil {
			return fmt.Errorf("preheat failed, categoryCode=%s interval=%s err=%w",
				item.CategoryCode, item.Interval, err)
		}

		logx.Infof("coin kline preheated, categoryCode=%s interval=%s", item.CategoryCode, item.Interval)
	}

	return nil
}

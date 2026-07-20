package cache

import cache "wklive/common/market"

type MarketDataCache = cache.MarketDataCache
type CachedMarketData = cache.CachedMarketData

var NewMarketDataCache = cache.NewMarketDataCache
var BuildTopicKey = cache.BuildTopicKey
var NormalizeClientMessage = cache.NormalizeClientMessage

func marketDataKey(msg cache.ClientMessage) string {
	msg = cache.NormalizeClientMessage(msg)
	if msg.Topic == cache.TopicKline {
		return "itick:v1:kline:" + msg.CategoryCode + ":" + msg.Market + ":" + msg.Symbol + ":" + msg.Interval
	}
	return "itick:" + string(msg.Topic) + ":" + msg.CategoryCode + ":" + msg.Market + ":" + msg.Symbol
}

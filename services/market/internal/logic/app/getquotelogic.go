package applogic

import (
	"context"
	"wklive/services/market/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/proto/market"
	"wklive/services/market/internal/market/cache"
	"wklive/services/market/internal/market/types"
	"wklive/services/market/internal/svc"

	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteLogic {
	return &GetQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取最新报价
func (l *GetQuoteLogic) GetQuote(in *market.GetQuoteReq) (*market.GetQuoteResp, error) {
	msg := cache.NormalizeClientMessage(types.ClientMessage{
		Topic:        types.TopicQuote,
		CategoryCode: in.CategoryCode,
		Market:       in.Market,
		Symbol:       in.Symbol,
	})
	items, err := l.svcCtx.MarketDataCache.ReadMany(l.ctx, []types.ClientMessage{msg})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, redis.Nil
	}
	data, ok := items[0].Payload.(*types.QuotePayload)
	if !ok || data == nil {
		return nil, redis.Nil
	}

	return &market.GetQuoteResp{
		Base: helper.OkResp(),
		Data: helpers.ToQuotePayloadProto(msg.CategoryCode, msg.Market, msg.Symbol, data),
	}, nil
}

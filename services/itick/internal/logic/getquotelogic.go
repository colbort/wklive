package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/market/cache"
	"wklive/services/itick/internal/market/types"
	"wklive/services/itick/internal/svc"

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
func (l *GetQuoteLogic) GetQuote(in *itick.GetQuoteReq) (*itick.GetQuoteResp, error) {
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

	return &itick.GetQuoteResp{
		Base: helper.OkResp(),
		Data: toQuotePayloadProto(msg.CategoryCode, msg.Market, msg.Symbol, data),
	}, nil
}

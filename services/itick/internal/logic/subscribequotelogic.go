package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubscribeQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubscribeQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubscribeQuoteLogic {
	return &SubscribeQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 订阅报价
func (l *SubscribeQuoteLogic) SubscribeQuote(in *itick.SubscribeQuoteReq) (*itick.SubscribeQuoteResp, error) {
	// todo: add your logic here and delete this line

	return &itick.SubscribeQuoteResp{}, nil
}

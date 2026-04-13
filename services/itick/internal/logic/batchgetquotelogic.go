package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetQuoteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetQuoteLogic {
	return &BatchGetQuoteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 批量获取最新报价
func (l *BatchGetQuoteLogic) BatchGetQuote(in *itick.BatchGetQuoteReq) (*itick.BatchGetQuoteResp, error) {
	result, err := l.svcCtx.ItickQuoteModel.FindQuotes(l.ctx, in.Data)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.Quote, 0)
	for _, quote := range result {
		data = append(data, toQuoteProto(quote))
	}

	return &itick.BatchGetQuoteResp{
		Base: helper.OkResp(),
		Data: data,
	}, nil
}

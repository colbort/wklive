package logic

import (
	"context"

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
	// todo: add your logic here and delete this line

	return &itick.BatchGetQuoteResp{}, nil
}

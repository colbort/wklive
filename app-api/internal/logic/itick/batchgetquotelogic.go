// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetQuoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetQuoteLogic {
	return &BatchGetQuoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetQuoteLogic) BatchGetQuote(req *types.BatchGetQuoteReq) (resp *types.BatchGetQuoteResp, err error) {
	return logicutil.Proxy[types.BatchGetQuoteResp](l.ctx, req, l.svcCtx.ItickCli.BatchGetQuote)
}

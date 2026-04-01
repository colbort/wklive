// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetQuoteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetQuoteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQuoteLogic {
	return &GetQuoteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQuoteLogic) GetQuote(req *types.GetQuoteReq) (resp *types.GetQuoteResp, err error) {
	// todo: add your logic here and delete this line

	return
}

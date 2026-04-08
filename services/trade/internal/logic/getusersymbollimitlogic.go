package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSymbolLimitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSymbolLimitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSymbolLimitLogic {
	return &GetUserSymbolLimitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserSymbolLimitLogic) GetUserSymbolLimit(in *trade.GetUserSymbolLimitReq) (*trade.GetUserSymbolLimitResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetUserSymbolLimitResp{}, nil
}

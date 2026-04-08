package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListLogic {
	return &GetSymbolListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对列表
func (l *GetSymbolListLogic) GetSymbolList(in *trade.GetSymbolListReq) (*trade.GetSymbolListResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetSymbolListResp{}, nil
}

package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolDetailAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolDetailAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailAdminLogic {
	return &GetSymbolDetailAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取交易对详情
func (l *GetSymbolDetailAdminLogic) GetSymbolDetailAdmin(in *trade.GetSymbolDetailAdminReq) (*trade.GetSymbolDetailAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetSymbolDetailAdminResp{}, nil
}

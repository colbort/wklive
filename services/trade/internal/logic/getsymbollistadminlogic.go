package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolListAdminLogic {
	return &GetSymbolListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台交易对列表
func (l *GetSymbolListAdminLogic) GetSymbolListAdmin(in *trade.GetSymbolListAdminReq) (*trade.GetSymbolListAdminResp, error) {
	// todo: add your logic here and delete this line

	return &trade.GetSymbolListAdminResp{}, nil
}

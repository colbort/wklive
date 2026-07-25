package tradelogic

import (
	"context"

	"wklive/proto/trade"
	adminlogic "wklive/services/trade/internal/logic/admin"
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

// 内部服务查询可配置的交易对，不依赖管理后台登录上下文。
func (l *GetSymbolListLogic) GetSymbolList(in *trade.GetSymbolListAdminReq) (*trade.GetSymbolListAdminResp, error) {
	if in == nil {
		in = &trade.GetSymbolListAdminReq{}
	}
	return adminlogic.NewGetSymbolListAdminLogic(l.ctx, l.svcCtx).GetSymbolListAdmin(in)
}

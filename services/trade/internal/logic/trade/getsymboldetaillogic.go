package tradelogic

import (
	"context"
	"fmt"

	"wklive/proto/trade"
	applogic "wklive/services/trade/internal/logic/app"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolDetailLogic {
	return &GetSymbolDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 内部服务获取指定交易对配置，不依赖用户端登录上下文。
func (l *GetSymbolDetailLogic) GetSymbolDetail(in *trade.GetSymbolDetailReq) (*trade.GetSymbolDetailResp, error) {
	if in.TenantId <= 0 {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if in.SymbolId <= 0 {
		return nil, fmt.Errorf("symbol_id is required")
	}
	return applogic.QuerySymbolDetail(l.ctx, l.svcCtx, in.TenantId, in.SymbolId)
}

package adminlogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateSymbolConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateSymbolConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateSymbolConfigLogic {
	return &UpdateSymbolConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateSymbolConfigLogic) UpdateSymbolConfig(in *liquidity.SaveSymbolConfigReq) (*liquidity.SymbolConfigResp, error) {
	current, err := l.svcCtx.SymbolConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if current.TenantId != in.TenantId {
		return nil, fmt.Errorf("symbol config not found")
	}
	if current.Version != in.Version {
		return nil, fmt.Errorf("symbol config version conflict")
	}
	if current.Status == int64(liquidity.SymbolLiquidityStatus_SYMBOL_LIQUIDITY_STATUS_RUNNING) {
		return nil, fmt.Errorf("pause liquidity before updating configuration")
	}
	row, err := buildSymbolConfig(l.ctx, l.svcCtx, in, current)
	if err != nil {
		return nil, err
	}
	row.Version++
	if err := l.svcCtx.SymbolConfigModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.SymbolConfigResp{Base: helper.OkResp(), Data: helpers.SymbolConfigToProto(row)}, nil
}

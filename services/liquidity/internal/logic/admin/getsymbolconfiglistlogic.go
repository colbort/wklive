package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSymbolConfigListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSymbolConfigListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSymbolConfigListLogic {
	return &GetSymbolConfigListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSymbolConfigListLogic) GetSymbolConfigList(in *liquidity.GetSymbolConfigListReq) (*liquidity.GetSymbolConfigListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := l.svcCtx.SymbolConfigModel.FindPage(l.ctx, models.LiquiditySymbolConfigPageFilter{
		TenantId: in.TenantId, SymbolId: in.SymbolId, ProductType: int64(in.ProductType),
		ContractType: int64(in.ContractType), LiquidityMode: int64(in.LiquidityMode),
		Status: int64(in.Status), Keyword: strings.TrimSpace(in.Keyword),
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquiditySymbolConfig, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.SymbolConfigToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquiditySymbolConfig) int64 { return row.Id })
	return &liquidity.GetSymbolConfigListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

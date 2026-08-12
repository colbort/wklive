package adminlogic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
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
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := l.svcCtx.SymbolConfigModel.FindPage(l.ctx, models.LiquiditySymbolConfigPageFilter{
		SymbolId: in.SymbolId, ProductType: int64(in.ProductType),
		ContractType: int64(in.ContractType), LiquidityMode: int64(in.LiquidityMode),
		Status: int64(in.Status), Keyword: strings.TrimSpace(in.Keyword),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquiditySymbolConfig, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.SymbolConfigToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquiditySymbolConfig) int64 { return row.Id })
	return &liquidity.GetSymbolConfigListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

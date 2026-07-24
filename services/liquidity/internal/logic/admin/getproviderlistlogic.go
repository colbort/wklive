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

type GetProviderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetProviderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetProviderListLogic {
	return &GetProviderListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetProviderListLogic) GetProviderList(in *liquidity.GetProviderListReq) (*liquidity.GetProviderListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := l.svcCtx.ProviderModel.FindPage(l.ctx, models.LiquidityProviderPageFilter{
		TenantId: in.TenantId, ProviderType: int64(in.ProviderType),
		Status: int64(in.Status), Keyword: strings.TrimSpace(in.Keyword),
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityProvider, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ProviderToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityProvider) int64 { return row.Id })
	return &liquidity.GetProviderListResp{
		Base: helper.OkResp(), Data: data,
		Page: pageMeta(len(rows), total, next),
	}, nil
}

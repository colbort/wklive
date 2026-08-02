package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPortfolioRiskConfigsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPortfolioRiskConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPortfolioRiskConfigsLogic {
	return &ListPortfolioRiskConfigsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询组合保证金参数版本
func (l *ListPortfolioRiskConfigsLogic) ListPortfolioRiskConfigs(in *option.ListPortfolioRiskConfigsReq) (*option.ListPortfolioRiskConfigsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListPortfolioRiskConfigsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionPortfolioRiskConfigModel.FindPage(
		l.ctx, models.OptionPortfolioRiskConfigPageFilter{
			TenantId: tenantId, SettleCoin: in.SettleCoin, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionPortfolioRiskConfig, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		data = append(data, helpers.ToPortfolioRiskConfigProto(item))
	}
	return &option.ListPortfolioRiskConfigsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}

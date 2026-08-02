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

type ListTradeCorrectionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTradeCorrectionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTradeCorrectionsLogic {
	return &ListTradeCorrectionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询异常成交更正案件
func (l *ListTradeCorrectionsLogic) ListTradeCorrections(in *option.ListTradeCorrectionsReq) (*option.ListTradeCorrectionsResp, error) {
	tenantId, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.ListTradeCorrectionsResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionTradeCorrectionModel.FindPage(
		l.ctx, models.OptionTradeCorrectionPageFilter{
			TenantId: tenantId, TradeId: in.TradeId,
			ContractId: in.ContractId, Status: int64(in.Status),
		}, cursor, limit,
	)
	if err != nil {
		return nil, err
	}
	data := make([]*option.OptionTradeCorrection, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		legs, findErr := l.svcCtx.OptionTradeCorrectionLegModel.FindByCorrection(
			l.ctx, item.TenantId, item.Id,
		)
		if findErr != nil {
			return nil, findErr
		}
		data = append(data, helpers.ToTradeCorrectionProto(item, legs))
		lastID = item.Id
	}
	return &option.ListTradeCorrectionsResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		Data: data, Total: total,
	}, nil
}

package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/common/pageutil"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReconciliationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReconciliationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconciliationListLogic {
	return &ReconciliationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取每日账实对账快照
func (l *ReconciliationListLogic) ReconciliationList(in *staking.ReconciliationListReq) (*staking.ReconciliationListResp, error) {
	tenantID, base, err := helpers.AdminTenantReadScopeResp(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if base != nil {
		return &staking.ReconciliationListResp{Page: base}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	items, total, err := l.svcCtx.StakeReconciliationModel.FindAdminPage(l.ctx, models.StakeReconciliationPageFilter{
		TenantId: tenantID, ReconciliationDate: in.GetReconciliationDate(),
		CoinSymbol: strings.TrimSpace(in.GetCoinSymbol()), Status: in.GetStatus(),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	resp := &staking.ReconciliationListResp{Page: helper.OkResp(), Data: make([]*staking.StakeReconciliation, 0, len(items))}
	lastID := int64(0)
	for _, item := range items {
		resp.Data = append(resp.Data, helpers.ReconciliationToProto(item))
		lastID = item.Id
	}
	resp.Page = pageutil.Base(cursor, limit, len(items), total, lastID)
	return resp, nil
}

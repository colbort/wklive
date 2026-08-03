package adminlogic

import (
	"context"
	"errors"
	"sort"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserTradeControlsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserTradeControlsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserTradeControlsLogic {
	return &ListUserTradeControlsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页获取统一用户交易控制
func (l *ListUserTradeControlsLogic) ListUserTradeControls(in *trade.ListUserTradeControlsReq) (*trade.ListUserTradeControlsResp, error) {
	if in.ScopeType < 0 || in.ScopeType > 2 {
		return nil, errors.New("invalid control scope type")
	}
	tenantID, allowed, forbidden, err := utils.ResolveAdminTenantReadScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &trade.ListUserTradeControlsResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	cursor, limit := pageutil.Input(in.Page)
	filter := models.UserTradeControlFilter{
		TenantId: tenantID, UserId: in.UserId, ProductType: int64(in.ProductType),
		ContractType: int64(in.ContractType), SymbolId: in.SymbolId, Enabled: int64(in.Enabled),
	}
	resp := &trade.ListUserTradeControlsResp{}
	last := int64(0)
	switch in.ScopeType {
	case 1:
		rows, total, findErr := l.svcCtx.RiskUserTradeLimitModel.FindControlPage(l.ctx, filter, cursor, limit)
		if findErr != nil {
			return nil, findErr
		}
		for _, row := range rows {
			resp.Data = append(resp.Data, &trade.UserTradeControlEntry{ScopeType: 1, ProductLimit: helpers.RiskUserTradeLimitToProto(row)})
			last = row.Id
		}
		resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	case 2:
		rows, total, findErr := l.svcCtx.RiskUserSymbolLimitModel.FindControlPage(l.ctx, filter, cursor, limit)
		if findErr != nil {
			return nil, findErr
		}
		for _, row := range rows {
			resp.Data = append(resp.Data, &trade.UserTradeControlEntry{ScopeType: 2, SymbolLimit: helpers.RiskUserSymbolLimitToProto(row)})
			last = row.Id
		}
		resp.Base = pageutil.Base(cursor, limit, len(rows), total, last)
	default:
		return l.listAllScopes(filter, cursor, limit)
	}
	return resp, nil
}

type keyedUserTradeControl struct {
	key   int64
	entry *trade.UserTradeControlEntry
}

func (l *ListUserTradeControlsLogic) listAllScopes(filter models.UserTradeControlFilter, cursor, limit int64) (*trade.ListUserTradeControlsResp, error) {
	productCursor, symbolCursor := int64(0), int64(0)
	if cursor > 0 {
		productCursor = (cursor + 1) / 2
		symbolCursor = cursor / 2
	}
	var products []*models.TRiskUserTradeLimit
	var productTotal int64
	var err error
	if filter.SymbolId == 0 {
		products, productTotal, err = l.svcCtx.RiskUserTradeLimitModel.FindControlPage(l.ctx, filter, productCursor, limit)
		if err != nil {
			return nil, err
		}
	}
	var symbols []*models.TRiskUserSymbolLimit
	var symbolTotal int64
	if filter.ProductType == 0 && filter.ContractType == 0 {
		symbols, symbolTotal, err = l.svcCtx.RiskUserSymbolLimitModel.FindControlPage(l.ctx, filter, symbolCursor, limit)
		if err != nil {
			return nil, err
		}
	}
	items := make([]keyedUserTradeControl, 0, len(products)+len(symbols))
	for _, row := range products {
		items = append(items, keyedUserTradeControl{key: row.Id * 2, entry: &trade.UserTradeControlEntry{ScopeType: 1, ProductLimit: helpers.RiskUserTradeLimitToProto(row)}})
	}
	for _, row := range symbols {
		items = append(items, keyedUserTradeControl{key: row.Id*2 + 1, entry: &trade.UserTradeControlEntry{ScopeType: 2, SymbolLimit: helpers.RiskUserSymbolLimitToProto(row)}})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].key > items[j].key })
	if int64(len(items)) > limit {
		items = items[:limit]
	}
	resp := &trade.ListUserTradeControlsResp{}
	last := int64(0)
	for _, item := range items {
		resp.Data = append(resp.Data, item.entry)
		last = item.key
	}
	resp.Base = pageutil.Base(cursor, limit, len(items), productTotal+symbolTotal, last)
	return resp, nil
}

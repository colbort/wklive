package adminlogic

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"
)

func pageMeta(rows int, total, next int64) *liquidity.PageMeta {
	return &liquidity.PageMeta{NextCursor: next, Total: total, HasMore: rows > 0 && next > 0 && int64(rows) < total}
}

func nextID[T any](rows []*T, id func(*T) int64) int64 {
	if len(rows) == 0 {
		return 0
	}
	return id(rows[len(rows)-1])
}

func listQuoteCycles(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetQuoteCycleListReq) (*liquidity.GetQuoteCycleListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.QuoteCycleModel.FindPage(ctx, models.LiquidityQuoteCyclePageFilter{
		TenantId: in.TenantId, ConfigId: in.ConfigId, SymbolId: in.SymbolId,
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityQuoteCycle, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.QuoteCycleToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityQuoteCycle) int64 { return row.Id })
	return &liquidity.GetQuoteCycleListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listQuoteOrders(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetQuoteOrderListReq) (*liquidity.GetQuoteOrderListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.QuoteOrderModel.FindPage(ctx, models.LiquidityQuoteOrderPageFilter{
		TenantId: in.TenantId, ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		SymbolId: in.SymbolId, Side: int64(in.Side), Status: int64(in.Status),
		Keyword: strings.TrimSpace(in.Keyword), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityQuoteOrder, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.QuoteOrderToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityQuoteOrder) int64 { return row.Id })
	return &liquidity.GetQuoteOrderListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listExternalOrders(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetExternalOrderListReq) (*liquidity.GetExternalOrderListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.ExternalOrderModel.FindPage(ctx, models.LiquidityExternalOrderPageFilter{
		TenantId: in.TenantId, ProviderId: in.ProviderId, ConfigId: in.ConfigId,
		SymbolId: in.SymbolId, Purpose: int64(in.Purpose), Side: int64(in.Side),
		Status: int64(in.Status), Keyword: strings.TrimSpace(in.Keyword),
		TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityExternalOrder, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ExternalOrderToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityExternalOrder) int64 { return row.Id })
	return &liquidity.GetExternalOrderListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listExternalFills(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetExternalFillListReq) (*liquidity.GetExternalFillListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.ExternalFillModel.FindPage(ctx, models.LiquidityExternalFillPageFilter{
		TenantId: in.TenantId, ProviderId: in.ProviderId, ExternalOrderId: in.ExternalOrderId,
		SettlementStatus: int64(in.SettlementStatus), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityExternalFill, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ExternalFillToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityExternalFill) int64 { return row.Id })
	return &liquidity.GetExternalFillListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listHedgeTasks(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetHedgeTaskListReq) (*liquidity.GetHedgeTaskListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.HedgeTaskModel.FindPage(ctx, models.LiquidityHedgeTaskPageFilter{
		TenantId: in.TenantId, ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityHedgeTask, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.HedgeTaskToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityHedgeTask) int64 { return row.Id })
	return &liquidity.GetHedgeTaskListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listInventories(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetInventorySnapshotListReq) (*liquidity.GetInventorySnapshotListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.InventorySnapshotModel.FindPage(ctx, models.LiquidityInventorySnapshotPageFilter{
		TenantId: in.TenantId, ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		Source: int64(in.Source), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityInventorySnapshot, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.InventoryToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityInventorySnapshot) int64 { return row.Id })
	return &liquidity.GetInventorySnapshotListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listRiskEvents(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetRiskEventListReq) (*liquidity.GetRiskEventListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.RiskEventModel.FindPage(ctx, models.LiquidityRiskEventPageFilter{
		TenantId: in.TenantId, ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		RiskType: strings.TrimSpace(in.RiskType), RiskLevel: int64(in.RiskLevel),
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityRiskEvent, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.RiskEventToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityRiskEvent) int64 { return row.Id })
	return &liquidity.GetRiskEventListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listReconcileBatches(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetReconcileBatchListReq) (*liquidity.GetReconcileBatchListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.ReconcileBatchModel.FindPage(ctx, models.LiquidityReconcileBatchPageFilter{
		TenantId: in.TenantId, ProviderId: in.ProviderId, ReconcileType: int64(in.ReconcileType),
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityReconcileBatch, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ReconcileBatchToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityReconcileBatch) int64 { return row.Id })
	return &liquidity.GetReconcileBatchListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

func listReconcileDetails(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetReconcileDetailListReq) (*liquidity.GetReconcileDetailListResp, error) {
	if err := helpers.RequireTenant(in.TenantId); err != nil {
		return nil, err
	}
	rows, total, err := svcCtx.ReconcileDetailModel.FindPage(ctx, models.LiquidityReconcileDetailPageFilter{
		TenantId: in.TenantId, BatchId: in.BatchId,
		DifferenceType: int64(in.DifferenceType), Status: int64(in.Status),
	}, in.Cursor, int64(in.Limit))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityReconcileDetail, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ReconcileDetailToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityReconcileDetail) int64 { return row.Id })
	return &liquidity.GetReconcileDetailListResp{Base: helper.OkResp(), Data: data, Page: pageMeta(len(rows), total, next)}, nil
}

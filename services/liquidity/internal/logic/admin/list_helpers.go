package adminlogic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"
)

func nextID[T any](rows []*T, id func(*T) int64) int64 {
	if len(rows) == 0 {
		return 0
	}
	return id(rows[len(rows)-1])
}

func listQuoteCycles(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetQuoteCycleListReq) (*liquidity.GetQuoteCycleListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.QuoteCycleModel.FindPage(ctx, models.LiquidityQuoteCyclePageFilter{
		ConfigId: in.ConfigId, SymbolId: in.SymbolId,
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit, pageutil.Count(in.Page))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityQuoteCycle, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.QuoteCycleToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityQuoteCycle) int64 { return row.Id })
	return &liquidity.GetQuoteCycleListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listQuoteOrders(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetQuoteOrderListReq) (*liquidity.GetQuoteOrderListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.QuoteOrderModel.FindPage(ctx, models.LiquidityQuoteOrderPageFilter{
		ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		SymbolId: in.SymbolId, Side: int64(in.Side), Status: int64(in.Status),
		Keyword: strings.TrimSpace(in.Keyword), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit, pageutil.Count(in.Page))
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityQuoteOrder, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.QuoteOrderToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityQuoteOrder) int64 { return row.Id })
	return &liquidity.GetQuoteOrderListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listExternalOrders(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetExternalOrderListReq) (*liquidity.GetExternalOrderListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.ExternalOrderModel.FindPage(ctx, models.LiquidityExternalOrderPageFilter{
		ProviderId: in.ProviderId, ConfigId: in.ConfigId,
		SymbolId: in.SymbolId, Purpose: int64(in.Purpose), Side: int64(in.Side),
		Status: int64(in.Status), Keyword: strings.TrimSpace(in.Keyword),
		TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityExternalOrder, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ExternalOrderToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityExternalOrder) int64 { return row.Id })
	return &liquidity.GetExternalOrderListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listExternalFills(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetExternalFillListReq) (*liquidity.GetExternalFillListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.ExternalFillModel.FindPage(ctx, models.LiquidityExternalFillPageFilter{
		ProviderId: in.ProviderId, ExternalOrderId: in.ExternalOrderId,
		SettlementStatus: int64(in.SettlementStatus), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityExternalFill, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ExternalFillToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityExternalFill) int64 { return row.Id })
	return &liquidity.GetExternalFillListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listHedgeTasks(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetHedgeTaskListReq) (*liquidity.GetHedgeTaskListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.HedgeTaskModel.FindPage(ctx, models.LiquidityHedgeTaskPageFilter{
		ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityHedgeTask, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.HedgeTaskToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityHedgeTask) int64 { return row.Id })
	return &liquidity.GetHedgeTaskListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listInventories(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetInventorySnapshotListReq) (*liquidity.GetInventorySnapshotListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.InventorySnapshotModel.FindPage(ctx, models.LiquidityInventorySnapshotPageFilter{
		ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		Source: int64(in.Source), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityInventorySnapshot, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.InventoryToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityInventorySnapshot) int64 { return row.Id })
	return &liquidity.GetInventorySnapshotListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listRiskEvents(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetRiskEventListReq) (*liquidity.GetRiskEventListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.RiskEventModel.FindPage(ctx, models.LiquidityRiskEventPageFilter{
		ConfigId: in.ConfigId, ProviderId: in.ProviderId,
		RiskType: strings.TrimSpace(in.RiskType), RiskLevel: int64(in.RiskLevel),
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityRiskEvent, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.RiskEventToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityRiskEvent) int64 { return row.Id })
	return &liquidity.GetRiskEventListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listReconcileBatches(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetReconcileBatchListReq) (*liquidity.GetReconcileBatchListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.ReconcileBatchModel.FindPage(ctx, models.LiquidityReconcileBatchPageFilter{
		ProviderId: in.ProviderId, ReconcileType: int64(in.ReconcileType),
		Status: int64(in.Status), TimeStart: in.StartTime, TimeEnd: in.EndTime,
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityReconcileBatch, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ReconcileBatchToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityReconcileBatch) int64 { return row.Id })
	return &liquidity.GetReconcileBatchListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

func listReconcileDetails(ctx context.Context, svcCtx *svc.ServiceContext, in *liquidity.GetReconcileDetailListReq) (*liquidity.GetReconcileDetailListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := svcCtx.ReconcileDetailModel.FindPage(ctx, models.LiquidityReconcileDetailPageFilter{
		BatchId:        in.BatchId,
		DifferenceType: int64(in.DifferenceType), Status: int64(in.Status),
	}, cursor, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*liquidity.LiquidityReconcileDetail, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.ReconcileDetailToProto(row))
	}
	next := nextID(rows, func(row *models.TLiquidityReconcileDetail) int64 { return row.Id })
	return &liquidity.GetReconcileDetailListResp{Base: pageutil.Base(cursor, limit, len(rows), total, next), Data: data}, nil
}

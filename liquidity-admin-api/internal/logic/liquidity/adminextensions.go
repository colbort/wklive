package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/common"
	pb "wklive/proto/liquidity"
)

func providerUpdate(ctx context.Context, svcCtx *svc.ServiceContext, req *types.UpdateProviderReq) (*types.ProviderDetailResp, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.UpdateProvider(ctx, &pb.UpdateProviderReq{
		Id: req.Id, ProviderName: req.ProviderName, TradeUserId: req.TradeUserId,
		VenueCode: req.VenueCode, Environment: pb.ProviderEnvironment(req.Environment),
		CredentialRef: req.CredentialRef, AccountRef: req.AccountRef,
		RateLimitPerSecond: req.RateLimitPerSecond, Remark: req.Remark,
		Version: req.Version, OperatorId: operatorID,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ProviderDetailResp](out), nil
}

func providerDetail(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ProviderDetailReq) (*types.ProviderDetailResp, error) {
	out, err := svcCtx.LiquidityCli.GetProviderDetail(ctx, &pb.GetProviderDetailReq{Id: req.Id})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ProviderDetailResp](out), nil
}

func cancelAllQuoteOrders(ctx context.Context, svcCtx *svc.ServiceContext, req *types.CancelAllQuoteOrdersReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.CancelAllQuoteOrders(ctx, &pb.SymbolActionReq{
		ConfigId: req.Id, Version: req.Version, OperatorId: operatorID, Reason: req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

func quoteCycleList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.QuoteCycleQuery) (*types.QuoteCycleListResp, error) {
	out, err := svcCtx.LiquidityCli.GetQuoteCycleList(ctx, &pb.GetQuoteCycleListReq{
		ConfigId: req.ConfigId, SymbolId: req.SymbolId, Status: pb.QuoteCycleStatus(req.Status),
		StartTime: req.StartTime, EndTime: req.EndTime, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.QuoteCycleListResp](out), nil
}

func externalFillList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ExternalFillQuery) (*types.ExternalFillListResp, error) {
	out, err := svcCtx.LiquidityCli.GetExternalFillList(ctx, &pb.GetExternalFillListReq{
		ProviderId: req.ProviderId, ExternalOrderId: req.ExternalOrderId,
		SettlementStatus: pb.ExternalFillSettlementStatus(req.SettlementStatus),
		StartTime:        req.StartTime, EndTime: req.EndTime, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ExternalFillListResp](out), nil
}

func cancelExternalOrder(ctx context.Context, svcCtx *svc.ServiceContext, req *types.CancelExternalOrderReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.CancelExternalOrder(ctx, &pb.CancelExternalOrderReq{
		OrderId: req.Id, Version: req.Version, OperatorId: operatorID, Reason: req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

func createManualHedge(ctx context.Context, svcCtx *svc.ServiceContext, req *types.CreateManualHedgeReq) (*types.HedgeTaskResp, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.CreateManualHedge(ctx, &pb.CreateManualHedgeReq{
		ConfigId: req.ConfigId, ProviderId: req.ProviderId, Side: common.Side(req.Side),
		Qty: req.Qty, TargetExposure: req.TargetExposure, OperatorId: operatorID, Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.HedgeTaskResp](out), nil
}

func cancelHedgeTask(ctx context.Context, svcCtx *svc.ServiceContext, req *types.HedgeTaskActionReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.CancelHedgeTask(ctx, &pb.CancelHedgeTaskReq{
		HedgeTaskId: req.Id, Version: req.Version, OperatorId: operatorID, Reason: req.Reason,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

func retryHedgeTask(ctx context.Context, svcCtx *svc.ServiceContext, req *types.HedgeTaskActionReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.RetryHedgeTask(ctx, &pb.RetryHedgeTaskReq{
		HedgeTaskId: req.Id, Version: req.Version, OperatorId: operatorID,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

func inventorySnapshotList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.InventoryQuery) (*types.InventoryListResp, error) {
	out, err := svcCtx.LiquidityCli.GetInventorySnapshotList(ctx, &pb.GetInventorySnapshotListReq{
		ConfigId: req.ConfigId, ProviderId: req.ProviderId, Source: pb.InventorySource(req.Source),
		StartTime: req.StartTime, EndTime: req.EndTime, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.InventoryListResp](out), nil
}

func latestInventory(ctx context.Context, svcCtx *svc.ServiceContext, req *types.LatestInventoryReq) (*types.InventoryResp, error) {
	out, err := svcCtx.LiquidityCli.GetLatestInventory(ctx, &pb.GetLatestInventoryReq{
		ConfigId: req.ConfigId, ProviderId: req.ProviderId, Source: pb.InventorySource(req.Source),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.InventoryResp](out), nil
}

func resolveRiskEvent(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ResolveRiskEventReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.ResolveRiskEvent(ctx, &pb.ResolveRiskEventReq{
		RiskEventId: req.Id, Status: pb.RiskEventStatus(req.Status), OperatorId: operatorID, Resolution: req.Resolution,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

func runReconcile(ctx context.Context, svcCtx *svc.ServiceContext, req *types.RunReconcileReq) (*types.ReconcileResp, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.RunReconcile(ctx, &pb.RunReconcileReq{
		ProviderId: req.ProviderId, ReconcileType: pb.ReconcileType(req.ReconcileType),
		WindowStart: req.WindowStart, WindowEnd: req.WindowEnd, OperatorId: operatorID,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ReconcileResp](out), nil
}

func reconcileDetailList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ReconcileDetailQuery) (*types.ReconcileDetailListResp, error) {
	out, err := svcCtx.LiquidityCli.GetReconcileDetailList(ctx, &pb.GetReconcileDetailListReq{
		BatchId: req.BatchId, DifferenceType: pb.ReconcileDifferenceType(req.DifferenceType),
		Status: pb.ReconcileDifferenceStatus(req.Status), Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ReconcileDetailListResp](out), nil
}

func resolveReconcileDifference(ctx context.Context, svcCtx *svc.ServiceContext, req *types.ResolveReconcileDifferenceReq) (*types.RespBase, error) {
	_, operatorID, err := logicutil.Identity(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.ResolveReconcileDifference(ctx, &pb.ResolveReconcileDifferenceReq{
		DifferenceId: req.Id, Status: pb.ReconcileDifferenceStatus(req.Status),
		OperatorId: operatorID, Resolution: req.Resolution,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

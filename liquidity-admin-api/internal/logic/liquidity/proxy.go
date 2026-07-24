package liquidity

import (
	"context"
	"fmt"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	commonpb "wklive/proto/common"
	pb "wklive/proto/liquidity"
)

func providerList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.PageQuery) (*types.ProviderListResp, error) {
	tenantID, err := logicutil.TenantID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.GetProviderList(ctx, &pb.GetProviderListReq{
		TenantId: tenantID, Status: pb.ProviderStatus(req.Status), Keyword: req.Keyword, Cursor: req.Cursor, Limit: req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ProviderListResp](out), nil
}

func symbolConfigList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.PageQuery) (*types.SymbolConfigListResp, error) {
	tenantID, err := logicutil.TenantID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := svcCtx.LiquidityCli.GetSymbolConfigList(ctx, &pb.GetSymbolConfigListReq{
		TenantId: tenantID, Status: pb.SymbolLiquidityStatus(req.Status), Keyword: req.Keyword, Cursor: req.Cursor, Limit: req.Limit,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.SymbolConfigListResp](out), nil
}

func orderList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.OrderQuery, external bool) (*types.OrderListResp, error) {
	tenantID, err := logicutil.TenantID(ctx)
	if err != nil {
		return nil, err
	}
	if external {
		out, callErr := svcCtx.LiquidityCli.GetExternalOrderList(ctx, &pb.GetExternalOrderListReq{
			TenantId: tenantID, ProviderId: req.ProviderId, ConfigId: req.ConfigId, SymbolId: req.SymbolId,
			Side: commonpb.Side(req.Side), Status: pb.ExternalOrderStatus(req.Status), Keyword: req.Keyword, Cursor: req.Cursor, Limit: listLimit(req.Limit),
		})
		if callErr != nil {
			return nil, callErr
		}
		return logicutil.Convert[types.OrderListResp](out), nil
	}
	out, err := svcCtx.LiquidityCli.GetQuoteOrderList(ctx, &pb.GetQuoteOrderListReq{
		TenantId: tenantID, ProviderId: req.ProviderId, ConfigId: req.ConfigId, SymbolId: req.SymbolId,
		Side: commonpb.Side(req.Side), Status: pb.QuoteOrderStatus(req.Status), Keyword: req.Keyword, Cursor: req.Cursor, Limit: listLimit(req.Limit),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.OrderListResp](out), nil
}

func listLimit(limit int32) int32 {
	if limit <= 0 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func invalidAction(action string) error {
	return fmt.Errorf("unsupported symbol action: %s", action)
}

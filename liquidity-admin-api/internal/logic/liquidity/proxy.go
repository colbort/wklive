package liquidity

import (
	"context"
	"fmt"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	"wklive/proto/common"
	pb "wklive/proto/liquidity"
)

func providerList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.PageQuery) (*types.ProviderListResp, error) {
	out, err := svcCtx.LiquidityCli.GetProviderList(ctx, &pb.GetProviderListReq{
		Status: pb.ProviderStatus(req.Status), Keyword: req.Keyword, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.ProviderListResp](out), nil
}

func symbolConfigList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.PageQuery) (*types.SymbolConfigListResp, error) {
	out, err := svcCtx.LiquidityCli.GetSymbolConfigList(ctx, &pb.GetSymbolConfigListReq{
		Status: pb.SymbolLiquidityStatus(req.Status), Keyword: req.Keyword, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.SymbolConfigListResp](out), nil
}

func orderList(ctx context.Context, svcCtx *svc.ServiceContext, req *types.OrderQuery, external bool) (*types.OrderListResp, error) {
	if external {
		out, callErr := svcCtx.LiquidityCli.GetExternalOrderList(ctx, &pb.GetExternalOrderListReq{
			ProviderId: req.ProviderId, ConfigId: req.ConfigId, SymbolId: req.SymbolId,
			Side: common.Side(req.Side), Status: pb.ExternalOrderStatus(req.Status), Keyword: req.Keyword, Page: protoPage(req.Cursor, req.Limit, req.Count),
		})
		if callErr != nil {
			return nil, callErr
		}
		return logicutil.Convert[types.OrderListResp](out), nil
	}
	out, err := svcCtx.LiquidityCli.GetQuoteOrderList(ctx, &pb.GetQuoteOrderListReq{
		ProviderId: req.ProviderId, ConfigId: req.ConfigId, SymbolId: req.SymbolId,
		Side: common.Side(req.Side), Status: pb.QuoteOrderStatus(req.Status), Keyword: req.Keyword, Page: protoPage(req.Cursor, req.Limit, req.Count),
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.OrderListResp](out), nil
}

func protoPage(cursor int64, limit int64, count int64) *common.PageReq {
	return &common.PageReq{
		Cursor: cursor,
		Limit:  listLimit(limit),
		Count:  count,
	}
}

func listLimit(limit int64) int64 {
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

package tradelogic

import (
	"context"
	"fmt"
	"strconv"

	"wklive/common/utils"
	"wklive/services/trade/internal/svc"

	"google.golang.org/grpc/metadata"
)

func liquidityUserContext(ctx context.Context, tenantID, userID int64) (context.Context, error) {
	if tenantID <= 0 || userID <= 0 {
		return nil, fmt.Errorf("valid trade_user_id and tenant are required")
	}
	md, _ := metadata.FromIncomingContext(ctx)
	md = md.Copy()
	md.Set(utils.CtxKeyTenantId, strconv.FormatInt(tenantID, 10))
	md.Set(utils.CtxKeyUid, strconv.FormatInt(userID, 10))
	return metadata.NewIncomingContext(ctx, md), nil
}

func symbolTenant(ctx context.Context, svcCtx *svc.ServiceContext, symbolID int64) (int64, error) {
	if symbolID <= 0 {
		return 0, fmt.Errorf("symbol_id is required")
	}
	symbol, err := svcCtx.TradeSymbolModel.FindOne(ctx, symbolID)
	if err != nil {
		return 0, err
	}
	return symbol.TenantId, nil
}

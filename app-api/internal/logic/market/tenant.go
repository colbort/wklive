package market

import (
	"context"
	"fmt"

	"wklive/app-api/internal/svc"
	"wklive/common/utils"
	"wklive/proto/system"
)

// resolveTenantContext resolves public market requests by tenant code. For
// authenticated requests, the tenant carried by the signed token takes
// precedence and is forwarded directly to market-rpc.
func resolveTenantContext(ctx context.Context, svcCtx *svc.ServiceContext) (context.Context, error) {
	if tenantID, err := utils.GetTrustedTenantIdFromCtx(ctx); err == nil && tenantID > 0 {
		return context.WithValue(ctx, utils.CtxKeyTenantId, tenantID), nil
	}

	tenantCode, err := utils.GetTenantCodeFromCtx(ctx)
	if err != nil || tenantCode == "" {
		return nil, fmt.Errorf("tenant code is required")
	}

	resp, err := svcCtx.SystemCli.SysTenantDetail(ctx, &system.SysTenantDetailReq{
		TenantCode: &tenantCode,
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil || resp.GetData().GetId() <= 0 {
		return nil, fmt.Errorf("tenant not found")
	}

	return context.WithValue(ctx, utils.CtxKeyTenantId, resp.GetData().GetId()), nil
}

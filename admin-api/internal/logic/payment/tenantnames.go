package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/proto/common"
	"wklive/proto/system"
)

func loadTenantNames(ctx context.Context, svcCtx *svc.ServiceContext, tenantIDs []int64) map[int64]string {
	names := make(map[int64]string)
	if len(tenantIDs) == 0 {
		return names
	}
	resp, err := svcCtx.SystemCli.SysTenantList(ctx, &system.SysTenantListReq{
		Page: &common.PageReq{Limit: int64(len(tenantIDs))},
		Ids:  tenantIDs,
	})
	if err != nil || resp == nil {
		return names
	}
	for _, tenant := range resp.Data {
		names[tenant.Id] = tenant.TenantName
	}
	return names
}

func uniqueTenantIDs[T any](items []T, tenantID func(T) int64) []int64 {
	ids := make([]int64, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for _, item := range items {
		id := tenantID(item)
		if id <= 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

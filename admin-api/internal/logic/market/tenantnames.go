package market

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/proto/common"
	"wklive/proto/system"
)

func loadTenantNames(ctx context.Context, svcCtx *svc.ServiceContext, ids []int64) (map[int64]string, error) {
	unique := make(map[int64]struct{}, len(ids))
	filtered := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, exists := unique[id]; exists {
			continue
		}
		unique[id] = struct{}{}
		filtered = append(filtered, id)
	}
	if len(filtered) == 0 {
		return map[int64]string{}, nil
	}
	resp, err := svcCtx.SystemCli.SysTenantList(ctx, &system.SysTenantListReq{
		Page: &common.PageReq{Limit: int64(len(filtered))},
		Ids:  filtered,
	})
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(resp.GetData()))
	for _, tenant := range resp.GetData() {
		names[tenant.GetId()] = tenant.GetTenantName()
	}
	return names, nil
}

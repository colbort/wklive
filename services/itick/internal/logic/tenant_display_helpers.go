package logic

import (
	"context"

	"wklive/services/itick/models"
)

const batchScanPageSize int64 = 500
const productLookupBatchSize = 500

func collectProductsByIDs(ctx context.Context, model models.ItickProductModel, ids []int64) (map[int64]*models.TItickProduct, error) {
	if len(ids) == 0 {
		return map[int64]*models.TItickProduct{}, nil
	}

	uniqueIDs := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	out := make(map[int64]*models.TItickProduct, len(uniqueIDs))
	for start := 0; start < len(uniqueIDs); start += productLookupBatchSize {
		end := start + productLookupBatchSize
		if end > len(uniqueIDs) {
			end = len(uniqueIDs)
		}

		items, err := model.FindByIds(ctx, uniqueIDs[start:end])
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			out[item.Id] = item
		}
	}

	return out, nil
}

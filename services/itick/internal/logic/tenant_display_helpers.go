package logic

import (
	"context"

	"wklive/services/itick/models"
)

const batchScanPageSize int64 = 500

func collectTenantCategories(ctx context.Context, model models.ItickTenantCategoryModel, tenantId int64) ([]*models.TItickTenantCategory, error) {
	cursor := int64(0)
	out := make([]*models.TItickTenantCategory, 0)

	for {
		items, nextCursor, err := model.FindPage(ctx, tenantId, cursor, batchScanPageSize)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		out = append(out, items...)
		if len(items) < int(batchScanPageSize) || nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return out, nil
}

func collectTenantProducts(ctx context.Context, model models.ItickTenantProductModel, tenantId int64) ([]*models.TItickTenantProduct, error) {
	cursor := int64(0)
	out := make([]*models.TItickTenantProduct, 0)

	for {
		items, nextCursor, err := model.FindPage(ctx, tenantId, cursor, batchScanPageSize)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		out = append(out, items...)
		if len(items) < int(batchScanPageSize) || nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}

	return out, nil
}

func collectProducts(ctx context.Context, model models.ItickProductModel) ([]*models.TItickProduct, error) {
	cursor := int64(0)
	out := make([]*models.TItickProduct, 0)

	for {
		items, _, err := model.FindPage(ctx, 0, "", "", "", 0, 0, cursor, batchScanPageSize)
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		out = append(out, items...)
		if len(items) < int(batchScanPageSize) {
			break
		}
		cursor = items[len(items)-1].Id
	}

	return out, nil
}

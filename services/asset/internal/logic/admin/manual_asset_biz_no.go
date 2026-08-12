package adminlogic

import (
	"context"

	"wklive/common/generate"
	"wklive/services/asset/internal/svc"
)

const (
	manualAddBizNoPrefix    = "AMADD"
	manualSubBizNoPrefix    = "AMSUB"
	manualFreezeBizNoPrefix = "AMFREEZE"
	manualLockBizNoPrefix   = "AMLOCK"
)

func generateManualAssetBizNo(ctx context.Context, svcCtx *svc.ServiceContext, prefix string) (string, error) {
	return generate.GenerateNo(svcCtx.Redis, ctx, "asset_manual", prefix, "")
}

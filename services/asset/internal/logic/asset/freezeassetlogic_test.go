package assetlogic

import (
	"testing"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/models"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

func TestAssetFreezeReplayMatchesEconomicIdentity(t *testing.T) {
	req := &asset.FreezeAssetReq{
		TenantId: 11, UserId: 22, WalletType: common.WalletType_WALLET_TYPE_OPTION,
		Coin: "USDT", Amount: "12.5", BizType: asset.BizType_BIZ_TYPE_OPTION,
		SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER, BizId: 33,
		BizNo: "OPTION-FREEZE-33", ExpireTime: 44,
	}
	stored := &models.TAssetFreeze{
		TenantId: 11, UserId: 22, WalletType: int64(common.WalletType_WALLET_TYPE_OPTION),
		Coin: "USDT", Amount: decimal.RequireFromString("12.5000"), BizType: "option",
		SceneType: "place_order", BizId: 33, BizNo: "OPTION-FREEZE-33", ExpireTime: 44,
	}
	if !assetFreezeReplayMatches(stored, req, decimal.RequireFromString(req.Amount)) {
		t.Fatal("identical freeze replay must match")
	}

	cases := []struct {
		name   string
		mutate func(*asset.FreezeAssetReq)
	}{
		{name: "user", mutate: func(in *asset.FreezeAssetReq) { in.UserId++ }},
		{name: "wallet", mutate: func(in *asset.FreezeAssetReq) { in.WalletType = common.WalletType_WALLET_TYPE_CONTRACT }},
		{name: "coin", mutate: func(in *asset.FreezeAssetReq) { in.Coin = "BTC" }},
		{name: "biz id", mutate: func(in *asset.FreezeAssetReq) { in.BizId++ }},
		{name: "expiry", mutate: func(in *asset.FreezeAssetReq) { in.ExpireTime++ }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := proto.Clone(req).(*asset.FreezeAssetReq)
			tc.mutate(changed)
			if assetFreezeReplayMatches(stored, changed, decimal.RequireFromString(changed.Amount)) {
				t.Fatal("changed economic identity must not match")
			}
		})
	}
	if assetFreezeReplayMatches(stored, req, decimal.RequireFromString("12.5001")) {
		t.Fatal("changed amount must not match")
	}
	stored.BizId = 0
	if !assetFreezeReplayMatches(stored, req, decimal.RequireFromString(req.Amount)) {
		t.Fatal("legacy freeze evidence without biz_id must be adopted instead of frozen again")
	}
}

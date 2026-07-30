package tasklogic

import (
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/option"
)

func TestOptionAssetRetryDelay(t *testing.T) {
	tests := []struct {
		retry int64
		want  time.Duration
	}{
		{retry: 0, want: time.Second},
		{retry: 1, want: time.Second},
		{retry: 2, want: 2 * time.Second},
		{retry: 10, want: 512 * time.Second},
		{retry: 20, want: 512 * time.Second},
	}
	for _, test := range tests {
		if got := optionAssetRetryDelay(test.retry); got != test.want {
			t.Fatalf("retry=%d got=%s want=%s", test.retry, got, test.want)
		}
	}
}

func TestOptionInstructionAssetFacts(t *testing.T) {
	tests := []struct {
		action option.AssetInstructionAction
		scene  asset.SceneType
		op     asset.AssetOpType
	}{
		{option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE, asset.SceneType_SCENE_TYPE_PLACE_ORDER, asset.AssetOpType_ASSET_OP_TYPE_FREEZE},
		{option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN, asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_FREEZE_DEDUCT},
		{option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN, asset.SceneType_SCENE_TYPE_CANCEL_ORDER, asset.AssetOpType_ASSET_OP_TYPE_UNFREEZE},
		{option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE, asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_ADD},
		{option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE, asset.SceneType_SCENE_TYPE_TRADE_MATCH, asset.AssetOpType_ASSET_OP_TYPE_SUB},
	}
	for _, test := range tests {
		scene, op, err := optionInstructionAssetFacts(int64(test.action))
		if err != nil {
			t.Fatalf("action=%s: %v", test.action, err)
		}
		if scene != test.scene || op != test.op {
			t.Fatalf("action=%s got scene=%s op=%s", test.action, scene, op)
		}
	}
	if _, _, err := optionInstructionAssetFacts(999); err == nil {
		t.Fatal("unknown instruction action must be rejected")
	}
}

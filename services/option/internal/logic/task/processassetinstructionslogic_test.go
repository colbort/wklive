package tasklogic

import (
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/option"
	"wklive/services/option/models"
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

func TestExerciseCompletionSkipsNonTerminalAssetStep(t *testing.T) {
	logic := &ProcessAssetInstructionsLogic{}
	if err := logic.completeExerciseTransition(&models.TOptionAssetInstruction{
		BizNo:  "EX-CAPACITY",
		StepNo: 1,
	}); err != nil {
		t.Fatalf("step-1 completion should not query the whole exercise: %v", err)
	}
}

func TestTradeCorrectionInstructionOutcome(t *testing.T) {
	success := int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS)
	pending := int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING)
	manual := int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_MANUAL_REVIEW)

	allSuccess, manualReview, lastError := tradeCorrectionInstructionOutcome(
		[]*models.TOptionAssetInstruction{{Status: success}, {Status: success}},
	)
	if !allSuccess || manualReview || lastError != "" {
		t.Fatalf("successful correction summarized incorrectly: %t %t %q", allSuccess, manualReview, lastError)
	}

	allSuccess, manualReview, _ = tradeCorrectionInstructionOutcome(
		[]*models.TOptionAssetInstruction{{Status: success}, {Status: pending}},
	)
	if allSuccess || manualReview {
		t.Fatalf("pending correction summarized incorrectly: %t %t", allSuccess, manualReview)
	}

	allSuccess, manualReview, lastError = tradeCorrectionInstructionOutcome(
		[]*models.TOptionAssetInstruction{
			{Status: success},
			{Status: manual, LastErrorMsg: "insufficient balance"},
		},
	)
	if allSuccess || !manualReview || lastError != "insufficient balance" {
		t.Fatalf("manual correction summarized incorrectly: %t %t %q", allSuccess, manualReview, lastError)
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

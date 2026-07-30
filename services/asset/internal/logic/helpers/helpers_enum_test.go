package helpers

import (
	"testing"

	"wklive/proto/asset"
)

func TestInsuranceFundBizTypeMapping(t *testing.T) {
	const stored = "insurance_fund"
	if got := FromBizTypeEnum(asset.BizType_BIZ_TYPE_INSURANCE_FUND); got != stored {
		t.Fatalf("FromBizTypeEnum() = %q, want %q", got, stored)
	}
	if got := ToBizTypeValue(stored); got != asset.BizType_BIZ_TYPE_INSURANCE_FUND {
		t.Fatalf("ToBizTypeValue() = %v, want %v", got, asset.BizType_BIZ_TYPE_INSURANCE_FUND)
	}
}

func TestInsuranceFundSceneTypeMapping(t *testing.T) {
	tests := []struct {
		name   string
		value  asset.SceneType
		stored string
	}{
		{
			name:   "cover",
			value:  asset.SceneType_SCENE_TYPE_INSURANCE_FUND_COVER,
			stored: "insurance_fund_cover",
		},
		{
			name:   "reversal",
			value:  asset.SceneType_SCENE_TYPE_INSURANCE_FUND_REVERSAL,
			stored: "insurance_fund_reversal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromSceneTypeEnum(tt.value); got != tt.stored {
				t.Fatalf("FromSceneTypeEnum() = %q, want %q", got, tt.stored)
			}
			if got := ToSceneTypeValue(tt.stored); got != tt.value {
				t.Fatalf("ToSceneTypeValue() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestPlatformBackstopEnumMappings(t *testing.T) {
	if got := FromBizTypeEnum(asset.BizType_BIZ_TYPE_PLATFORM_BACKSTOP); got != "platform_backstop" {
		t.Fatalf("FromBizTypeEnum() = %q", got)
	}
	if got := ToBizTypeValue("platform_backstop"); got != asset.BizType_BIZ_TYPE_PLATFORM_BACKSTOP {
		t.Fatalf("ToBizTypeValue() = %v", got)
	}
	if got := FromSceneTypeEnum(asset.SceneType_SCENE_TYPE_PLATFORM_BACKSTOP_COVER); got != "platform_backstop_cover" {
		t.Fatalf("FromSceneTypeEnum() = %q", got)
	}
	if got := ToSceneTypeValue("platform_backstop_cover"); got != asset.SceneType_SCENE_TYPE_PLATFORM_BACKSTOP_COVER {
		t.Fatalf("ToSceneTypeValue() = %v", got)
	}
}

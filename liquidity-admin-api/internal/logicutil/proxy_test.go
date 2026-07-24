package logicutil

import (
	"testing"

	commonpb "wklive/proto/common"
	liquiditypb "wklive/proto/liquidity"
)

type RespBase struct {
	Code int32
	Msg  string
}

type testItem struct {
	Id     int64
	Status int32
}

type testList struct {
	RespBase
	Data []testItem
}

func TestConvertMapsProtoBaseAndItems(t *testing.T) {
	src := &liquiditypb.GetProviderListResp{
		Base: &commonpb.RespBase{Code: 7, Msg: "ok"},
		Data: []*liquiditypb.LiquidityProvider{
			{Id: 12, Status: liquiditypb.ProviderStatus_PROVIDER_STATUS_ENABLED},
		},
	}

	got := Convert[testList](src)
	if got.Code != 7 || got.Msg != "ok" {
		t.Fatalf("base not copied: %+v", got)
	}
	if len(got.Data) != 1 || got.Data[0].Id != 12 ||
		got.Data[0].Status != int32(liquiditypb.ProviderStatus_PROVIDER_STATUS_ENABLED) {
		t.Fatalf("data not copied: %+v", got.Data)
	}
}

package tasklogic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/market"

	"google.golang.org/grpc"
)

type authoritativeSnapshotClientStub struct {
	request *market.GetAuthoritativeSnapshotReq
}

func (s *authoritativeSnapshotClientStub) GetAuthoritativeSnapshot(_ context.Context, in *market.GetAuthoritativeSnapshotReq, _ ...grpc.CallOption) (*market.GetAuthoritativeSnapshotResp, error) {
	s.request = in
	return &market.GetAuthoritativeSnapshotResp{
		Base: &common.RespBase{Code: 200},
		Data: &market.AuthoritativeSnapshot{
			SnapshotId:        "snapshot-1",
			Authority:         "market-ws",
			SnapshotKind:      "FINAL_QUOTE",
			CategoryCode:      "crypto",
			Market:            "BA",
			Symbol:            "BTCUSDT",
			Price:             "100.25",
			SourceTimestamp:   1_000,
			SnapshotTimestamp: 1_001,
			Revision:          1_000,
		},
	}, nil
}

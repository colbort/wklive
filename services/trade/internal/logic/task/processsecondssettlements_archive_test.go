package tasklogic

import (
	"context"

	commonpb "wklive/proto/common"
	"wklive/proto/itick"

	"google.golang.org/grpc"
)

type authoritativeSnapshotClientStub struct {
	request *itick.GetAuthoritativeSnapshotReq
}

func (s *authoritativeSnapshotClientStub) GetAuthoritativeSnapshot(_ context.Context, in *itick.GetAuthoritativeSnapshotReq, _ ...grpc.CallOption) (*itick.GetAuthoritativeSnapshotResp, error) {
	s.request = in
	return &itick.GetAuthoritativeSnapshotResp{
		Base: &commonpb.RespBase{Code: 200},
		Data: &itick.AuthoritativeSnapshot{
			SnapshotId:        "snapshot-1",
			Authority:         "itick-ws",
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

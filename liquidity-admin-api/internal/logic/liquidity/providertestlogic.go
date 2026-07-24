// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package liquidity

import (
	"context"

	"wklive/liquidity-admin-api/internal/logicutil"
	"wklive/liquidity-admin-api/internal/svc"
	"wklive/liquidity-admin-api/internal/types"
	pb "wklive/proto/liquidity"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProviderTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderTestLogic {
	return &ProviderTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderTestLogic) ProviderTest(req *types.ProviderActionReq) (resp *types.RespBase, err error) {
	_, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	out, err := l.svcCtx.LiquidityCli.TestProviderConnection(l.ctx, &pb.TestProviderConnectionReq{
		Id: req.Id, OperatorId: userID,
	})
	if err != nil {
		return nil, err
	}
	resp = logicutil.Convert[types.RespBase](out.Base)
	if out.Message != "" {
		resp.Msg = out.Message
	}
	return resp, nil
}

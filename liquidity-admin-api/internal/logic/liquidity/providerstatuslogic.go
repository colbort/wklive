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

type ProviderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewProviderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProviderStatusLogic {
	return &ProviderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ProviderStatusLogic) ProviderStatus(req *types.ProviderStatusReq) (resp *types.RespBase, err error) {
	_, userID, err := logicutil.Identity(l.ctx)
	if err != nil {
		return nil, err
	}
	out, err := l.svcCtx.LiquidityCli.SetProviderStatus(l.ctx, &pb.SetProviderStatusReq{
		Id: req.Id, Status: pb.ProviderStatus(req.Status),
		Version: req.Version, OperatorId: userID, Remark: req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return logicutil.Convert[types.RespBase](out.Base), nil
}

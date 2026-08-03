// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package staking

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OperationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOperationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OperationListLogic {
	return &OperationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OperationListLogic) OperationList(req *types.OperationListReq) (resp *types.OperationListResp, err error) {
	return logicutil.Proxy[types.OperationListResp](l.ctx, req, l.svcCtx.StakingCli.OperationList)
}

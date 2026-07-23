// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSettlementsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSettlementsLogic {
	return &ListSettlementsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSettlementsLogic) ListSettlements(req *types.ListSettlementsReq) (resp *types.ListSettlementsResp, err error) {
	return logicutil.Proxy[types.ListSettlementsResp](l.ctx, req, l.svcCtx.OptionCli.ListSettlements)
}

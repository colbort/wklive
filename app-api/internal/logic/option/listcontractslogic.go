// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package option

import (
	"context"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListContractsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListContractsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListContractsLogic {
	return &ListContractsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListContractsLogic) ListContracts(req *types.ListContractsReq) (resp *types.ListContractsResp, err error) {
	return logicutil.Proxy[types.ListContractsResp](l.ctx, req, l.svcCtx.OptionCli.ListContracts)
}

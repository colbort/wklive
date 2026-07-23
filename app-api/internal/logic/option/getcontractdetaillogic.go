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

type GetContractDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContractDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractDetailLogic {
	return &GetContractDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetContractDetailLogic) GetContractDetail(req *types.GetContractDetailReq) (resp *types.GetContractDetailResp, err error) {
	return logicutil.Proxy[types.GetContractDetailResp](l.ctx, req, l.svcCtx.OptionCli.GetContractDetail)
}

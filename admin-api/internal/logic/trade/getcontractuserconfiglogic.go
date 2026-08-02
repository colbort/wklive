// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContractUserConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContractUserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractUserConfigLogic {
	return &GetContractUserConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetContractUserConfigLogic) GetContractUserConfig(req *types.GetContractUserConfigReq) (resp *types.GetContractUserConfigResp, err error) {
	return logicutil.Proxy[types.GetContractUserConfigResp](l.ctx, req, l.svcCtx.TradeCli.GetContractUserConfig)
}

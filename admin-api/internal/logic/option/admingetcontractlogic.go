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

type AdminGetContractLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetContractLogic {
	return &AdminGetContractLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetContractLogic) AdminGetContract(req *types.GetContractReq) (resp *types.GetContractResp, err error) {
	return logicutil.Proxy[types.GetContractResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetContract)
}

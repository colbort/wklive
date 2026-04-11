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

type AdminCreateContractLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminCreateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateContractLogic {
	return &AdminCreateContractLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminCreateContractLogic) AdminCreateContract(req *types.CreateContractReq) (resp *types.CreateContractResp, err error) {
	return logicutil.Proxy[types.CreateContractResp](l.ctx, req, l.svcCtx.OptionCli.AdminCreateContract)
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractUserConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetContractUserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractUserConfigLogic {
	return &SetContractUserConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetContractUserConfigLogic) SetContractUserConfig(req *types.SetContractUserConfigReq) (resp *types.AdminCommonResp, err error) {
	// todo: add your logic here and delete this line

	return
}

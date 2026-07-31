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

type CreateCorporateActionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCorporateActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCorporateActionLogic {
	return &CreateCorporateActionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCorporateActionLogic) CreateCorporateAction(req *types.CreateCorporateActionReq) (resp *types.GetCorporateActionResp, err error) {
	return logicutil.Proxy[types.GetCorporateActionResp](
		l.ctx, req, l.svcCtx.OptionCli.CreateCorporateAction,
	)
}

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

type AdminListExercisesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminListExercisesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListExercisesLogic {
	return &AdminListExercisesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminListExercisesLogic) AdminListExercises(req *types.ListExercisesReq) (resp *types.ListExercisesResp, err error) {
	return logicutil.Proxy[types.ListExercisesResp](l.ctx, req, l.svcCtx.OptionCli.AdminListExercises)
}

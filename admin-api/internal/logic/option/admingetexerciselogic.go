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

type AdminGetExerciseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminGetExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetExerciseLogic {
	return &AdminGetExerciseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminGetExerciseLogic) AdminGetExercise(req *types.GetExerciseReq) (resp *types.GetExerciseResp, err error) {
	return logicutil.Proxy[types.GetExerciseResp](l.ctx, req, l.svcCtx.OptionCli.AdminGetExercise)
}

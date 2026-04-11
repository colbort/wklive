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

type AppExerciseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAppExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppExerciseLogic {
	return &AppExerciseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AppExerciseLogic) AppExercise(req *types.AppExerciseReq) (resp *types.AppExerciseResp, err error) {
	return logicutil.Proxy[types.AppExerciseResp](l.ctx, req, l.svcCtx.OptionCli.AppExercise)
}

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

type ExerciseLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExerciseLogic {
	return &ExerciseLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExerciseLogic) Exercise(req *types.ExerciseReq) (resp *types.ExerciseResp, err error) {
	return logicutil.Proxy[types.ExerciseResp](l.ctx, req, l.svcCtx.OptionCli.Exercise)
}

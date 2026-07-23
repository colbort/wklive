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

type ListExercisesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListExercisesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListExercisesLogic {
	return &ListExercisesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListExercisesLogic) ListExercises(req *types.ListExercisesReq) (resp *types.ListExercisesResp, err error) {
	return logicutil.Proxy[types.ListExercisesResp](l.ctx, req, l.svcCtx.OptionCli.ListExercises)
}

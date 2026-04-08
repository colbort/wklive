package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppListExercisesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppListExercisesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppListExercisesLogic {
	return &AppListExercisesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取行权记录列表
func (l *AppListExercisesLogic) AppListExercises(in *option.AppListExercisesReq) (*option.AppListExercisesResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppListExercisesResp{}, nil
}

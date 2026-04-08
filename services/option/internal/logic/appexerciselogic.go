package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppExerciseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppExerciseLogic {
	return &AppExerciseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发起行权
func (l *AppExerciseLogic) AppExercise(in *option.AppExerciseReq) (*option.AppExerciseResp, error) {
	// todo: add your logic here and delete this line

	return &option.AppExerciseResp{}, nil
}

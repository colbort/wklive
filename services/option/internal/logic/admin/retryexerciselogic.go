package adminlogic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetryExerciseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryExerciseLogic {
	return &RetryExerciseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按行权单重试失败或人工处理的资产指令
func (l *RetryExerciseLogic) RetryExercise(in *option.RetryExerciseReq) (*option.CommonResp, error) {
	// todo: add your logic here and delete this line

	return &option.CommonResp{}, nil
}

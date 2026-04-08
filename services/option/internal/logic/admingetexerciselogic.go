package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetExerciseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetExerciseLogic {
	return &AdminGetExerciseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个行权记录详情
func (l *AdminGetExerciseLogic) AdminGetExercise(in *option.GetExerciseReq) (*option.GetExerciseResp, error) {
	// todo: add your logic here and delete this line

	return &option.GetExerciseResp{}, nil
}

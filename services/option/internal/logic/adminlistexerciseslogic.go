package logic

import (
	"context"

	"wklive/proto/option"
	"wklive/services/option/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminListExercisesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminListExercisesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminListExercisesLogic {
	return &AdminListExercisesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询行权记录列表
func (l *AdminListExercisesLogic) AdminListExercises(in *option.ListExercisesReq) (*option.ListExercisesResp, error) {
	// todo: add your logic here and delete this line

	return &option.ListExercisesResp{}, nil
}

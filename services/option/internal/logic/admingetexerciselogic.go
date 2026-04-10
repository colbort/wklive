package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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
	item, err := findExerciseByNoOrID(l.ctx, l.svcCtx, in.TenantId, in.Id, in.ExerciseNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetExerciseResp{Base: helper.GetErrResp(404, "行权记录不存在")}, nil
		}
		return nil, err
	}
	data, err := buildExerciseDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}

	return &option.GetExerciseResp{Base: helper.OkResp(), Data: data}, nil
}

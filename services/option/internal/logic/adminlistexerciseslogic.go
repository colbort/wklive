package logic

import (
	"context"
	"errors"

	pageutil "wklive/common/pageutil"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionExerciseModel.FindPage(l.ctx, models.OptionExercisePageFilter{
		TenantId:          in.TenantId,
		Uid:               in.Uid,
		AccountId:         in.AccountId,
		ContractId:        in.ContractId,
		ExerciseType:      int64(in.ExerciseType),
		Status:            int64(in.Status),
		ExerciseTimeStart: pageutil.TimeRangeStart(in.ExerciseTimeRange),
		ExerciseTimeEnd:   pageutil.TimeRangeEnd(in.ExerciseTimeRange),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	list := make([]*option.OptionExerciseDetail, 0, len(items))
	lastID := int64(0)
	for _, item := range items {
		lastID = item.Id
		detail, err := buildExerciseDetail(l.ctx, l.svcCtx, item)
		if err != nil {
			return nil, err
		}
		list = append(list, detail)
	}

	return &option.ListExercisesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}

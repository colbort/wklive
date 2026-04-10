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
	cursor, limit := pageutil.Input(in.Page)
	items, total, err := l.svcCtx.OptionExerciseModel.FindPage(l.ctx, models.OptionExercisePageFilter{
		TenantId:          in.TenantId,
		Uid:               in.Uid,
		AccountId:         in.AccountId,
		ContractId:        in.ContractId,
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

	return &option.AppListExercisesResp{
		Base: pageutil.Base(cursor, limit, len(items), total, lastID),
		List: list,
		Page: pageutil.Output(in.Page, limit),
	}, nil
}

package logic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

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
	position, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.PositionId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppExerciseResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if position.TenantId != in.TenantId || position.Uid != in.Uid || position.AccountId != in.AccountId {
		return &option.AppExerciseResp{Base: helper.GetErrResp(403, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx))}, nil
	}
	if in.ContractId != 0 && position.ContractId != in.ContractId {
		return &option.AppExerciseResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ContractPositionMismatch, l.ctx))}, nil
	}

	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppExerciseResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	exerciseQty, err := conv.ParseFloatField(in.ExerciseQty)
	if err != nil || exerciseQty <= 0 {
		return &option.AppExerciseResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ExerciseQuantityFormatError, l.ctx))}, nil
	}
	if position.ExerciseableQty > 0 && exerciseQty > position.ExerciseableQty {
		return &option.AppExerciseResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}

	exerciseNo, err := l.svcCtx.GenerateBizNo(l.ctx, "EX")
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	item := &models.TOptionExercise{
		TenantId:        in.TenantId,
		ExerciseNo:      exerciseNo,
		Uid:             in.Uid,
		AccountId:       in.AccountId,
		ContractId:      position.ContractId,
		PositionId:      position.Id,
		ExerciseType:    int64(option.ExerciseType_EXERCISE_TYPE_USER),
		ExerciseQty:     exerciseQty,
		StrikePrice:     contract.StrikePrice,
		SettlementPrice: 0,
		ExerciseAmount:  0,
		ProfitAmount:    0,
		Fee:             0,
		FeeCoin:         contract.SettleCoin,
		Status:          int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING),
		ExerciseTime:    now,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	result, err := l.svcCtx.OptionExerciseModel.Insert(l.ctx, item)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &option.AppExerciseResp{Base: helper.OkResp(), ExerciseNo: item.ExerciseNo, ExerciseId: id}, nil
}

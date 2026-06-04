package logic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	position, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.PositionId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if position.TenantId != tenantId || position.UserId != userId || position.AccountId != in.AccountId {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx))}, nil
	}
	if position.Side != int64(option.PositionSide_POSITION_SIDE_LONG) || position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx))}, nil
	}
	if in.ContractId != 0 && position.ContractId != in.ContractId {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.ContractPositionMismatch, i18n.Translate(i18n.ContractPositionMismatch, l.ctx))}, nil
	}

	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}

	exerciseQty, err := conv.ParseFloatField(in.ExerciseQty)
	if err != nil || exerciseQty <= 0 {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.ExerciseQuantityFormatError, i18n.Translate(i18n.ExerciseQuantityFormatError, l.ctx))}, nil
	}
	if position.ExerciseableQty+optionFloatEpsilon < exerciseQty {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	if position.AvailableQty+optionFloatEpsilon < exerciseQty {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	now := time.Now().Unix()
	if contract.ExerciseStyle == int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN) && now < contract.ExpireTime {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.EuropeanOptionNotExpired, i18n.Translate(i18n.EuropeanOptionNotExpired, l.ctx))}, nil
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, tenantId, contract.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.MarketNotFound, i18n.Translate(i18n.MarketNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	settlementPrice := market.UnderlyingPrice
	profitAmount := optionSettlementPayoff(contract, settlementPrice, exerciseQty)
	if profitAmount <= optionFloatEpsilon {
		return &option.AppExerciseResp{Base: helper.GetErrResp(i18n.OptionNotInTheMoney, i18n.Translate(i18n.OptionNotInTheMoney, l.ctx))}, nil
	}

	exerciseNo, err := l.svcCtx.GenerateBizNo(l.ctx, "EX")
	if err != nil {
		return nil, err
	}

	item := &models.TOptionExercise{
		TenantId:        tenantId,
		ExerciseNo:      exerciseNo,
		UserId:          userId,
		AccountId:       in.AccountId,
		ContractId:      position.ContractId,
		PositionId:      position.Id,
		ExerciseType:    int64(option.ExerciseType_EXERCISE_TYPE_USER),
		ExerciseQty:     exerciseQty,
		StrikePrice:     contract.StrikePrice,
		SettlementPrice: settlementPrice,
		ExerciseAmount:  optionExerciseAmount(contract, exerciseQty),
		ProfitAmount:    profitAmount,
		Fee:             0,
		FeeCoin:         contract.SettleCoin,
		Status:          int64(option.ExerciseStatus_EXERCISE_STATUS_DONE),
		ExerciseTime:    now,
		FinishTime:      now,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionExerciseModel)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionPositionModel)
		accountModel := models.NewTOptionAccountModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionAccountModel)
		billModel := models.NewTOptionBillModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionBillModel)

		result, err := exerciseModel.Insert(ctx, item)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		item.Id = id

		position.PositionQty = normalizeZero(maxFloat64(position.PositionQty-exerciseQty, 0))
		position.AvailableQty = normalizeZero(maxFloat64(position.AvailableQty-exerciseQty, 0))
		position.ExerciseableQty = normalizeZero(maxFloat64(position.ExerciseableQty-exerciseQty, 0))
		position.PositionValue = position.MarkPrice * position.PositionQty * optionMultiplier(contract)
		position.RealizedPnl = normalizeZero(position.RealizedPnl + profitAmount)
		position.UpdateTimes = now
		if position.PositionQty <= optionFloatEpsilon {
			position.PositionQty = 0
			position.AvailableQty = 0
			position.FrozenQty = 0
			position.ExerciseableQty = 0
			position.PositionValue = 0
			position.UnrealizedPnl = 0
			position.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
		}
		if err := positionModel.Update(ctx, position); err != nil {
			return err
		}
		return applyOptionAccountDelta(ctx, accountModel, billModel, tenantId, userId, in.AccountId, contract.SettleCoin, profitAmount, int64(option.BillRefType_BILL_REF_TYPE_EXERCISE), id, item.ExerciseNo, "option exercise profit", true, now)
	})
	if err != nil {
		return nil, err
	}

	return &option.AppExerciseResp{Base: helper.OkResp(), ExerciseNo: item.ExerciseNo, ExerciseId: id}, nil
}

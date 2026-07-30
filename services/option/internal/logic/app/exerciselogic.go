package applogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ExerciseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExerciseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExerciseLogic {
	return &ExerciseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发起行权
func (l *ExerciseLogic) Exercise(in *option.ExerciseReq) (*option.ExerciseResp, error) {
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
			return &option.ExerciseResp{Base: helper.ErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if position.TenantId != tenantId || position.UserId != userId || position.AccountId != in.AccountId {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx))}, nil
	}
	if position.Side != int64(common.PositionSide_POSITION_SIDE_LONG) || position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.NoPermissionOperatePosition, i18n.Translate(i18n.NoPermissionOperatePosition, l.ctx))}, nil
	}
	if in.ContractId != 0 && position.ContractId != in.ContractId {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ContractPositionMismatch, i18n.Translate(i18n.ContractPositionMismatch, l.ctx))}, nil
	}

	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.ExerciseResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if contract.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN) ||
		contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}

	exerciseQty, err := conv.ParseDecimalField(in.ExerciseQty)
	if err != nil || !exerciseQty.IsPositive() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExerciseQuantityFormatError, i18n.Translate(i18n.ExerciseQuantityFormatError, l.ctx))}, nil
	}
	if position.ExerciseableQty.LessThan(exerciseQty) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	if position.AvailableQty.LessThan(exerciseQty) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	now := time.Now().Unix()
	if now < contract.ListTime || now >= contract.ExpireTime {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, tenantId, contract.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.ExerciseResp{Base: helper.ErrResp(i18n.MarketNotFound, i18n.Translate(i18n.MarketNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	settlementPrice := market.UnderlyingPrice
	if !settlementPrice.IsPositive() || market.SnapshotTime <= 0 ||
		market.SnapshotTime > now || now-market.SnapshotTime > 30 {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.MarketNotFound, i18n.Translate(i18n.MarketNotFound, l.ctx))}, nil
	}
	profitAmount := optionSettlementPayoff(contract, settlementPrice, exerciseQty)
	if !profitAmount.IsPositive() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OptionNotInTheMoney, i18n.Translate(i18n.OptionNotInTheMoney, l.ctx))}, nil
	}

	exerciseNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "EX", "")
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
		Fee:             profitAmount.Mul(contract.ExerciseFeeRate).Round(16),
		FeeCoin:         contract.SettleCoin,
		Status:          int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING),
		ExerciseTime:    now,
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := positionModel.FindOneForUpdate(ctx, position.Id)
		if err != nil {
			return err
		}
		if current.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) ||
			current.AvailableQty.LessThan(exerciseQty) ||
			current.ExerciseableQty.LessThan(exerciseQty) {
			return i18n.StatusError(ctx, i18n.ExercisableQuantityExceeded)
		}
		current.AvailableQty = current.AvailableQty.Sub(exerciseQty)
		current.FrozenQty = current.FrozenQty.Add(exerciseQty)
		current.UpdateTimes = now
		if err := positionModel.Update(ctx, current); err != nil {
			return err
		}
		result, err := exerciseModel.Insert(ctx, item)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		item.Id = id

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &option.ExerciseResp{Base: helper.OkResp(), Data: &option.ExerciseData{ExerciseNo: item.ExerciseNo, ExerciseId: id}}, nil
}

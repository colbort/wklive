package applogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
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
	clientExerciseID := strings.TrimSpace(in.ClientExerciseId)
	if clientExerciseID == "" || len(clientExerciseID) > 64 {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	exerciseQty, err := conv.ParseDecimalField(in.ExerciseQty)
	if err != nil || !exerciseQty.IsPositive() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExerciseQuantityFormatError, i18n.Translate(i18n.ExerciseQuantityFormatError, l.ctx))}, nil
	}
	existing, err := l.svcCtx.OptionExerciseModel.FindOneByClientExerciseId(
		l.ctx, tenantId, userId, clientExerciseID,
	)
	if err == nil {
		return exerciseReplayResponse(l.ctx, existing, in, exerciseQty), nil
	}
	if !errors.Is(err, models.ErrNotFound) {
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
	if !allowsExerciseSubmission(contract.Status) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	if blocked, checkErr := l.svcCtx.OptionCorporateActionContractModel.IsContractMigrationActive(
		l.ctx, contract.TenantId, contract.Id,
	); checkErr != nil {
		return nil, checkErr
	} else if blocked {
		return &option.ExerciseResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}

	if position.ExerciseableQty.LessThan(exerciseQty) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	if position.AvailableQty.LessThan(exerciseQty) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExercisableQuantityExceeded, i18n.Translate(i18n.ExercisableQuantityExceeded, l.ctx))}, nil
	}
	now := time.Now().Unix()
	if now < contract.ListTime || contract.ExerciseCutoffTime <= 0 || now >= contract.ExerciseCutoffTime {
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
	if !logichelpers.IsUnderlyingFresh(market, now, 30) {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.MarketNotFound, i18n.Translate(i18n.MarketNotFound, l.ctx))}, nil
	}
	profitAmount := optionSettlementPayoff(contract, settlementPrice, exerciseQty)
	if !profitAmount.IsPositive() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OptionNotInTheMoney, i18n.Translate(i18n.OptionNotInTheMoney, l.ctx))}, nil
	}
	fee := profitAmount.Mul(contract.ExerciseFeeRate).Round(16)
	if !profitAmount.Sub(fee).IsPositive() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.OptionNotInTheMoney, i18n.Translate(i18n.OptionNotInTheMoney, l.ctx))}, nil
	}
	if contract.QtyStep.IsPositive() && !exerciseQty.Mod(contract.QtyStep).IsZero() {
		return &option.ExerciseResp{Base: helper.ErrResp(i18n.ExerciseQuantityFormatError, i18n.Translate(i18n.ExerciseQuantityFormatError, l.ctx))}, nil
	}

	exerciseNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "EX", "")
	if err != nil {
		return nil, err
	}

	item := &models.TOptionExercise{
		TenantId:         tenantId,
		ExerciseNo:       exerciseNo,
		ClientExerciseId: clientExerciseID,
		UserId:           userId,
		AccountId:        in.AccountId,
		ContractId:       position.ContractId,
		PositionId:       position.Id,
		ExerciseType:     int64(option.ExerciseType_EXERCISE_TYPE_USER),
		ExerciseQty:      exerciseQty,
		StrikePrice:      contract.StrikePrice,
		SettlementPrice:  settlementPrice,
		ExerciseAmount:   optionExerciseAmount(contract, exerciseQty),
		ProfitAmount:     profitAmount,
		Fee:              fee,
		FeeCoin:          contract.SettleCoin,
		Status:           int64(option.ExerciseStatus_EXERCISE_STATUS_PENDING),
		ExerciseTime:     now,
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		actionContractModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		currentContract, err := contractModel.FindOneForUpdate(ctx, contract.Id)
		if err != nil {
			return err
		}
		txNow := time.Now().Unix()
		if currentContract.TenantId != tenantId ||
			currentContract.ExerciseStyle != int64(option.ExerciseStyle_EXERCISE_STYLE_AMERICAN) ||
			currentContract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) ||
			!allowsExerciseSubmission(currentContract.Status) ||
			txNow < currentContract.ListTime || currentContract.ExerciseCutoffTime <= 0 ||
			txNow >= currentContract.ExerciseCutoffTime {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		current, err := positionModel.FindOneForUpdate(ctx, position.Id)
		if err != nil {
			return err
		}
		existing, findErr := exerciseModel.FindOneByClientExerciseId(
			ctx, tenantId, userId, clientExerciseID,
		)
		if findErr == nil {
			if !sameExerciseRequest(existing, in, exerciseQty) {
				return i18n.StatusError(ctx, i18n.ParamError)
			}
			item = existing
			id = existing.Id
			return nil
		}
		if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		blocked, checkErr := actionContractModel.IsContractMigrationActive(
			ctx, current.TenantId, current.ContractId,
		)
		if checkErr != nil {
			return checkErr
		}
		if blocked {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
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
		if isExerciseDuplicateKeyError(err) {
			existing, findErr := l.svcCtx.OptionExerciseModel.FindOneByClientExerciseId(
				l.ctx, tenantId, userId, clientExerciseID,
			)
			if findErr == nil {
				return exerciseReplayResponse(l.ctx, existing, in, exerciseQty), nil
			}
		}
		return nil, err
	}

	return &option.ExerciseResp{Base: helper.OkResp(), Data: &option.ExerciseData{ExerciseNo: item.ExerciseNo, ExerciseId: id}}, nil
}

func sameExerciseRequest(item *models.TOptionExercise, in *option.ExerciseReq, qty decimal.Decimal) bool {
	if item == nil || in == nil {
		return false
	}
	return item.AccountId == in.AccountId &&
		item.PositionId == in.PositionId &&
		(in.ContractId == 0 || item.ContractId == in.ContractId) &&
		item.ExerciseQty.Equal(qty)
}

// Trading halts stop new orders but must not strip an option holder's exercise
// right. Contracts outside the listed lifecycle cannot accept a new exercise or
// expiry instruction; expiry processing owns them after the cutoff transition.
func allowsExerciseSubmission(status int64) bool {
	return status == int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
		status == int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
}

func exerciseReplayResponse(
	ctx context.Context,
	item *models.TOptionExercise,
	in *option.ExerciseReq,
	qty decimal.Decimal,
) *option.ExerciseResp {
	if !sameExerciseRequest(item, in, qty) {
		return &option.ExerciseResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
		}
	}
	return &option.ExerciseResp{
		Base: helper.OkResp(),
		Data: &option.ExerciseData{ExerciseNo: item.ExerciseNo, ExerciseId: item.Id},
	}
}

func isExerciseDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}

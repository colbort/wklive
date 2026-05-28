package logic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/conv"
	"wklive/proto/asset"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessContractLifecycleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessContractLifecycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessContractLifecycleLogic {
	return &ProcessContractLifecycleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 期权合约生命周期处理（状态流转/订单过期/自动行权/到期结算）
func (l *ProcessContractLifecycleLogic) ProcessContractLifecycle(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return runOptionTaskWithLock(l.ctx, l.svcCtx, "process_contract_lifecycle", func() (*option.OptionTaskResp, error) {
		now := time.Now().Unix()
		if err := l.syncContracts(option.ContractStatus_CONTRACT_STATUS_PENDING, now, 0, option.ContractStatus_CONTRACT_STATUS_TRADING, now); err != nil {
			return nil, err
		}
		if err := l.syncContracts(option.ContractStatus_CONTRACT_STATUS_TRADING, 0, now, option.ContractStatus_CONTRACT_STATUS_EXPIRED, now); err != nil {
			return nil, err
		}
		if err := l.processExpiredContracts(now); err != nil {
			return nil, err
		}
		return okOptionTaskResp(), nil
	})
}

func (l *ProcessContractLifecycleLogic) syncContracts(from option.ContractStatus, listEnd, expireEnd int64, to option.ContractStatus, now int64) error {
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status:        int64(from),
			ListTimeEnd:   listEnd,
			ExpireTimeEnd: expireEnd,
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			item.Status = int64(to)
			item.UpdateTimes = now
			if err := l.svcCtx.OptionContractModel.Update(l.ctx, item); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) processExpiredContracts(now int64) error {
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status: int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, contract := range contracts {
			cursor = contract.Id
			if err := l.expireContractOrders(contract, now); err != nil {
				return err
			}
			if contract.IsAutoExercise == int64(option.YesNo_YES_NO_YES) {
				if err := l.autoExerciseContract(contract, now); err != nil {
					return err
				}
			}
			if err := l.settleContract(contract, now); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) expireContractOrders(contract *models.TOptionContract, now int64) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
			ContractId: contract.Id,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			if order.MarginAmount > 0 {
				resp, err := l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{
					TenantId:      order.TenantId,
					TargetBizType: asset.BizType_BIZ_TYPE_OPTION,
					TargetBizNo:   order.OrderNo,
					Amount:        conv.FloatString(order.MarginAmount),
					BizType:       asset.BizType_BIZ_TYPE_OPTION,
					SceneType:     asset.SceneType_SCENE_TYPE_CANCEL_ORDER,
					BizId:         order.Id,
					BizNo:         order.OrderNo,
					Remark:        "option expired order unfreeze",
				})
				if err != nil {
					return err
				}
				if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
					continue
				}
			}
			order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRED)
			order.CancelReason = "CONTRACT_EXPIRED"
			order.CancelTime = now
			order.UpdateTimes = now
			err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionOrderModel)
				positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionPositionModel)
				if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
					return err
				}
				return orderModel.Update(ctx, order)
			})
			if err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) autoExerciseContract(contract *models.TOptionContract, now int64) error {
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	deliveryPrice := 0.0
	if market != nil {
		deliveryPrice = market.UnderlyingPrice
	}
	intrinsicValue := optionIntrinsicValue(contract, deliveryPrice)
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.OptionPositionModel.FindPage(l.ctx, models.OptionPositionPageFilter{
			ContractId: contract.Id,
			Status:     int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if position.Side != int64(option.PositionSide_POSITION_SIDE_LONG) {
				continue
			}
			if position.ExerciseableQty <= 0 || intrinsicValue <= optionFloatEpsilon {
				position.Status = int64(option.PositionStatus_POSITION_STATUS_EXPIRED)
				position.ExerciseableQty = 0
				position.UpdateTimes = now
				if err := l.svcCtx.OptionPositionModel.Update(l.ctx, position); err != nil {
					return err
				}
				continue
			}
			exists, _, err := l.svcCtx.OptionExerciseModel.FindPage(l.ctx, models.OptionExercisePageFilter{
				TenantId:   position.TenantId,
				PositionId: position.Id,
			}, 0, 1)
			if err != nil {
				return err
			}
			if len(exists) > 0 {
				continue
			}
			exerciseNo, err := l.svcCtx.GenerateBizNo(l.ctx, "EX")
			if err != nil {
				return err
			}
			_, err = l.svcCtx.OptionExerciseModel.Insert(l.ctx, &models.TOptionExercise{
				TenantId:        position.TenantId,
				ExerciseNo:      exerciseNo,
				UserId:          position.UserId,
				AccountId:       position.AccountId,
				ContractId:      contract.Id,
				PositionId:      position.Id,
				ExerciseType:    int64(option.ExerciseType_EXERCISE_TYPE_AUTO),
				ExerciseQty:     position.ExerciseableQty,
				StrikePrice:     contract.StrikePrice,
				SettlementPrice: deliveryPrice,
				ExerciseAmount:  optionExerciseAmount(contract, position.ExerciseableQty),
				ProfitAmount:    optionSettlementPayoff(contract, deliveryPrice, position.ExerciseableQty),
				FeeCoin:         contract.SettleCoin,
				Status:          int64(option.ExerciseStatus_EXERCISE_STATUS_DONE),
				Remark:          "option auto exercise task",
				ExerciseTime:    now,
				FinishTime:      now,
				CreateTimes:     now,
				UpdateTimes:     now,
			})
			if err != nil && !errors.Is(err, models.ErrNotFound) {
				return err
			}
			position.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
			position.ExerciseableQty = 0
			position.UpdateTimes = now
			if err := l.svcCtx.OptionPositionModel.Update(l.ctx, position); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) settleContract(contract *models.TOptionContract, now int64) error {
	_, err := l.svcCtx.OptionSettlementModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	deliveryPrice := 0.0
	theoreticalPrice := 0.0
	iv := 0.0
	isITM := int64(option.YesNo_YES_NO_NO)
	if market != nil {
		deliveryPrice = market.UnderlyingPrice
		theoreticalPrice = market.TheoreticalPrice
		iv = market.Iv
		if (contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) && deliveryPrice > contract.StrikePrice) ||
			(contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) && deliveryPrice < contract.StrikePrice) {
			isITM = int64(option.YesNo_YES_NO_YES)
		}
	}
	exerciseResult := int64(option.ExerciseResult_EXERCISE_RESULT_NONE)
	if contract.IsAutoExercise == int64(option.YesNo_YES_NO_YES) {
		if isITM == int64(option.YesNo_YES_NO_YES) {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_EXERCISE)
		} else {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_ABANDON)
		}
	}
	settlementNo, err := l.svcCtx.GenerateBizNo(l.ctx, "OPS")
	if err != nil {
		return err
	}
	contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_SETTLED)
	contract.UpdateTimes = now
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		settlementModel := models.NewTOptionSettlementModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionSettlementModel)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionMarketSnapshotModel)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionContractModel)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionPositionModel)
		accountModel := models.NewTOptionAccountModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionAccountModel)
		billModel := models.NewTOptionBillModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionBillModel)

		result, err := settlementModel.Insert(ctx, &models.TOptionSettlement{
			TenantId:         contract.TenantId,
			SettlementNo:     settlementNo,
			ContractId:       contract.Id,
			UnderlyingSymbol: contract.UnderlyingSymbol,
			ExpireTime:       contract.ExpireTime,
			SettlementTime:   now,
			DeliveryPrice:    deliveryPrice,
			TheoreticalPrice: theoreticalPrice,
			Iv:               iv,
			IsItm:            isITM,
			ExerciseResult:   exerciseResult,
			Status:           int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE),
			Remark:           "option settlement task",
			CreateTimes:      now,
			UpdateTimes:      now,
		})
		if err != nil {
			return err
		}
		settlementId, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if err := settleContractPositions(ctx, positionModel, accountModel, billModel, contract, settlementNo, settlementId, deliveryPrice, now); err != nil {
			return err
		}
		if err := insertMarketSnapshot(ctx, snapshotModel, market, now); err != nil {
			return err
		}
		return contractModel.Update(ctx, contract)
	})
}

func settleContractPositions(ctx context.Context, positionModel models.OptionPositionModel, accountModel models.OptionAccountModel, billModel models.OptionBillModel, contract *models.TOptionContract, settlementNo string, settlementId int64, deliveryPrice float64, now int64) error {
	cursor := int64(0)
	for {
		positions, _, err := positionModel.FindPage(ctx, models.OptionPositionPageFilter{
			ContractId: contract.Id,
			Statuses: []int64{
				int64(option.PositionStatus_POSITION_STATUS_HOLDING),
				int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
				int64(option.PositionStatus_POSITION_STATUS_EXPIRED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			qty := position.PositionQty
			payoff := optionSettlementPayoff(contract, deliveryPrice, qty)
			changeAmount := 0.0
			if position.Side == int64(option.PositionSide_POSITION_SIDE_LONG) {
				if contract.IsAutoExercise == int64(option.YesNo_YES_NO_YES) || position.Status == int64(option.PositionStatus_POSITION_STATUS_EXERCISED) {
					changeAmount = payoff
				}
			} else if position.Side == int64(option.PositionSide_POSITION_SIDE_SHORT) {
				if contract.IsAutoExercise == int64(option.YesNo_YES_NO_YES) {
					changeAmount = -payoff
				}
			}

			position.PositionQty = 0
			position.AvailableQty = 0
			position.FrozenQty = 0
			position.PositionValue = 0
			position.MarginAmount = 0
			position.MaintenanceMargin = 0
			position.UnrealizedPnl = 0
			position.ExerciseableQty = 0
			position.RealizedPnl = normalizeZero(position.RealizedPnl + changeAmount)
			position.Status = int64(option.PositionStatus_POSITION_STATUS_SETTLED)
			position.LastCalcTime = now
			position.UpdateTimes = now
			if err := positionModel.Update(ctx, position); err != nil {
				return err
			}
			if err := applyOptionAccountDelta(ctx, accountModel, billModel, position.TenantId, position.UserId, position.AccountId, contract.SettleCoin, changeAmount, int64(option.BillRefType_BILL_REF_TYPE_SETTLEMENT), settlementId, fmt.Sprintf("%s-P%d", settlementNo, position.Id), "option contract settlement", true, now); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

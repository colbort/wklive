package logic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	spotSettlementBatchSize = int64(100)
	spotSettlementMaxSteps  = 1000
	spotSettlementMaxRetry  = int64(20)
)

type ProcessSpotSettlementsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessSpotSettlementsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessSpotSettlementsLogic {
	return &ProcessSpotSettlementsLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ProcessSpotSettlementsLogic) Process(tenantID int64) error {
	for processed := 0; processed < spotSettlementMaxSteps; {
		now := utils.NowMillis()
		items, err := l.svcCtx.TradeSettlementInstrModel.FindPendingSpot(l.ctx, tenantID, now, spotSettlementBatchSize)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		progressed := false
		for _, item := range items {
			claimed, err := l.svcCtx.TradeSettlementInstrModel.Claim(l.ctx, item.Id, now)
			if err != nil {
				return err
			}
			if !claimed {
				continue
			}
			progressed = true
			processed++
			if err := l.processInstruction(item); err != nil {
				l.Errorf("spot settlement instruction failed, instructionNo=%s fillId=%d orderId=%d err=%v", item.InstructionNo, item.FillId, item.OrderId, err)
				if markErr := l.markFailed(item, err); markErr != nil {
					return markErr
				}
			}
		}
		if !progressed || int64(len(items)) < spotSettlementBatchSize {
			// Query again because completing step N may make step N+1 eligible.
			continue
		}
	}
	return nil
}

func (l *ProcessSpotSettlementsLogic) processInstruction(item *models.TTradeSettlementInstruction) error {
	fill, err := l.svcCtx.TradeFillModel.FindOne(l.ctx, item.FillId)
	if err != nil {
		return err
	}
	order, err := l.svcCtx.TradeOrderModel.FindOne(l.ctx, item.OrderId)
	if err != nil {
		return err
	}
	if fill.TenantId != item.TenantId || order.TenantId != item.TenantId || fill.OrderId != order.Id || fill.ProductType != int64(trade.ProductType_PRODUCT_TYPE_SPOT) {
		return fmt.Errorf("settlement instruction does not match spot fill")
	}
	if fill.SettlementStatus != int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_PROCESSING) {
		fill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_PROCESSING)
		if err := l.svcCtx.TradeFillModel.Update(l.ctx, fill); err != nil {
			return err
		}
	}
	if err := l.executeAssetInstruction(item, fill, order); err != nil {
		return err
	}
	return l.markSucceeded(item, fill, order)
}

func (l *ProcessSpotSettlementsLogic) executeAssetInstruction(item *models.TTradeSettlementInstruction, fill *models.TTradeFill, order *models.TTradeOrder) error {
	walletType := walletTypeForProduct(trade.ProductType_PRODUCT_TYPE_SPOT)
	matchReq := func(scene asset.SceneType) (asset.BizType, asset.SceneType, int64, string) {
		return asset.BizType_BIZ_TYPE_TRADE, scene, fill.Id, item.InstructionNo
	}
	var resp *asset.ChangeAssetResp
	var err error
	switch trade.SettlementInstructionAction(item.Action) {
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill consume frozen"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill release remainder"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CREDIT_AVAILABLE:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_MATCH)
		resp, err = l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill credit"})
	case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE:
		bizType, scene, bizID, bizNo := matchReq(asset.SceneType_SCENE_TYPE_TRADE_FEE)
		if order.Side == int64(common.Side_SIDE_BUY) {
			resp, err = l.svcCtx.AssetClient.DeductFrozenAssetByBizNo(l.ctx, &asset.DeductFrozenAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill fee from frozen"})
		} else {
			resp, err = l.svcCtx.AssetClient.SubAvailable(l.ctx, &asset.SubAvailableReq{TenantId: item.TenantId, UserId: item.UserId, WalletType: walletType, Coin: item.Asset, Amount: item.Amount.String(), BizType: bizType, SceneType: scene, BizId: bizID, BizNo: bizNo, Remark: "spot fill fee from proceeds"})
		}
	default:
		return fmt.Errorf("unsupported spot settlement action: %d", item.Action)
	}
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(l.ctx, resp.Base.Code)
	}
	return nil
}

func (l *ProcessSpotSettlementsLogic) markSucceeded(item *models.TTradeSettlementInstruction, fill *models.TTradeFill, order *models.TTradeOrder) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTBizTradeEventModel(conn, l.svcCtx.Config.CacheRedis)

		current, err := instructionModel.FindOne(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if current.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING) {
			return fmt.Errorf("settlement instruction is not processing")
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.LastErrorMsg = ""
		current.NextRetryAt = 0
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}

		reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, item.TenantId, item.ReservationNo)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return err
		}
		if reservation != nil {
			var ok bool
			switch trade.SettlementInstructionAction(item.Action) {
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_CONSUME_FROZEN:
				ok, err = reservationModel.AddConsumed(ctx, reservation.Id, item.Amount, now)
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_DEDUCT_FEE:
				if order.Side == int64(common.Side_SIDE_BUY) {
					ok, err = reservationModel.AddConsumed(ctx, reservation.Id, item.Amount, now)
				} else {
					ok = true
				}
			case trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN:
				ok, err = reservationModel.AddReleased(ctx, reservation.Id, item.Amount, now)
			default:
				ok = true
			}
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("reservation settlement amount exceeds reserved amount")
			}
		}

		if err := ensureSpotRemainderRelease(ctx, instructionModel, reservationModel, order, fill, now); err != nil {
			return err
		}
		instructions, err := instructionModel.FindByFillId(ctx, fill.TenantId, fill.Id)
		if err != nil {
			return err
		}
		for _, instruction := range instructions {
			if instruction.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
				return nil
			}
		}
		currentFill, err := fillModel.FindOne(ctx, fill.Id)
		if err != nil {
			return err
		}
		if currentFill.SettlementStatus == int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED) {
			return nil
		}
		currentFill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_SETTLED)
		currentFill.SettledAt = now
		if err := fillModel.Update(ctx, currentFill); err != nil {
			return err
		}
		return insertMatchOutboxEvent(ctx, eventModel, order, derivedTradeBizNo(fill.FillNo, "SETTLED"), "SPOT_FILL_SETTLED", fill.FillNo, "fill", "{}", now)
	})
}

func ensureSpotRemainderRelease(ctx context.Context, instructionModel models.TTradeSettlementInstructionModel, reservationModel models.TTradeAssetReservationModel, order *models.TTradeOrder, fill *models.TTradeFill, now int64) error {
	if order.Status != int64(trade.OrderStatus_ORDER_STATUS_FILLED) {
		return nil
	}
	unfinished, err := instructionModel.CountUnfinishedByOrder(ctx, order.TenantId, order.Id)
	if err != nil || unfinished > 0 {
		return err
	}
	reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, order.TenantId, order.OrderNo)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	remaining := reservation.ReservedAmount.Sub(reservation.ConsumedAmount).Sub(reservation.ReleasedAmount)
	if !remaining.IsPositive() {
		return nil
	}
	instructions, err := instructionModel.FindByFillId(ctx, fill.TenantId, fill.Id)
	if err != nil {
		return err
	}
	stepNo := int64(1)
	for _, instruction := range instructions {
		if instruction.StepNo >= stepNo {
			stepNo = instruction.StepNo + 1
		}
	}
	return insertSettlementInstructionIdempotent(ctx, instructionModel, &models.TTradeSettlementInstruction{TenantId: fill.TenantId, InstructionNo: derivedTradeBizNo(fill.FillNo, "RELEASE"), BizType: "fill", BizId: fill.FillNo, BatchNo: fill.MatchNo, FillId: fill.Id, OrderId: order.Id, ReservationNo: order.OrderNo, UserId: order.UserId, Action: int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN), Asset: reservation.Asset, Amount: remaining, StepNo: stepNo, Status: int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PENDING), NextRetryAt: now, CreateTimes: now, UpdateTimes: now})
}

func (l *ProcessSpotSettlementsLogic) markFailed(item *models.TTradeSettlementInstruction, cause error) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		fillModel := models.NewTTradeFillModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := instructionModel.FindOne(ctx, item.Id)
		if err != nil {
			return err
		}
		current.RetryCount++
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)
		if current.RetryCount >= spotSettlementMaxRetry {
			current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)
			current.NextRetryAt = 0
		} else {
			delaySeconds := int64(1) << min(current.RetryCount, int64(10))
			current.NextRetryAt = now + delaySeconds*1000
		}
		current.LastErrorMsg = cause.Error()
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		fill, err := fillModel.FindOne(ctx, item.FillId)
		if err != nil {
			return err
		}
		fill.SettlementStatus = int64(trade.FillSettlementStatus_FILL_SETTLEMENT_STATUS_FAILED)
		fill.SettlementRetryCount++
		return fillModel.Update(ctx, fill)
	})
}

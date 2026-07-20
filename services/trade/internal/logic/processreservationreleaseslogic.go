package logic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const reservationReleaseBatchSize = int64(100)

type ProcessReservationReleasesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessReservationReleasesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessReservationReleasesLogic {
	return &ProcessReservationReleasesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ProcessReservationReleasesLogic) Process(tenantID int64) error {
	for processed := 0; processed < spotSettlementMaxSteps; {
		now := utils.NowMillis()
		items, err := l.svcCtx.TradeSettlementInstrModel.FindPendingOrderReleases(l.ctx, tenantID, now, reservationReleaseBatchSize)
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
			if err := l.executeClaimed(item); err != nil {
				l.Errorf("reservation release failed, instructionNo=%s orderId=%d err=%v", item.InstructionNo, item.OrderId, err)
				if markErr := l.markFailed(item, err); markErr != nil {
					return markErr
				}
			}
		}
		if !progressed {
			return nil
		}
	}
	return nil
}

func (l *ProcessReservationReleasesLogic) ProcessInstruction(instructionID int64) error {
	item, err := l.svcCtx.TradeSettlementInstrModel.FindOne(l.ctx, instructionID)
	if err != nil {
		return err
	}
	if item.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
		return nil
	}
	claimed, err := l.svcCtx.TradeSettlementInstrModel.Claim(l.ctx, item.Id, utils.NowMillis())
	if err != nil || !claimed {
		return err
	}
	if err := l.executeClaimed(item); err != nil {
		if markErr := l.markFailed(item, err); markErr != nil {
			return markErr
		}
		return err
	}
	return nil
}

func (l *ProcessReservationReleasesLogic) executeClaimed(item *models.TTradeSettlementInstruction) error {
	if item.BizType != "order" || item.Action != int64(trade.SettlementInstructionAction_SETTLEMENT_INSTRUCTION_ACTION_RELEASE_FROZEN) || item.ReservationNo == "" || !item.Amount.IsPositive() {
		return fmt.Errorf("invalid order reservation release instruction")
	}
	resp, err := l.svcCtx.AssetClient.UnfreezeAssetByBizNo(l.ctx, &asset.UnfreezeAssetByBizNoReq{TenantId: item.TenantId, TargetBizType: asset.BizType_BIZ_TYPE_TRADE, TargetBizNo: item.ReservationNo, Amount: item.Amount.String(), BizType: asset.BizType_BIZ_TYPE_TRADE, SceneType: asset.SceneType_SCENE_TYPE_CANCEL_ORDER, BizId: item.OrderId, BizNo: item.InstructionNo, Remark: "trade order reservation release"})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	if resp.Base.Code != 200 {
		return i18n.StatusError(l.ctx, resp.Base.Code)
	}
	return l.markSucceeded(item)
}

func (l *ProcessReservationReleasesLogic) markSucceeded(item *models.TTradeSettlementInstruction) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := instructionModel.FindOne(ctx, item.Id)
		if err != nil {
			return err
		}
		if current.Status == int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS) {
			return nil
		}
		if current.Status != int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_PROCESSING) {
			return fmt.Errorf("reservation release instruction is not processing")
		}
		reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, item.TenantId, item.ReservationNo)
		if err != nil {
			return err
		}
		ok, err := reservationModel.AddReleased(ctx, reservation.Id, current.Amount, now)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("reservation release exceeds remaining amount")
		}
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_SUCCESS)
		current.NextRetryAt = 0
		current.LastErrorMsg = ""
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		_, err = finalizeOrderTermination(ctx, conn, l.svcCtx, item.OrderId, now)
		return err
	})
}

func (l *ProcessReservationReleasesLogic) markFailed(item *models.TTradeSettlementInstruction, cause error) error {
	now := utils.NowMillis()
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		instructionModel := models.NewTTradeSettlementInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		reservationModel := models.NewTTradeAssetReservationModel(conn, l.svcCtx.Config.CacheRedis)
		current, err := instructionModel.FindOne(ctx, item.Id)
		if err != nil {
			return err
		}
		current.RetryCount++
		terminal := current.RetryCount >= spotSettlementMaxRetry
		current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_FAILED)
		current.NextRetryAt = now + (int64(1)<<min(current.RetryCount, int64(10)))*1000
		if terminal {
			current.Status = int64(trade.SettlementInstructionStatus_SETTLEMENT_INSTRUCTION_STATUS_MANUAL_REVIEW)
			current.NextRetryAt = 0
		}
		current.LastErrorMsg = cause.Error()
		current.UpdateTimes = now
		if err := instructionModel.Update(ctx, current); err != nil {
			return err
		}
		reservation, err := reservationModel.FindOneByTenantIdReservationNo(ctx, item.TenantId, item.ReservationNo)
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		return reservationModel.MarkSettlementFailure(ctx, reservation.Id, int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_RELEASING), terminal, current.NextRetryAt, cause.Error(), now)
	})
}

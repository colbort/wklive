package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RetryPhysicalDeliveryUnitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetryPhysicalDeliveryUnitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryPhysicalDeliveryUnitLogic {
	return &RetryPhysicalDeliveryUnitLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工确认补资后，按原指令号重试一个实物交割单元
func (l *RetryPhysicalDeliveryUnitLogic) RetryPhysicalDeliveryUnit(in *option.RetryPhysicalDeliveryUnitReq) (*option.CommonResp, error) {
	reason := strings.TrimSpace(in.Reason)
	if in.DeliveryUnitId <= 0 || reason == "" {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	item, err := l.svcCtx.OptionPhysicalDeliveryUnitModel.FindOne(l.ctx, in.DeliveryUnitId)
	if errors.Is(err, models.ErrNotFound) ||
		(err == nil && in.TenantId > 0 && item.TenantId != in.TenantId) {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		unitModel := models.NewTOptionPhysicalDeliveryUnitModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		locked, lockErr := unitModel.FindOneForUpdate(ctx, item.Id)
		if lockErr != nil {
			return lockErr
		}
		if locked.Status != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_MANUAL_REVIEW) &&
			locked.Status != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_DEFAULTED) {
			return errors.New("physical delivery unit is not in manual retry state")
		}
		now := time.Now().Unix()
		affected, resetErr := instructionModel.ResetFailedByDeliveryUnit(
			ctx, locked.TenantId, locked.Id, now,
		)
		if resetErr != nil {
			return resetErr
		}
		if affected == 0 {
			return errors.New("physical delivery unit has no failed instruction to retry")
		}
		locked.Status = int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_ASSET_PROCESSING)
		locked.ManualRetryCount++
		locked.FailedInstructionId = 0
		locked.LastErrorMsg = ""
		locked.UpdateTimes = now
		if err := unitModel.Update(ctx, locked); err != nil {
			return err
		}
		units, err := unitModel.FindByBatch(ctx, locked.TenantId, locked.BatchId)
		if err != nil {
			return err
		}
		hasOtherException := false
		for _, unit := range units {
			if unit.Id == locked.Id {
				continue
			}
			if unit.Status == int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_MANUAL_REVIEW) ||
				unit.Status == int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_DEFAULTED) {
				hasOtherException = true
				break
			}
		}
		if !hasOtherException {
			batchModel := models.NewTOptionSettlementBatchModel(conn, l.svcCtx.Config.CacheRedis)
			batch, err := batchModel.FindOneByTenantIdBatchNo(
				ctx, locked.TenantId, locked.BatchNo,
			)
			if err != nil {
				return err
			}
			batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_ASSET_PROCESSING)
			batch.LastErrorMsg = ""
			batch.UpdateTimes = now
			if err := batchModel.Update(ctx, batch); err != nil {
				return err
			}
			settlementModel := models.NewTOptionSettlementModel(conn, l.svcCtx.Config.CacheRedis)
			settlement, err := settlementModel.FindOneByTenantIdSettlementNo(
				ctx, locked.TenantId, locked.BatchNo,
			)
			if err != nil {
				return err
			}
			settlement.Status = int64(option.SettlementStatus_SETTLEMENT_STATUS_PROCESSING)
			settlement.Remark = fmt.Sprintf(
				"physical delivery unit %s manually retried", locked.DeliveryUnitNo,
			)
			settlement.UpdateTimes = now
			if err := settlementModel.Update(ctx, settlement); err != nil {
				return err
			}
		}
		_, err = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: locked.TenantId, ContractId: locked.ContractId,
			EventType: "PHYSICAL_DELIVERY_MANUAL_RETRY", Reason: reason,
			Detail: fmt.Sprintf(
				"deliveryUnit=%s retryCount=%d resetInstructions=%d",
				locked.DeliveryUnitNo, locked.ManualRetryCount, affected,
			),
			OperatorId: operatorID, CreateTimes: now,
		})
		return err
	})
	if err != nil {
		if strings.Contains(err.Error(), "manual retry state") ||
			strings.Contains(err.Error(), "no failed instruction") {
			return &option.CommonResp{
				Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
			}, nil
		}
		return nil, err
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}

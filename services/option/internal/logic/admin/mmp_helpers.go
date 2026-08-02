package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type mmpConfigInput struct {
	tenantID            int64
	userID              int64
	contractID          int64
	groupCode           string
	enabled             int64
	qtyThreshold        decimal.Decimal
	tradeCountThreshold int64
	lossThreshold       decimal.Decimal
	windowSeconds       int64
	cooldownSeconds     int64
	reason              string
	operatorID          int64
}

var errMMPNotTriggered = errors.New("MMP_NOT_TRIGGERED")
var errMMPReleasePending = errors.New("MMP_RELEASE_PENDING")

func validateMMPConfigInput(
	ctx context.Context, in *option.UpsertMMPConfigReq,
) (*mmpConfigInput, *option.GetMMPConfigResp, error) {
	operatorID, err := utils.GetUserIdFromMd(ctx)
	if err != nil {
		return nil, nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(ctx, in.TenantId)
	if err != nil {
		return nil, nil, err
	}
	if forbidden || !allowed {
		return nil, &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)),
		}, nil
	}
	groupCode, validGroup := applogic.NormalizeMMPGroup(in.GroupCode)
	reason := strings.TrimSpace(in.Reason)
	qtyThreshold, qtyErr := conv.ParseDecimalField(in.QtyThreshold)
	lossThreshold, lossErr := conv.ParseDecimalField(in.LossThreshold)
	validEnabled := in.Enabled == common.YesNo_YES_NO_YES || in.Enabled == common.YesNo_YES_NO_NO
	if in.TenantId <= 0 || in.UserId <= 0 || in.ContractId <= 0 || operatorID <= 0 ||
		!validGroup || !validEnabled || reason == "" || len(reason) > 500 ||
		qtyErr != nil || lossErr != nil ||
		qtyThreshold.IsNegative() || lossThreshold.IsNegative() ||
		qtyThreshold.Exponent() < -16 || lossThreshold.Exponent() < -16 ||
		in.TradeCountThreshold < 0 ||
		in.WindowSeconds < 1 || in.WindowSeconds > 3600 ||
		in.CooldownSeconds < 0 || in.CooldownSeconds > 86400 {
		return nil, &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
		}, nil
	}
	if in.Enabled == common.YesNo_YES_NO_YES &&
		!qtyThreshold.IsPositive() && in.TradeCountThreshold == 0 && !lossThreshold.IsPositive() {
		return nil, &option.GetMMPConfigResp{
			Base: helper.ErrResp(i18n.ParamError, "at least one MMP threshold must be positive"),
		}, nil
	}
	return &mmpConfigInput{
		tenantID: in.TenantId, userID: in.UserId, contractID: in.ContractId,
		groupCode: groupCode, enabled: int64(in.Enabled),
		qtyThreshold: qtyThreshold, tradeCountThreshold: in.TradeCountThreshold,
		lossThreshold: lossThreshold, windowSeconds: in.WindowSeconds,
		cooldownSeconds: in.CooldownSeconds, reason: reason, operatorID: operatorID,
	}, nil, nil
}

func stageMMPConfig(
	ctx context.Context, svcCtx *svc.ServiceContext, input *mmpConfigInput,
) (*models.TOptionMmpConfig, error) {
	now := time.Now().Unix()
	var result *models.TOptionMmpConfig
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
		contract, err := contractModel.FindOneForUpdate(txCtx, input.contractID)
		if err != nil {
			return err
		}
		if contract.TenantId != input.tenantID {
			return models.ErrNotFound
		}
		model := models.NewTOptionMmpConfigModel(conn, svcCtx.Config.CacheRedis)
		item, err := model.FindForUpdate(
			txCtx, input.tenantID, input.userID, input.contractID, input.groupCode,
		)
		if errors.Is(err, models.ErrNotFound) {
			item = &models.TOptionMmpConfig{
				TenantId: input.tenantID, UserId: input.userID,
				ContractId: input.contractID, GroupCode: input.groupCode,
				CreatedBy: input.operatorID, CreateTimes: now,
			}
		} else if err != nil {
			return err
		}
		item.Enabled = input.enabled
		item.QtyThreshold = input.qtyThreshold
		item.TradeCountThreshold = input.tradeCountThreshold
		item.LossThreshold = input.lossThreshold
		item.WindowSeconds = input.windowSeconds
		item.CooldownSeconds = input.cooldownSeconds
		item.Status = int64(option.MMPStatus_MMP_STATUS_DISABLED)
		item.WindowStart = now
		item.AccumulatedQty = decimal.Zero
		item.TradeCount = 0
		item.AccumulatedLoss = decimal.Zero
		item.TriggeredAt = 0
		item.CooldownUntil = 0
		item.TriggerReason = ""
		item.LastErrorMsg = ""
		item.UpdatedBy = input.operatorID
		item.UpdateTimes = now
		if item.Id == 0 {
			insertResult, err := model.Insert(txCtx, item)
			if err != nil {
				return err
			}
			item.Id, err = insertResult.LastInsertId()
			if err != nil {
				return err
			}
		} else if err := model.Update(txCtx, item); err != nil {
			return err
		}
		eventModel := models.NewTOptionTradingControlEventModel(conn, svcCtx.Config.CacheRedis)
		if _, err := eventModel.Insert(txCtx, &models.TOptionTradingControlEvent{
			TenantId: input.tenantID, UserId: input.userID, ContractId: input.contractID,
			EventType: "MMP_CONFIGURED", Reason: input.reason,
			Detail: fmt.Sprintf(
				"group=%s enabled=%d qty=%s count=%d loss=%s window=%d cooldown=%d",
				input.groupCode, input.enabled, input.qtyThreshold, input.tradeCountThreshold,
				input.lossThreshold, input.windowSeconds, input.cooldownSeconds,
			),
			OperatorId: input.operatorID, CreateTimes: now,
		}); err != nil {
			return err
		}
		copyItem := *item
		result = &copyItem
		return nil
	})
	return result, err
}

func activateStagedMMPConfig(
	ctx context.Context, svcCtx *svc.ServiceContext, input *mmpConfigInput, eventType string,
) (*models.TOptionMmpConfig, error) {
	now := time.Now().Unix()
	var result *models.TOptionMmpConfig
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
		if _, err := contractModel.FindOneForUpdate(txCtx, input.contractID); err != nil {
			return err
		}
		model := models.NewTOptionMmpConfigModel(conn, svcCtx.Config.CacheRedis)
		item, err := model.FindForUpdate(
			txCtx, input.tenantID, input.userID, input.contractID, input.groupCode,
		)
		if err != nil {
			return err
		}
		if eventType == "MMP_MANUAL_RESET" &&
			item.Status != int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
			return errMMPNotTriggered
		}
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		unsafeOrder, unsafeErr := orderModel.FindFirstUnsafeMMPOrderForUpdate(
			txCtx, input.tenantID, input.userID, input.contractID, input.groupCode,
		)
		if unsafeErr != nil && !errors.Is(unsafeErr, models.ErrNotFound) {
			return unsafeErr
		}
		if unsafeOrder != nil {
			return fmt.Errorf(
				"%w: order_id=%d status=%d",
				errMMPReleasePending, unsafeOrder.Id, unsafeOrder.Status,
			)
		}
		if item.Enabled == int64(common.YesNo_YES_NO_YES) {
			item.Status = int64(option.MMPStatus_MMP_STATUS_ACTIVE)
		} else {
			item.Status = int64(option.MMPStatus_MMP_STATUS_DISABLED)
		}
		item.WindowStart = now
		item.AccumulatedQty = decimal.Zero
		item.TradeCount = 0
		item.AccumulatedLoss = decimal.Zero
		item.TriggeredAt = 0
		item.CooldownUntil = 0
		item.TriggerReason = ""
		item.LastErrorMsg = ""
		item.UpdatedBy = input.operatorID
		item.UpdateTimes = now
		if err := model.Update(txCtx, item); err != nil {
			return err
		}
		if eventType != "" {
			eventModel := models.NewTOptionTradingControlEventModel(conn, svcCtx.Config.CacheRedis)
			if _, err := eventModel.Insert(txCtx, &models.TOptionTradingControlEvent{
				TenantId: input.tenantID, UserId: input.userID, ContractId: input.contractID,
				EventType: eventType, Reason: input.reason,
				Detail:     fmt.Sprintf("group=%s config_id=%d", input.groupCode, item.Id),
				OperatorId: input.operatorID, CreateTimes: now,
			}); err != nil {
				return err
			}
		}
		copyItem := *item
		result = &copyItem
		return nil
	})
	return result, err
}

func stageMMPReset(
	ctx context.Context, svcCtx *svc.ServiceContext, input *mmpConfigInput,
) error {
	return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
		contract, err := contractModel.FindOneForUpdate(txCtx, input.contractID)
		if err != nil {
			return err
		}
		if contract.TenantId != input.tenantID {
			return models.ErrNotFound
		}
		model := models.NewTOptionMmpConfigModel(conn, svcCtx.Config.CacheRedis)
		item, err := model.FindForUpdate(
			txCtx, input.tenantID, input.userID, input.contractID, input.groupCode,
		)
		if err != nil {
			return err
		}
		if item.Enabled != int64(common.YesNo_YES_NO_YES) ||
			item.Status != int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
			return errMMPNotTriggered
		}
		item.LastErrorMsg = ""
		item.UpdatedBy = input.operatorID
		item.UpdateTimes = time.Now().Unix()
		return model.Update(txCtx, item)
	})
}

func mmpConfigResponse(item *models.TOptionMmpConfig) *option.GetMMPConfigResp {
	return &option.GetMMPConfigResp{Base: helper.OkResp(), Data: helpers.ToMMPConfigProto(item)}
}

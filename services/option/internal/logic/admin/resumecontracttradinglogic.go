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
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ResumeContractTradingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResumeContractTradingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResumeContractTradingLogic {
	return &ResumeContractTradingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 引用原 halt 解除临时休市
func (l *ResumeContractTradingLogic) ResumeContractTrading(in *option.ResumeContractTradingReq) (*option.GetTradingHaltResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return haltPermissionDenied(l.ctx), nil
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.HaltId <= 0 || operatorID <= 0 || reason == "" || len(reason) > 500 {
		return haltParamError(l.ctx), nil
	}
	now := time.Now().Unix()
	var resumed *models.TOptionTradingHalt
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		halt, findErr := haltModel.FindOneForUpdate(ctx, in.HaltId)
		if findErr != nil {
			return findErr
		}
		if halt.TenantId != in.TenantId ||
			halt.Status != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE) {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		contract, findErr := contractModel.FindOneForUpdate(ctx, halt.ContractId)
		if findErr != nil {
			return findErr
		}
		if contract.TenantId != halt.TenantId ||
			contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) ||
			now < contract.ListTime || contract.LastTradeTime <= 0 || now >= contract.LastTradeTime {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if contract.LiquidationDeficitPolicy == int64(
			option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_PLATFORM_BACKSTOP,
		) && !l.svcCtx.Config.PlatformBackstop.Enabled {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if ready, _ := helpers.ContractLaunchProductScopeReady(
			contract, l.svcCtx.Config.ProductScope,
		); !ready {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		activeOrders, findErr := orderModel.HasUnsafeContractResumeOrders(ctx, contract.TenantId, contract.Id)
		if findErr != nil {
			return findErr
		}
		if activeOrders {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		code, valid := helpers.NormalizeTradingCalendarCode(contract.TradingCalendarCode)
		if !valid {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if _, findErr = calendarModel.FindEffective(ctx, contract.TenantId, code, now); findErr != nil {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		halt.Status = int64(option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED)
		halt.ActiveKey = "HALT:" + halt.HaltNo
		halt.LiftedAt = now
		halt.LiftedBy = operatorID
		halt.LiftReason = reason
		halt.UpdateTimes = now
		if updateErr := haltModel.Update(ctx, halt); updateErr != nil {
			return updateErr
		}
		contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_TRADING)
		contract.UpdateTimes = now
		if updateErr := contractModel.Update(ctx, contract); updateErr != nil {
			return updateErr
		}
		if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: contract.TenantId, ContractId: contract.Id, EventType: "CONTRACT_TRADING_RESUMED",
			Reason: reason, Detail: fmt.Sprintf("haltNo=%s", halt.HaltNo),
			OperatorId: operatorID, CreateTimes: now,
		}); insertErr != nil {
			return insertErr
		}
		resumed = halt
		return nil
	})
	if errors.Is(err, models.ErrNotFound) {
		return haltContractNotFound(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetTradingHaltResp{Base: helper.OkResp(), Data: helpers.ToTradingHaltProto(resumed)}, nil
}

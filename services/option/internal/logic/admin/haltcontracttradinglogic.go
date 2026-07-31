package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/generate"
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

type HaltContractTradingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewHaltContractTradingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HaltContractTradingLogic {
	return &HaltContractTradingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 紧急暂停合约交易并撤销活动订单
func (l *HaltContractTradingLogic) HaltContractTrading(in *option.HaltContractTradingReq) (*option.GetTradingHaltResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if errors.Is(err, models.ErrNotFound) ||
		(err == nil && in.TenantId > 0 && contract.TenantId != in.TenantId) {
		return haltContractNotFound(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, contract.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return haltPermissionDenied(l.ctx), nil
	}
	reason := strings.TrimSpace(in.Reason)
	evidenceRef := strings.TrimSpace(in.EvidenceRef)
	if operatorID <= 0 || reason == "" || evidenceRef == "" || len(reason) > 500 || len(evidenceRef) > 500 {
		return haltParamError(l.ctx), nil
	}
	haltNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "option_halt_id", "OH", "")
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var halt *models.TOptionTradingHalt
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		locked, findErr := contractModel.FindOneForUpdate(ctx, contract.Id)
		if findErr != nil {
			return findErr
		}
		if locked.TenantId != contract.TenantId ||
			locked.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		if _, findErr = haltModel.FindActiveByContract(ctx, locked.TenantId, locked.Id); findErr == nil {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		halt = &models.TOptionTradingHalt{
			TenantId: locked.TenantId, HaltNo: haltNo, ActiveKey: fmt.Sprintf("CONTRACT:%d", locked.Id),
			ContractId: locked.Id, Source: int64(option.TradingHaltSource_TRADING_HALT_SOURCE_MANUAL),
			Status: int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE),
			Reason: reason, EvidenceRef: evidenceRef, StartedAt: now, CreatedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := haltModel.Insert(ctx, halt)
		if insertErr != nil {
			return insertErr
		}
		halt.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}
		locked.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
		locked.UpdateTimes = now
		if updateErr := contractModel.Update(ctx, locked); updateErr != nil {
			return updateErr
		}
		_, insertErr = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: locked.TenantId, ContractId: locked.Id, EventType: "CONTRACT_TRADING_HALTED",
			Reason: reason, Detail: fmt.Sprintf("haltNo=%s evidence=%s", haltNo, evidenceRef),
			OperatorId: operatorID, CreateTimes: now,
		})
		return insertErr
	})
	if err != nil {
		return nil, err
	}

	cancelLogic := NewForceCancelContractOrdersLogic(l.ctx, l.svcCtx)
	total, success, failed, cancelErr := cancelLogic.forceCancelAll(contract, "TRADING_HALT:"+haltNo, true)
	lastError := ""
	if cancelErr != nil {
		lastError = cancelErr.Error()
		if len(lastError) > 1000 {
			lastError = lastError[:1000]
		}
		l.Errorf(
			"option trading halt left orders requiring retry, tenantId=%d contractId=%d haltId=%d failed=%d err=%v",
			contract.TenantId, contract.Id, halt.Id, failed, cancelErr,
		)
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		locked, findErr := haltModel.FindOneForUpdate(ctx, halt.Id)
		if findErr != nil {
			return findErr
		}
		locked.CancelTotal = total
		locked.CancelSuccess = success
		locked.CancelFailed = failed
		locked.LastErrorMsg = lastError
		locked.UpdateTimes = time.Now().Unix()
		if updateErr := haltModel.Update(ctx, locked); updateErr != nil {
			return updateErr
		}
		halt = locked
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &option.GetTradingHaltResp{Base: helper.OkResp(), Data: helpers.ToTradingHaltProto(halt)}, nil
}

func haltParamError(ctx context.Context) *option.GetTradingHaltResp {
	return &option.GetTradingHaltResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}

func haltPermissionDenied(ctx context.Context) *option.GetTradingHaltResp {
	return &option.GetTradingHaltResp{
		Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)),
	}
}

func haltContractNotFound(ctx context.Context) *option.GetTradingHaltResp {
	return &option.GetTradingHaltResp{
		Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, ctx)),
	}
}

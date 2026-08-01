package adminlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ForceCancelContractOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewForceCancelContractOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ForceCancelContractOrdersLogic {
	return &ForceCancelContractOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 暂停合约后强制撤销全部活动订单
func (l *ForceCancelContractOrdersLogic) ForceCancelContractOrders(in *option.ForceCancelContractOrdersReq) (*option.CommonResp, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if errors.Is(err, models.ErrNotFound) || (err == nil && in.TenantId > 0 && contract.TenantId != in.TenantId) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, contract.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx))}, nil
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "ADMIN_FORCE_CANCEL"
	}
	total, success, failed, cancelErr := l.forceCancelAll(contract, reason, false)
	if recordErr := l.recordActiveHaltCancelAttempt(contract, total, success, failed, cancelErr); recordErr != nil {
		return nil, recordErr
	}
	if cancelErr != nil {
		return nil, cancelErr
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}

func (l *ForceCancelContractOrdersLogic) recordActiveHaltCancelAttempt(
	contract *models.TOptionContract,
	total, success, failed int64,
	cancelErr error,
) error {
	halt, err := l.svcCtx.OptionTradingHaltModel.FindActiveByContract(
		l.ctx, contract.TenantId, contract.Id,
	)
	if errors.Is(err, models.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	lastError := ""
	if cancelErr != nil {
		lastError = cancelErr.Error()
		if len(lastError) > 500 {
			lastError = lastError[:500]
		}
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		locked, findErr := haltModel.FindOneForUpdate(ctx, halt.Id)
		if findErr != nil {
			return findErr
		}
		if locked.Status != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE) {
			return nil
		}
		locked.CancelTotal += total
		locked.CancelSuccess += success
		locked.CancelFailed += failed
		locked.LastErrorMsg = lastError
		locked.UpdateTimes = time.Now().Unix()
		return haltModel.Update(ctx, locked)
	})
}

func (l *ForceCancelContractOrdersLogic) forceCancelAll(
	contract *models.TOptionContract,
	reason string,
	continueOnError bool,
) (total, success, failed int64, lastErr error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	if operatorID <= 0 {
		return 0, 0, 0, errors.New("authenticated operator is required for force cancel")
	}
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
			TenantId: contract.TenantId, ContractId: contract.Id,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_FUNDING),
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return total, success, failed, err
		}
		for _, order := range orders {
			cursor = order.Id
			total++
			if err := l.forceCancelOne(order.Id, reason, operatorID); err != nil {
				failed++
				lastErr = err
				if !continueOnError {
					return total, success, failed, err
				}
				continue
			}
			success++
		}
		if len(orders) < 100 {
			return total, success, failed, lastErr
		}
	}
}

func (l *ForceCancelContractOrdersLogic) forceCancelOne(
	orderID int64,
	reason string,
	operatorID int64,
) error {
	_, err := applogic.CancelOrderByControlWithAudit(
		l.ctx, l.svcCtx, orderID, reason, applogic.ControlCancelAudit{
			EventType:  "ADMIN_FORCE_CANCEL_ORDER",
			OperatorID: operatorID,
			Detail:     "administrative contract-wide force cancel",
		},
	)
	return err
}

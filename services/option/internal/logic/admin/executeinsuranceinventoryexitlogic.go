package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ExecuteInsuranceInventoryExitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExecuteInsuranceInventoryExitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExecuteInsuranceInventoryExitLogic {
	return &ExecuteInsuranceInventoryExitLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ExecuteInsuranceInventoryExitLogic) ExecuteInsuranceInventoryExit(
	in *option.ExecuteInsuranceInventoryExitReq,
) (*option.GetInsuranceInventoryExitResp, error) {
	limits, limitErr := insuranceInventoryExitRuntimeLimits(l.svcCtx)
	if limitErr != nil {
		return insuranceInventoryExitError(l.ctx, limitErr), nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	if in.TenantId <= 0 || in.ExitId <= 0 || operatorID <= 0 {
		return &option.GetInsuranceInventoryExitResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.GetInsuranceInventoryExitResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	item, err := l.svcCtx.OptionInsuranceInventoryExitModel.FindOne(l.ctx, in.ExitId)
	if err != nil {
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return insuranceInventoryExitError(l.ctx, errInvalidInsuranceInventoryExit), nil
	}
	if item.Status == int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_SUBMITTED) {
		return &option.GetInsuranceInventoryExitResp{Base: helper.OkResp(), Data: helpers.ToInsuranceInventoryExitProto(item)}, nil
	}
	if item.Status != int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_APPROVED) {
		return insuranceInventoryExitError(l.ctx, errInvalidInsuranceInventoryExit), nil
	}
	clientOrderID := insuranceInventoryExitClientOrderID(item.Id)
	order, findErr := l.svcCtx.OptionOrderModel.FindOneByTenantIdUserIdClientOrderId(
		l.ctx, item.TenantId, item.InsuranceUserId, clientOrderID,
	)
	if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	if errors.Is(findErr, models.ErrNotFound) {
		position, validationErr := l.svcCtx.OptionPositionModel.FindOne(l.ctx, item.PositionId)
		if validationErr != nil {
			return nil, validationErr
		}
		contract, validationErr := l.svcCtx.OptionContractModel.FindOne(l.ctx, item.ContractId)
		if validationErr != nil {
			return nil, validationErr
		}
		market, validationErr := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(
			l.ctx, item.TenantId, item.ContractId,
		)
		if validationErr != nil {
			return nil, validationErr
		}
		if validationErr = validateInsuranceInventoryExit(
			position, contract, market, item.Quantity, item.LimitPrice, time.Now().Unix(),
		); validationErr == nil {
			validationErr = validateInsuranceInventoryExitRuntimeLimits(
				contract, market, item.Quantity, item.LimitPrice, limits,
			)
		}
		if validationErr == nil {
			reserved, reserveErr := l.svcCtx.OptionInsuranceInventoryExitModel.SumReservedQuantity(
				l.ctx, item.TenantId, item.ContractId, insuranceInventoryExitUTCDayStart(time.Now()),
			)
			if reserveErr != nil {
				return nil, reserveErr
			}
			if reserved.GreaterThan(limits.maxDailyQuantity) {
				validationErr = fmt.Errorf("%w: daily quantity limit exceeded", errInsuranceInventoryExitLimits)
			}
		}
		if validationErr == nil {
			orders, depthErr := l.svcCtx.OptionOrderModel.FindAllMatchableOrders(
				l.ctx, item.TenantId, item.ContractId, int64(common.Side_SIDE_SELL),
				item.InsuranceUserId, item.InsuranceAccountId, item.LimitPrice,
			)
			if depthErr != nil {
				return nil, depthErr
			}
			validationErr = validateInsuranceInventoryExitOrderBookDepth(orders, limits)
		}
		if validationErr != nil {
			saveInsuranceInventoryExitFailure(l.ctx, l.svcCtx, item.Id, validationErr)
			return insuranceInventoryExitError(l.ctx, validationErr), nil
		}
		userCtx := insuranceInventoryExitUserContext(l.ctx, item.TenantId, item.InsuranceUserId)
		resp, placeErr := applogic.NewAdministrativePlaceOrderLogic(userCtx, l.svcCtx).PlaceOrder(&option.PlaceOrderReq{
			AccountId: item.InsuranceAccountId, ContractId: item.ContractId,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_CLOSE,
			OrderType: option.OrderType_ORDER_TYPE_IOC, Price: item.LimitPrice.String(),
			Qty: item.Quantity.String(), ClientOrderId: clientOrderID,
			ReduceOnly: common.YesNo_YES_NO_YES,
		})
		if placeErr != nil {
			saveInsuranceInventoryExitFailure(l.ctx, l.svcCtx, item.Id, placeErr)
			return nil, placeErr
		}
		if resp == nil || resp.Data == nil || resp.Data.OrderId <= 0 {
			executionErr := errInvalidInsuranceInventoryExit
			if resp != nil && resp.Base != nil && resp.Base.Msg != "" {
				executionErr = fmt.Errorf("insurance inventory exit order rejected: %s", resp.Base.Msg)
			}
			saveInsuranceInventoryExitFailure(l.ctx, l.svcCtx, item.Id, executionErr)
			return insuranceInventoryExitError(l.ctx, executionErr), nil
		}
		order, findErr = l.svcCtx.OptionOrderModel.FindOne(l.ctx, resp.Data.OrderId)
		if findErr != nil {
			return nil, findErr
		}
	}
	if order.TenantId != item.TenantId || order.UserId != item.InsuranceUserId ||
		order.AccountId != item.InsuranceAccountId || order.ContractId != item.ContractId ||
		order.Source != int64(option.OrderSource_ORDER_SOURCE_ADMIN) ||
		order.Side != int64(common.Side_SIDE_BUY) ||
		order.PositionEffect != int64(option.PositionEffect_POSITION_EFFECT_CLOSE) ||
		order.OrderType != int64(option.OrderType_ORDER_TYPE_IOC) ||
		order.ReduceOnly != int64(common.YesNo_YES_NO_YES) ||
		order.ClientOrderId != clientOrderID || !order.Qty.Equal(item.Quantity) ||
		!order.Price.Equal(item.LimitPrice) {
		executionErr := errors.New("insurance inventory exit order evidence mismatch")
		saveInsuranceInventoryExitFailure(l.ctx, l.svcCtx, item.Id, executionErr)
		return insuranceInventoryExitError(l.ctx, executionErr), nil
	}
	now := time.Now().Unix()
	var submitted *models.TOptionInsuranceInventoryExit
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exitModel := models.NewTOptionInsuranceInventoryExitModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		current, lockErr := exitModel.FindOneForUpdate(ctx, item.Id)
		if lockErr != nil {
			return lockErr
		}
		if current.Status == int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_SUBMITTED) {
			if current.OrderId != order.Id {
				return errors.New("insurance inventory exit is linked to another order")
			}
			submitted = current
			return nil
		}
		if current.Status != int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_APPROVED) {
			return errInvalidInsuranceInventoryExit
		}
		current.Status = int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_SUBMITTED)
		current.ActiveKey = fmt.Sprintf("REQUEST:%s", current.RequestNo)
		current.OrderId = order.Id
		current.SubmittedBy = operatorID
		current.SubmittedAt = now
		current.LastErrorMsg = ""
		current.UpdateTimes = now
		if updateErr := exitModel.Update(ctx, current); updateErr != nil {
			return updateErr
		}
		if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: current.TenantId, UserId: current.InsuranceUserId,
			ContractId: current.ContractId, OrderId: order.Id,
			EventType: "INSURANCE_EXIT_SUBMITTED", Reason: "APPROVED_REDUCE_ONLY_IOC",
			Detail:     fmt.Sprintf("request_no=%s order_id=%d client_order_id=%s", current.RequestNo, order.Id, clientOrderID),
			OperatorId: operatorID, CreateTimes: now,
		}); insertErr != nil {
			return insertErr
		}
		submitted = current
		return nil
	})
	if errors.Is(err, errInvalidInsuranceInventoryExit) {
		return insuranceInventoryExitError(l.ctx, err), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetInsuranceInventoryExitResp{Base: helper.OkResp(), Data: helpers.ToInsuranceInventoryExitProto(submitted)}, nil
}

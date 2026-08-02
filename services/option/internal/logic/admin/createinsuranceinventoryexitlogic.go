package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/conv"
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

type CreateInsuranceInventoryExitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateInsuranceInventoryExitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateInsuranceInventoryExitLogic {
	return &CreateInsuranceInventoryExitLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreateInsuranceInventoryExitLogic) CreateInsuranceInventoryExit(
	in *option.CreateInsuranceInventoryExitReq,
) (*option.GetInsuranceInventoryExitResp, error) {
	limits, limitErr := insuranceInventoryExitRuntimeLimits(l.svcCtx)
	if limitErr != nil {
		return insuranceInventoryExitError(l.ctx, limitErr), nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason, evidenceRef := strings.TrimSpace(in.Reason), strings.TrimSpace(in.EvidenceRef)
	quantity, quantityErr := conv.ParseDecimalField(in.Quantity)
	limitPrice, priceErr := conv.ParseDecimalField(in.LimitPrice)
	if in.TenantId <= 0 || in.PositionId <= 0 || operatorID <= 0 ||
		quantityErr != nil || priceErr != nil || reason == "" || evidenceRef == "" ||
		len(reason) > 500 || len(evidenceRef) > 500 {
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
	positionSnapshot, err := l.svcCtx.OptionPositionModel.FindOne(l.ctx, in.PositionId)
	if err != nil {
		return nil, err
	}
	if positionSnapshot.TenantId != in.TenantId {
		return insuranceInventoryExitError(l.ctx, errInvalidInsuranceInventoryExit), nil
	}
	requestNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OIX", "")
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	var created *models.TOptionInsuranceInventoryExit
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		exitModel := models.NewTOptionInsuranceInventoryExitModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		contract, findErr := contractModel.FindOneForUpdate(ctx, positionSnapshot.ContractId)
		if findErr != nil {
			return findErr
		}
		position, findErr := positionModel.FindOneForUpdate(ctx, in.PositionId)
		if findErr != nil {
			return findErr
		}
		if position.TenantId != in.TenantId || position.ContractId != contract.Id {
			return errInvalidInsuranceInventoryExit
		}
		market, findErr := marketModel.FindOneByTenantIdContractId(ctx, in.TenantId, contract.Id)
		if findErr != nil {
			return findErr
		}
		if validateErr := validateInsuranceInventoryExit(position, contract, market, quantity, limitPrice, now); validateErr != nil {
			return validateErr
		}
		if validateErr := validateInsuranceInventoryExitRuntimeLimits(
			contract, market, quantity, limitPrice, limits,
		); validateErr != nil {
			return validateErr
		}
		reserved, reserveErr := exitModel.SumReservedQuantity(
			ctx, in.TenantId, contract.Id, insuranceInventoryExitUTCDayStart(time.Unix(now, 0)),
		)
		if reserveErr != nil {
			return reserveErr
		}
		if reserved.Add(quantity).GreaterThan(limits.maxDailyQuantity) {
			return fmt.Errorf("%w: daily quantity limit exceeded", errInsuranceInventoryExitLimits)
		}
		if _, findErr = exitModel.FindOpenByPositionForUpdate(ctx, in.TenantId, position.Id); findErr == nil {
			return errOpenInsuranceInventoryExit
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		created = &models.TOptionInsuranceInventoryExit{
			TenantId: in.TenantId, RequestNo: requestNo,
			ActiveKey:  fmt.Sprintf("POSITION:%d", position.Id),
			PositionId: position.Id, ContractId: contract.Id,
			InsuranceUserId: contract.InsuranceUserId, InsuranceAccountId: contract.InsuranceAccountId,
			Quantity: quantity, LimitPrice: limitPrice,
			Status: int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_PENDING_REVIEW),
			Reason: reason, EvidenceRef: evidenceRef, RequestedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := exitModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}
		_, insertErr = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: in.TenantId, UserId: contract.InsuranceUserId,
			ContractId: contract.Id, EventType: "INSURANCE_EXIT_CREATED",
			Reason: "GOVERNED_INVENTORY_REDUCTION",
			Detail: fmt.Sprintf("request_no=%s position_id=%d quantity=%s limit_price=%s evidence=%s",
				requestNo, position.Id, quantity, limitPrice, evidenceRef),
			OperatorId: operatorID, CreateTimes: now,
		})
		return insertErr
	})
	if err != nil && strings.Contains(err.Error(), "uk_option_insurance_exit_active") {
		err = errOpenInsuranceInventoryExit
	}
	if errors.Is(err, errInvalidInsuranceInventoryExit) || errors.Is(err, errOpenInsuranceInventoryExit) ||
		errors.Is(err, errInsuranceInventoryExitLimits) {
		return insuranceInventoryExitError(l.ctx, err), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetInsuranceInventoryExitResp{Base: helper.OkResp(), Data: helpers.ToInsuranceInventoryExitProto(created)}, nil
}

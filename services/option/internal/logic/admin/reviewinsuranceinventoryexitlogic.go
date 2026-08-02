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

type ReviewInsuranceInventoryExitLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewInsuranceInventoryExitLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewInsuranceInventoryExitLogic {
	return &ReviewInsuranceInventoryExitLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ReviewInsuranceInventoryExitLogic) ReviewInsuranceInventoryExit(
	in *option.ReviewInsuranceInventoryExitReq,
) (*option.GetInsuranceInventoryExitResp, error) {
	limits, limitErr := insuranceInventoryExitRuntimeLimits(l.svcCtx)
	if limitErr != nil {
		return insuranceInventoryExitError(l.ctx, limitErr), nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.ExitId <= 0 || operatorID <= 0 || reason == "" || len(reason) > 500 {
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
	now := time.Now().Unix()
	var reviewed *models.TOptionInsuranceInventoryExit
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		exitModel := models.NewTOptionInsuranceInventoryExitModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		item, findErr := exitModel.FindOneForUpdate(ctx, in.ExitId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId ||
			item.Status != int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_PENDING_REVIEW) ||
			item.RequestedBy == operatorID {
			return errInvalidInsuranceInventoryExit
		}
		if in.Approve {
			position, validationErr := positionModel.FindOneForUpdate(ctx, item.PositionId)
			if validationErr != nil {
				return validationErr
			}
			contract, validationErr := contractModel.FindOne(ctx, item.ContractId)
			if validationErr != nil {
				return validationErr
			}
			market, validationErr := marketModel.FindOneByTenantIdContractId(ctx, item.TenantId, item.ContractId)
			if validationErr != nil {
				return validationErr
			}
			if validationErr = validateInsuranceInventoryExit(position, contract, market, item.Quantity, item.LimitPrice, now); validationErr != nil {
				return validationErr
			}
			if validationErr = validateInsuranceInventoryExitRuntimeLimits(
				contract, market, item.Quantity, item.LimitPrice, limits,
			); validationErr != nil {
				return validationErr
			}
			reserved, reserveErr := exitModel.SumReservedQuantity(
				ctx, item.TenantId, item.ContractId, insuranceInventoryExitUTCDayStart(time.Unix(now, 0)),
			)
			if reserveErr != nil {
				return reserveErr
			}
			if reserved.GreaterThan(limits.maxDailyQuantity) {
				return fmt.Errorf("%w: daily quantity limit exceeded", errInsuranceInventoryExitLimits)
			}
			item.Status = int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_APPROVED)
		} else {
			item.Status = int64(option.InsuranceInventoryExitStatus_INSURANCE_INVENTORY_EXIT_STATUS_REJECTED)
			item.ActiveKey = fmt.Sprintf("REQUEST:%s", item.RequestNo)
		}
		item.ReviewedBy = operatorID
		item.ReviewReason = reason
		item.ReviewedAt = now
		item.UpdateTimes = now
		if updateErr := exitModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		eventType := "INSURANCE_EXIT_REJECTED"
		if in.Approve {
			eventType = "INSURANCE_EXIT_APPROVED"
		}
		if _, insertErr := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: item.TenantId, UserId: item.InsuranceUserId,
			ContractId: item.ContractId, EventType: eventType,
			Reason:     "INDEPENDENT_REVIEW",
			Detail:     fmt.Sprintf("request_no=%s approve=%t reason=%s", item.RequestNo, in.Approve, reason),
			OperatorId: operatorID, CreateTimes: now,
		}); insertErr != nil {
			return insertErr
		}
		reviewed = item
		return nil
	})
	if errors.Is(err, errInvalidInsuranceInventoryExit) || errors.Is(err, errInsuranceInventoryExitLimits) {
		return insuranceInventoryExitError(l.ctx, err), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetInsuranceInventoryExitResp{Base: helper.OkResp(), Data: helpers.ToInsuranceInventoryExitProto(reviewed)}, nil
}

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
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReviewCorporateActionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewCorporateActionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewCorporateActionLogic {
	return &ReviewCorporateActionLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// 独立管理员复核公司行动；不支持的事件只能进入人工处理
func (l *ReviewCorporateActionLogic) ReviewCorporateAction(
	in *option.ReviewCorporateActionReq,
) (*option.GetCorporateActionResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return corporateActionError(l.ctx, i18n.PermissionDenied), nil
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.ActionId <= 0 || operatorID <= 0 || reason == "" || len(reason) > 500 {
		return corporateActionError(l.ctx, i18n.ParamError), nil
	}

	now := time.Now().Unix()
	var reviewed *models.TOptionCorporateAction
	var mappings []*models.TOptionCorporateActionContract
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		actionModel := models.NewTOptionCorporateActionModel(conn, l.svcCtx.Config.CacheRedis)
		mappingModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)

		action, findErr := actionModel.FindOneForUpdate(ctx, in.ActionId)
		if findErr != nil {
			return findErr
		}
		if action.TenantId != in.TenantId ||
			action.Status != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_DRAFT) ||
			action.CreatedBy == operatorID {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		mappings, findErr = mappingModel.FindByAction(ctx, action.TenantId, action.Id)
		if findErr != nil {
			return findErr
		}
		if len(mappings) == 0 {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}

		manualOnly := false
		for _, sourceMapping := range mappings {
			mapping, lockErr := mappingModel.FindOneForUpdate(ctx, sourceMapping.Id)
			if lockErr != nil {
				return lockErr
			}
			halt, lockErr := haltModel.FindOneForUpdate(ctx, mapping.HaltId)
			if lockErr != nil {
				return lockErr
			}
			if halt.TenantId != action.TenantId || halt.ContractId != mapping.SourceContractId ||
				halt.Source != int64(option.TradingHaltSource_TRADING_HALT_SOURCE_CORPORATE_ACTION) ||
				halt.Status != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE) ||
				halt.LastErrorMsg != "" {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			source, lockErr := contractModel.FindOneForUpdate(ctx, mapping.SourceContractId)
			if lockErr != nil {
				return lockErr
			}
			if source.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			activeOrders, _, lockErr := l.svcCtx.OptionOrderModel.FindPage(ctx, models.OptionOrderPageFilter{
				TenantId: action.TenantId, ContractId: source.Id,
				Statuses: []int64{
					int64(option.OrderStatus_ORDER_STATUS_FUNDING),
					int64(option.OrderStatus_ORDER_STATUS_PENDING),
					int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
					int64(option.OrderStatus_ORDER_STATUS_CANCELING),
				},
			}, 0, 1)
			if lockErr != nil {
				return lockErr
			}
			if len(activeOrders) > 0 {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			if mapping.ExecutionMode == int64(option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_MANUAL_ONLY) {
				manualOnly = true
			} else if validateErr := validateCorporateActionSuccessor(
				ctx, contractModel, positionModel, source, &option.CorporateActionContractInput{
					SourceContractId: source.Id, SuccessorContractId: mapping.SuccessorContractId,
					ExecutionMode: option.CorporateActionExecutionMode(mapping.ExecutionMode),
				},
			); validateErr != nil {
				return validateErr
			}
			mappings[sourceMappingIndex(mappings, mapping.Id)] = mapping
		}

		action.ReviewedBy = operatorID
		action.ReviewReason = reason
		action.ReviewedAt = now
		action.UpdateTimes = now
		switch {
		case !in.Approve:
			action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_REJECTED)
		case manualOnly:
			action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_MANUAL_REVIEW)
		default:
			action.Status = int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_APPROVED)
		}
		for _, mapping := range mappings {
			if !in.Approve || manualOnly {
				mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_MANUAL_REVIEW)
			} else {
				mapping.Status = int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_READY)
			}
			mapping.UpdateTimes = now
			if updateErr := mappingModel.Update(ctx, mapping); updateErr != nil {
				return updateErr
			}
		}
		if updateErr := actionModel.Update(ctx, action); updateErr != nil {
			return updateErr
		}
		reviewed = action
		return nil
	})
	if errors.Is(err, models.ErrNotFound) {
		return corporateActionError(l.ctx, i18n.ParamError), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetCorporateActionResp{
		Base: helper.OkResp(), Data: helpers.ToCorporateActionProto(reviewed, mappings),
	}, nil
}

func sourceMappingIndex(items []*models.TOptionCorporateActionContract, id int64) int {
	for index, item := range items {
		if item.Id == id {
			return index
		}
	}
	return 0
}

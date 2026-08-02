package applogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CancelComboOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelComboOrderLogic {
	return &CancelComboOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 原子撤销未完成组合订单并释放各腿冻结
func (l *CancelComboOrderLogic) CancelComboOrder(in *option.CancelComboOrderReq) (*option.UserCommonResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := findComboOrderByNoOrID(
		l.ctx, l.svcCtx, tenantID, in.ComboOrderId, in.ComboNo,
	)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.UserCommonResp{
				Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx)),
			}, nil
		}
		return nil, err
	}
	if item.UserId != userID || item.AccountId != in.AccountId {
		return &option.UserCommonResp{Base: helper.ErrResp(
			i18n.NoPermissionOperateOrder,
			i18n.Translate(i18n.NoPermissionOperateOrder, l.ctx),
		)}, nil
	}
	var changedChildren []*models.TOptionOrder
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		current, lockErr := comboModel.FindOneForUpdate(ctx, item.Id)
		if lockErr != nil {
			return lockErr
		}
		switch option.ComboOrderStatus(current.Status) {
		case option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED:
		default:
			return i18n.StatusError(l.ctx, i18n.CurrentStatusCannotCancel)
		}
		children, lockErr := orderModel.FindComboChildrenForUpdate(
			ctx, current.TenantId, current.Id,
		)
		if lockErr != nil {
			return lockErr
		}
		if len(children) < minComboLegs || len(children) > maxComboLegs {
			return errors.New("combo child-order cardinality invariant violated")
		}
		now := time.Now().Unix()
		requiresRelease := false
		for _, child := range children {
			cancelBeforeFreeze := false
			if child.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
				freeze, findErr := instructionModel.FindOneByTenantIdInstructionNo(
					ctx, child.TenantId, child.OrderNo+"-FREEZE",
				)
				if findErr != nil {
					return findErr
				}
				freeze, findErr = instructionModel.FindOneForUpdate(ctx, freeze.Id)
				if findErr != nil {
					return findErr
				}
				if freeze.Status == int64(
					option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING,
				) {
					freeze.Status = int64(
						option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED,
					)
					freeze.UpdateTimes = now
					if findErr = instructionModel.Update(ctx, freeze); findErr != nil {
						return findErr
					}
					cancelBeforeFreeze = true
					child.MarginAmount = decimal.Zero
				}
			}
			child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
			if child.MarginAmount.IsPositive() && !cancelBeforeFreeze {
				requiresRelease = true
				child.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
				if _, insertErr := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
					TenantId:      child.TenantId,
					InstructionNo: child.OrderNo + "-COMBO-CANCEL-RELEASE",
					BizNo:         child.OrderNo, OrderId: child.Id, UserId: child.UserId,
					AccountId: child.AccountId,
					Action: int64(
						option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN,
					),
					TargetBizNo: child.OrderNo, Coin: OptionOrderMarginCoin(child),
					Amount: child.MarginAmount, StepNo: 2,
					Status: int64(
						option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING,
					),
					ReconciliationStatus: int64(
						option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING,
					),
					CreateTimes: now, UpdateTimes: now,
				}); insertErr != nil {
					return insertErr
				}
			}
			child.CancelReason = "COMBO_USER_CANCEL"
			child.CancelTime = now
			child.UpdateTimes = now
			if updateErr := orderModel.Update(ctx, child); updateErr != nil {
				return updateErr
			}
			childCopy := *child
			changedChildren = append(changedChildren, &childCopy)
		}
		if updateErr := transitionComboToCancellation(
			ctx, comboModel, current, requiresRelease, "USER_CANCEL", now,
		); updateErr != nil {
			return updateErr
		}
		*item = *current
		return nil
	})
	if err != nil {
		if i18n.IsStatusError(err, i18n.CurrentStatusCannotCancel) {
			return &option.UserCommonResp{Base: helper.ErrResp(
				i18n.CurrentStatusCannotCancel,
				i18n.Translate(i18n.CurrentStatusCannotCancel, l.ctx),
			)}, nil
		}
		return nil, err
	}
	for _, child := range changedChildren {
		publishOptionOrderChanged(l.ctx, l.svcCtx, child)
	}
	return &option.UserCommonResp{Base: helper.OkResp()}, nil
}

package applogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CancelOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelOrderLogic {
	return &CancelOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 撤销期权委托订单
func (l *CancelOrderLogic) CancelOrder(in *option.CancelOrderReq) (*option.UserCommonResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := helpers.FindOrderByNoOrID(l.ctx, l.svcCtx, tenantId, in.OrderId, in.OrderNo)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.UserCommonResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.UserId != userId || item.AccountId != in.AccountId {
		return &option.UserCommonResp{Base: helper.ErrResp(i18n.NoPermissionOperateOrder, i18n.Translate(i18n.NoPermissionOperateOrder, l.ctx))}, nil
	}
	if item.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) && item.Status != int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
		if item.Status != int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
			return &option.UserCommonResp{Base: helper.ErrResp(i18n.CurrentStatusCannotCancel, i18n.Translate(i18n.CurrentStatusCannotCancel, l.ctx))}, nil
		}
	}

	wasFunding := item.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING)
	now := time.Now().Unix()
	item.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
	if item.MarginAmount.IsPositive() {
		item.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
	}
	item.CancelReason = "USER_CANCEL"
	item.CancelTime = now
	item.UpdateTimes = now
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		cancelBeforeFreeze := false
		if wasFunding {
			freezeInstruction, err := instructionModel.FindOneByTenantIdInstructionNo(ctx, item.TenantId, item.OrderNo+"-FREEZE")
			if err != nil {
				return err
			}
			locked, err := instructionModel.FindOneForUpdate(ctx, freezeInstruction.Id)
			if err != nil {
				return err
			}
			if locked.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) {
				locked.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED)
				locked.UpdateTimes = now
				if err := instructionModel.Update(ctx, locked); err != nil {
					return err
				}
				cancelBeforeFreeze = true
				item.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
				item.MarginAmount = decimal.Zero
			}
		}
		if err := releaseClosePositionFrozenQty(ctx, positionModel, item, item.UnfilledQty, now); err != nil {
			return err
		}
		if item.MarginAmount.IsPositive() && !cancelBeforeFreeze {
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: item.TenantId, InstructionNo: item.OrderNo + "-CANCEL-RELEASE",
				BizNo: item.OrderNo, OrderId: item.Id, UserId: item.UserId, AccountId: item.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: item.OrderNo, Coin: OptionOrderMarginCoin(item), Amount: item.MarginAmount,
				StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		return orderModel.Update(ctx, item)
	})
	if err != nil {
		return nil, err
	}
	publishOptionOrderChanged(l.ctx, l.svcCtx, item)

	return &option.UserCommonResp{Base: helper.OkResp()}, nil
}

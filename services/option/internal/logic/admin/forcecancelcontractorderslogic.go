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

	"github.com/shopspring/decimal"
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
			return nil, err
		}
		for _, order := range orders {
			cursor = order.Id
			if err := l.forceCancelOne(order.Id, reason); err != nil {
				return nil, err
			}
		}
		if len(orders) < 100 {
			break
		}
	}
	return &option.CommonResp{Base: helper.OkResp()}, nil
}

func (l *ForceCancelContractOrdersLogic) forceCancelOne(orderId int64, reason string) error {
	var canceled *models.TOptionOrder
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		order, err := orderModel.FindOneForUpdate(ctx, orderId)
		if err != nil {
			return err
		}
		switch option.OrderStatus(order.Status) {
		case option.OrderStatus_ORDER_STATUS_FUNDING,
			option.OrderStatus_ORDER_STATUS_PENDING,
			option.OrderStatus_ORDER_STATUS_PART_FILLED:
		default:
			return nil
		}
		now := time.Now().Unix()
		cancelBeforeFreeze := false
		if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) {
			freeze, err := instructionModel.FindOneByTenantIdInstructionNo(ctx, order.TenantId, order.OrderNo+"-FREEZE")
			if err != nil {
				return err
			}
			freeze, err = instructionModel.FindOneForUpdate(ctx, freeze.Id)
			if err != nil {
				return err
			}
			if freeze.Status == int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) {
				freeze.Status = int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_CANCELED)
				freeze.UpdateTimes = now
				if err := instructionModel.Update(ctx, freeze); err != nil {
					return err
				}
				cancelBeforeFreeze = true
				order.MarginAmount = decimal.Zero
			}
		}
		if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
			return err
		}
		order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELED)
		if order.MarginAmount.IsPositive() && !cancelBeforeFreeze {
			order.Status = int64(option.OrderStatus_ORDER_STATUS_CANCELING)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: order.TenantId, InstructionNo: order.OrderNo + "-ADMIN-RELEASE",
				BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: order.OrderNo, Coin: order.FeeCoin, Amount: order.MarginAmount,
				StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
		}
		order.CancelReason = reason
		order.CancelTime = now
		order.UpdateTimes = now
		if err := orderModel.Update(ctx, order); err != nil {
			return err
		}
		canceled = order
		return nil
	})
	if err == nil && canceled != nil {
		applogic.PublishOptionOrderChanged(l.ctx, l.svcCtx, canceled)
	}
	return err
}

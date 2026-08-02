package applogic

import (
	"context"
	"errors"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// CompleteComboFunding enforces an all-legs funding barrier. No shadow child
// is admitted to the strategy matcher until every leg freeze is successful.
func CompleteComboFunding(
	ctx context.Context, svcCtx *svc.ServiceContext, comboOrderID int64,
) error {
	if comboOrderID <= 0 {
		return nil
	}
	var parent *models.TOptionComboOrder
	var changedChildren []*models.TOptionOrder
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, svcCtx.Config.CacheRedis)
		userControlModel := models.NewTOptionUserTradingControlModel(conn, svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, svcCtx.Config.CacheRedis)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, svcCtx.Config.CacheRedis)
		calendarSessionModel := models.NewTOptionTradingCalendarSessionModel(conn, svcCtx.Config.CacheRedis)
		calendarExceptionModel := models.NewTOptionTradingCalendarExceptionModel(conn, svcCtx.Config.CacheRedis)

		current, err := comboModel.FindOneForUpdate(txCtx, comboOrderID)
		if err != nil {
			return err
		}
		parent = current
		if current.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING) {
			return nil
		}
		children, err := orderModel.FindComboChildrenForUpdate(
			txCtx, current.TenantId, current.Id,
		)
		if err != nil {
			return err
		}
		if len(children) < minComboLegs || len(children) > maxComboLegs {
			return errors.New("combo child-order cardinality invariant violated")
		}
		for _, child := range children {
			freeze, err := instructionModel.FindOneByTenantIdInstructionNo(
				txCtx, child.TenantId, child.OrderNo+"-FREEZE",
			)
			if err != nil {
				return err
			}
			freeze, err = instructionModel.FindOneForUpdate(txCtx, freeze.Id)
			if err != nil {
				return err
			}
			if freeze.Status != int64(
				option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS,
			) {
				return nil
			}
		}

		now := time.Now().Unix()
		admit := true
		admissionReason := "COMBO_NOT_TRADABLE_AFTER_FUNDING"
		rejectAdmission := func(reason string) {
			if admit && reason != "" {
				admissionReason = reason
			}
			admit = false
		}
		userControl, err := userControlModel.EnsureForUpdate(
			txCtx, current.TenantId, current.UserId, now,
		)
		if err != nil {
			return err
		}
		if userControl.KillSwitch == int64(common.YesNo_YES_NO_YES) {
			rejectAdmission("COMBO_USER_KILL_SWITCH_AFTER_FUNDING")
		}
		for _, child := range children {
			contract, err := contractModel.FindOneForUpdate(txCtx, child.ContractId)
			if err != nil {
				return err
			}
			if contract.TenantId != child.TenantId ||
				contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
				contract.IsDeleted == int64(common.YesNo_YES_NO_YES) ||
				now < contract.ListTime ||
				(contract.LastTradeTime <= 0 || now >= contract.LastTradeTime) {
				rejectAdmission("COMBO_CONTRACT_NOT_TRADABLE_AFTER_FUNDING")
				continue
			}
			decision, calendarErr := logichelpers.IsContractTradingOpenWithModels(
				txCtx, haltModel, calendarModel, calendarSessionModel,
				calendarExceptionModel, contract, now,
			)
			if calendarErr != nil || decision == nil || !decision.Open {
				rejectAdmission("COMBO_CALENDAR_CLOSED_AFTER_FUNDING")
				continue
			}
			market, marketErr := marketModel.FindOneByTenantIdContractIdForUpdate(
				txCtx, child.TenantId, child.ContractId,
			)
			if marketErr != nil || !logichelpers.IsMarkFresh(market, now, 30) ||
				!market.MarkPrice.IsPositive() {
				if marketErr != nil && !errors.Is(marketErr, models.ErrNotFound) {
					return marketErr
				}
				rejectAdmission("COMBO_STALE_MARK_AFTER_FUNDING")
				continue
			}
			if _, _, withinBand := optionOrderPriceBand(
				child.Price, market.MarkPrice, contract.OrderPriceBandRatio,
			); !withinBand {
				rejectAdmission("COMBO_PRICE_BAND_AFTER_FUNDING")
				continue
			}
			if child.Side == int64(common.Side_SIDE_SELL) {
				if !logichelpers.IsUnderlyingFresh(market, now, 30) ||
					!market.UnderlyingPrice.IsPositive() {
					rejectAdmission("COMBO_STALE_UNDERLYING_AFTER_FUNDING")
					continue
				}
				requiredMargin := optionSellerMargin(
					contract, market.UnderlyingPrice, child.Price, child.UnfilledQty, false,
				)
				if !requiredMargin.IsPositive() || child.MarginAmount.LessThan(requiredMargin) {
					rejectAdmission("COMBO_SELL_MARGIN_INSUFFICIENT_AFTER_FUNDING")
				}
			}
		}
		for _, child := range children {
			if admit {
				child.Status = int64(option.OrderStatus_ORDER_STATUS_PENDING)
			} else {
				child.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRING)
				child.CancelReason = admissionReason
				child.CancelTime = now
				if _, err := instructionModel.Insert(txCtx, &models.TOptionAssetInstruction{
					TenantId:      child.TenantId,
					InstructionNo: child.OrderNo + "-COMBO-FUNDING-RELEASE",
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
				}); err != nil {
					return err
				}
			}
			child.UpdateTimes = now
			if err := orderModel.Update(txCtx, child); err != nil {
				return err
			}
			childCopy := *child
			changedChildren = append(changedChildren, &childCopy)
		}
		if admit {
			current.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE)
		} else {
			current.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELING)
			current.CancelReason = admissionReason
			current.CancelTime = now
		}
		current.UpdateTimes = now
		return comboModel.Update(txCtx, current)
	})
	if err != nil {
		return err
	}
	for _, child := range changedChildren {
		publishOptionOrderChanged(ctx, svcCtx, child)
	}
	if parent != nil &&
		parent.Status == int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE) {
		return MatchFundedComboOrder(ctx, svcCtx, parent)
	}
	return nil
}

// FinalizeComboCancellation promotes the parent only after every shadow child
// has reached a terminal state following its release instruction.
func FinalizeComboCancellation(
	ctx context.Context, svcCtx *svc.ServiceContext, comboOrderID int64,
) error {
	if comboOrderID <= 0 {
		return nil
	}
	return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		parent, err := comboModel.FindOneForUpdate(txCtx, comboOrderID)
		if err != nil {
			return err
		}
		if parent.Status != int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELING) {
			return nil
		}
		children, err := orderModel.FindComboChildrenForUpdate(
			txCtx, parent.TenantId, parent.Id,
		)
		if err != nil {
			return err
		}
		allTerminal := len(children) >= minComboLegs && len(children) <= maxComboLegs
		for _, child := range children {
			switch option.OrderStatus(child.Status) {
			case option.OrderStatus_ORDER_STATUS_CANCELED,
				option.OrderStatus_ORDER_STATUS_EXPIRED,
				option.OrderStatus_ORDER_STATUS_FILLED,
				option.OrderStatus_ORDER_STATUS_REJECTED:
			default:
				allTerminal = false
			}
		}
		if allTerminal {
			parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED)
		} else {
			return nil
		}
		parent.UpdateTimes = time.Now().Unix()
		return comboModel.Update(txCtx, parent)
	})
}

// MarkComboManualReview surfaces an unrecoverable funding/release instruction
// at the parent level without mutating any economic field or hiding the
// affected shadow child.
func MarkComboManualReview(
	ctx context.Context, svcCtx *svc.ServiceContext, comboOrderID int64,
) error {
	if comboOrderID <= 0 {
		return nil
	}
	return svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		comboModel := models.NewTOptionComboOrderModel(conn, svcCtx.Config.CacheRedis)
		parent, err := comboModel.FindOneForUpdate(txCtx, comboOrderID)
		if err != nil {
			return err
		}
		switch option.ComboOrderStatus(parent.Status) {
		case option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_PART_FILLED,
			option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELING:
			parent.Status = int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_MANUAL_REVIEW)
			parent.UpdateTimes = time.Now().Unix()
			return comboModel.Update(txCtx, parent)
		default:
			return nil
		}
	})
}

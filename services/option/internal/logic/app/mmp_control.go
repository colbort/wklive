package applogic

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	controlEventMMPConfigured     = "MMP_CONFIGURED"
	controlEventMMPAutoReset      = "MMP_AUTO_RESET"
	controlEventMMPManualReset    = "MMP_MANUAL_RESET"
	controlEventMMPTriggered      = "MMP_TRIGGERED"
	controlEventMMPOrderCanceled  = "MMP_ORDER_CANCELED"
	controlReasonMMPNotConfigured = "MMP_NOT_CONFIGURED"
	controlReasonMMPDisabled      = "MMP_DISABLED"
	controlReasonMMPTriggered     = "MMP_TRIGGERED"
	controlReasonMMPInvalidOrder  = "MMP_INVALID_ORDER"
)

var mmpGroupPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,32}$`)

func NormalizeMMPGroup(value string) (string, bool) {
	// The table uses the service's default case-insensitive utf8mb4 collation.
	// Canonical lowercase prevents callers from treating "Desk-A" and
	// "desk-a" as separate groups while the database treats them as equal.
	value = strings.ToLower(strings.TrimSpace(value))
	return value, mmpGroupPattern.MatchString(value)
}

func ensureMMPAdmission(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	conn sqlx.SqlConn,
	order *models.TOptionOrder,
	now int64,
) (*orderControlRejection, error) {
	if order.Mmp != int64(common.YesNo_YES_NO_YES) {
		if order.MmpGroup != "" {
			return &orderControlRejection{
				reason: controlReasonMMPInvalidOrder,
				detail: "mmp_group is only valid for an MMP order",
			}, nil
		}
		return nil, nil
	}
	if order.OrderType != int64(option.OrderType_ORDER_TYPE_POST_ONLY) {
		return &orderControlRejection{
			reason: controlReasonMMPInvalidOrder,
			detail: "MMP orders must be POST_ONLY",
		}, nil
	}
	groupCode, valid := NormalizeMMPGroup(order.MmpGroup)
	if !valid {
		return &orderControlRejection{
			reason: controlReasonMMPInvalidOrder,
			detail: "mmp_group must match [A-Za-z0-9_-]{1,32}",
		}, nil
	}
	order.MmpGroup = groupCode
	model := models.NewTOptionMmpConfigModel(conn, svcCtx.Config.CacheRedis)
	config, err := model.FindForUpdate(
		ctx, order.TenantId, order.UserId, order.ContractId, groupCode,
	)
	if errors.Is(err, models.ErrNotFound) {
		return &orderControlRejection{
			reason: controlReasonMMPNotConfigured,
			detail: "no MMP configuration exists for this user, contract and group",
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if config.Enabled != int64(common.YesNo_YES_NO_YES) ||
		config.Status == int64(option.MMPStatus_MMP_STATUS_DISABLED) {
		return &orderControlRejection{
			reason: controlReasonMMPDisabled,
			detail: fmt.Sprintf("config_id=%d", config.Id),
		}, nil
	}
	if config.Status == int64(option.MMPStatus_MMP_STATUS_TRIGGERED) {
		if config.CooldownUntil <= 0 || now < config.CooldownUntil {
			return &orderControlRejection{
				reason: controlReasonMMPTriggered,
				detail: fmt.Sprintf("triggered_at=%d cooldown_until=%d", config.TriggeredAt, config.CooldownUntil),
			}, nil
		}
		if config.LastErrorMsg != "" {
			return &orderControlRejection{
				reason: controlReasonMMPTriggered,
				detail: fmt.Sprintf(
					"cooldown expired but cancellation error requires manual reset: %s",
					config.LastErrorMsg,
				),
			}, nil
		}
		orderModel := models.NewTOptionOrderModel(conn, svcCtx.Config.CacheRedis)
		residual, residualErr := orderModel.FindFirstActiveMMPOrderForUpdate(
			ctx, config.TenantId, config.UserId, config.ContractId, config.GroupCode,
		)
		if residualErr != nil && !errors.Is(residualErr, models.ErrNotFound) {
			return nil, residualErr
		}
		if residual != nil {
			return &orderControlRejection{
				reason: controlReasonMMPTriggered,
				detail: fmt.Sprintf(
					"cooldown expired but residual order_id=%d requires cancellation",
					residual.Id,
				),
			}, nil
		}
		resetMMPWindow(config, now)
		config.Status = int64(option.MMPStatus_MMP_STATUS_ACTIVE)
		config.TriggerReason = ""
		config.LastErrorMsg = ""
		config.UpdateTimes = now
		if err := model.Update(ctx, config); err != nil {
			return nil, err
		}
		if err := insertTradingControlEvent(ctx, svcCtx, conn, &models.TOptionTradingControlEvent{
			TenantId: config.TenantId, UserId: config.UserId, ContractId: config.ContractId,
			EventType: controlEventMMPAutoReset, Reason: "COOLDOWN_EXPIRED",
			Detail:     fmt.Sprintf("group=%s config_id=%d", config.GroupCode, config.Id),
			OperatorId: config.UserId, CreateTimes: now,
		}); err != nil {
			return nil, err
		}
	}
	if config.Status != int64(option.MMPStatus_MMP_STATUS_ACTIVE) {
		return &orderControlRejection{
			reason: controlReasonMMPDisabled,
			detail: fmt.Sprintf("status=%d", config.Status),
		}, nil
	}
	return nil, nil
}

func resetMMPWindow(config *models.TOptionMmpConfig, now int64) {
	config.WindowStart = now
	config.AccumulatedQty = decimal.Zero
	config.TradeCount = 0
	config.AccumulatedLoss = decimal.Zero
	config.TriggeredAt = 0
	config.CooldownUntil = 0
}

func applyMMPFill(
	config *models.TOptionMmpConfig,
	makerSide int64,
	tradePrice, tradeQty, markPrice, multiplier, makerFee decimal.Decimal,
	now int64,
) (bool, string) {
	if config == nil || config.Enabled != int64(common.YesNo_YES_NO_YES) ||
		config.Status != int64(option.MMPStatus_MMP_STATUS_ACTIVE) {
		return false, ""
	}
	if config.WindowStart <= 0 || now-config.WindowStart >= config.WindowSeconds {
		resetMMPWindow(config, now)
	}
	config.AccumulatedQty = config.AccumulatedQty.Add(tradeQty).Round(16)
	config.TradeCount++
	adverse := decimal.Zero
	if makerSide == int64(common.Side_SIDE_BUY) {
		adverse = decimal.Max(tradePrice.Sub(markPrice), decimal.Zero)
	} else {
		adverse = decimal.Max(markPrice.Sub(tradePrice), decimal.Zero)
	}
	loss := adverse.Mul(tradeQty).Mul(multiplier).Add(makerFee).Round(16)
	config.AccumulatedLoss = config.AccumulatedLoss.Add(loss).Round(16)
	reasons := make([]string, 0, 3)
	if config.QtyThreshold.IsPositive() &&
		config.AccumulatedQty.GreaterThanOrEqual(config.QtyThreshold) {
		reasons = append(reasons, "QTY_THRESHOLD")
	}
	if config.TradeCountThreshold > 0 && config.TradeCount >= config.TradeCountThreshold {
		reasons = append(reasons, "TRADE_COUNT_THRESHOLD")
	}
	if config.LossThreshold.IsPositive() &&
		config.AccumulatedLoss.GreaterThanOrEqual(config.LossThreshold) {
		reasons = append(reasons, "LOSS_THRESHOLD")
	}
	config.UpdateTimes = now
	if len(reasons) == 0 {
		return false, ""
	}
	reason := strings.Join(reasons, ",")
	config.Status = int64(option.MMPStatus_MMP_STATUS_TRIGGERED)
	config.TriggeredAt = now
	config.CooldownUntil = now + config.CooldownSeconds
	config.TriggerReason = reason
	config.LastErrorMsg = ""
	return true, reason
}

func MMPMakerFee(trade *models.TOptionTrade, makerSide int64) decimal.Decimal {
	if trade == nil {
		return decimal.Zero
	}
	if makerSide == int64(common.Side_SIDE_BUY) {
		return trade.BuyFee
	}
	return trade.SellFee
}

// CancelMMPGroupOrders prevents a triggered/disabled group from leaving live
// quotes behind. New group orders must already be blocked by the config state.
func CancelMMPGroupOrders(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	tenantID, userID, contractID int64,
	groupCode, reason string,
) (int64, error) {
	var canceled int64
	cursor := int64(0)
	for {
		orders, err := svcCtx.OptionOrderModel.FindActiveMMPOrders(
			ctx, tenantID, userID, contractID, groupCode, cursor, 100,
		)
		if err != nil {
			return canceled, err
		}
		for _, order := range orders {
			cursor = order.Id
			item, err := CancelOrderByControl(ctx, svcCtx, order.Id, reason)
			if err != nil {
				return canceled, err
			}
			if item != nil {
				canceled++
			}
		}
		if len(orders) < 100 {
			return canceled, nil
		}
	}
}

func SetMMPConfigLastError(
	ctx context.Context, svcCtx *svc.ServiceContext,
	tenantID, userID, contractID int64, groupCode, message string,
) {
	runes := []rune(message)
	if len(runes) > 500 {
		message = string(runes[:500])
	}
	err := svcCtx.DB.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		model := models.NewTOptionMmpConfigModel(
			sqlx.NewSqlConnFromSession(session), svcCtx.Config.CacheRedis,
		)
		config, err := model.FindForUpdate(txCtx, tenantID, userID, contractID, groupCode)
		if err != nil {
			return err
		}
		config.LastErrorMsg = message
		config.UpdateTimes = time.Now().Unix()
		return model.Update(txCtx, config)
	})
	if err != nil {
		logx.WithContext(ctx).Errorf(
			"option mmp failed to persist last error tenantId=%d userId=%d contractId=%d group=%s err=%v",
			tenantID, userID, contractID, groupCode, err,
		)
	}
}

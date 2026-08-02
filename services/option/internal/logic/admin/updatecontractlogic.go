package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/observability"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContractLogic {
	return &UpdateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权合约
func (l *UpdateContractLogic) UpdateContract(in *option.UpdateContractReq) (*option.CommonResp, error) {
	item, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	original := *item
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	targetTenantId := item.TenantId
	if allowTenantUpdate {
		targetTenantId = in.TenantId
	}

	if targetTenantId != item.TenantId || (in.ContractCode != "" && in.ContractCode != item.ContractCode) {
		contractCode := item.ContractCode
		if in.ContractCode != "" {
			contractCode = in.ContractCode
		}
		dup, err := l.svcCtx.OptionContractModel.FindOneByTenantIdContractCode(l.ctx, targetTenantId, contractCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if dup != nil && dup.Id != item.Id {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractCodeAlreadyExists, i18n.Translate(i18n.ContractCodeAlreadyExists, l.ctx))}, nil
		}
		item.ContractCode = contractCode
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}
	if in.UnderlyingSymbol != "" {
		item.UnderlyingSymbol = in.UnderlyingSymbol
	}
	if in.UnderlyingCoin != "" {
		item.UnderlyingCoin = in.UnderlyingCoin
	}
	if in.SettleCoin != "" {
		item.SettleCoin = in.SettleCoin
	}
	if in.QuoteCoin != "" {
		item.QuoteCoin = in.QuoteCoin
	}
	if in.OptionType != 0 {
		item.OptionType = int64(in.OptionType)
	}
	if in.ExerciseStyle != 0 {
		item.ExerciseStyle = int64(in.ExerciseStyle)
	}
	if in.SettlementType != 0 {
		item.SettlementType = int64(in.SettlementType)
	}
	if in.StrikePrice != "" {
		value, err := conv.ParseDecimalField(in.StrikePrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.StrikePriceFormatError, i18n.Translate(i18n.StrikePriceFormatError, l.ctx))}, nil
		}
		item.StrikePrice = value
	}
	if in.ContractUnit != "" {
		value, err := conv.ParseDecimalField(in.ContractUnit)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractUnitFormatError, i18n.Translate(i18n.ContractUnitFormatError, l.ctx))}, nil
		}
		item.ContractUnit = value
	}
	if in.MinOrderQty != "" {
		value, err := conv.ParseDecimalField(in.MinOrderQty)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MinOrderQuantityFormatError, i18n.Translate(i18n.MinOrderQuantityFormatError, l.ctx))}, nil
		}
		item.MinOrderQty = value
	}
	if in.MaxOrderQty != "" {
		value, err := conv.ParseDecimalField(in.MaxOrderQty)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MaxOrderQuantityFormatError, i18n.Translate(i18n.MaxOrderQuantityFormatError, l.ctx))}, nil
		}
		item.MaxOrderQty = value
	}
	if in.PriceTick != "" {
		value, err := conv.ParseDecimalField(in.PriceTick)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.PriceTickFormatError, i18n.Translate(i18n.PriceTickFormatError, l.ctx))}, nil
		}
		item.PriceTick = value
	}
	if in.QtyStep != "" {
		value, err := conv.ParseDecimalField(in.QtyStep)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.QuantityStepFormatError, i18n.Translate(i18n.QuantityStepFormatError, l.ctx))}, nil
		}
		item.QtyStep = value
	}
	if in.Multiplier != "" {
		value, err := conv.ParseDecimalField(in.Multiplier)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MultiplierFormatError, i18n.Translate(i18n.MultiplierFormatError, l.ctx))}, nil
		}
		item.Multiplier = value
	}
	if in.ListTime != 0 {
		item.ListTime = in.ListTime
	}
	if in.LastTradeTime != 0 {
		item.LastTradeTime = in.LastTradeTime
	}
	if in.ExpireTime != 0 {
		item.ExpireTime = in.ExpireTime
	}
	if in.DeliverTime != 0 {
		item.DeliverTime = in.DeliverTime
	}
	if in.ExerciseCutoffTime != 0 {
		item.ExerciseCutoffTime = in.ExerciseCutoffTime
	}
	if in.AutoExerciseThreshold != "" {
		value, err := conv.ParseDecimalField(in.AutoExerciseThreshold)
		if err != nil || value.IsNegative() {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.AutoExerciseThreshold = value
	}
	if in.MaxUserLongQty != "" {
		value, err := conv.ParseDecimalField(in.MaxUserLongQty)
		if err != nil || value.IsNegative() {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MaxUserLongQty = value
	}
	if in.MaxUserShortQty != "" {
		value, err := conv.ParseDecimalField(in.MaxUserShortQty)
		if err != nil || value.IsNegative() {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MaxUserShortQty = value
	}
	if in.MaxOpenInterest != "" {
		value, err := conv.ParseDecimalField(in.MaxOpenInterest)
		if err != nil || value.IsNegative() {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MaxOpenInterest = value
	}
	if in.OrderPriceBandRatio != "" {
		value, err := parseOptionalOptionRate(in.OrderPriceBandRatio)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.OrderPriceBandRatio = value
	}
	if in.CircuitBreakerRatio != "" {
		value, err := parseOptionalOptionRate(in.CircuitBreakerRatio)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.CircuitBreakerRatio = value
	}
	if in.GreeksMaxAgeSeconds != 0 {
		item.GreeksMaxAgeSeconds = in.GreeksMaxAgeSeconds
	}
	if in.SettlementPriceSource != "" {
		item.SettlementPriceSource = in.SettlementPriceSource
	}
	if in.SettlementPriceMethod != "" {
		item.SettlementPriceMethod = in.SettlementPriceMethod
	}
	if in.SettlementWindowSeconds != 0 {
		item.SettlementWindowSeconds = in.SettlementWindowSeconds
	}
	if in.SettlementMinSamples != 0 {
		item.SettlementMinSamples = in.SettlementMinSamples
	}
	if in.IsAutoExercise != 0 {
		item.IsAutoExercise = int64(in.IsAutoExercise)
	}
	if in.MakerFeeRate != "" {
		value, err := parseOptionalOptionRate(in.MakerFeeRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MakerFeeRate = value
	}
	if in.TakerFeeRate != "" {
		value, err := parseOptionalOptionRate(in.TakerFeeRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.TakerFeeRate = value
	}
	if in.ExerciseFeeRate != "" {
		value, err := parseOptionalOptionRate(in.ExerciseFeeRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.ExerciseFeeRate = value
	}
	if in.FeeUserId != 0 {
		item.FeeUserId = in.FeeUserId
	}
	if in.FeeAccountId != 0 {
		item.FeeAccountId = in.FeeAccountId
	}
	if in.SellerMarginMode != option.SellerMarginMode_SELLER_MARGIN_MODE_UNKNOWN {
		item.SellerMarginMode = int64(in.SellerMarginMode)
	}
	if in.InitialMarginRate != "" {
		value, err := parseOptionalOptionRate(in.InitialMarginRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.InitialMarginRate = value
	}
	if in.MaintenanceMarginRate != "" {
		value, err := parseOptionalOptionRate(in.MaintenanceMarginRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MaintenanceMarginRate = value
	}
	if in.MinMarginRate != "" {
		value, err := parseOptionalOptionRate(in.MinMarginRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MinMarginRate = value
	}
	if in.LiquidationFeeRate != "" {
		value, err := parseOptionalOptionRate(in.LiquidationFeeRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.LiquidationFeeRate = value
	}
	if in.InsuranceUserId != 0 {
		item.InsuranceUserId = in.InsuranceUserId
	}
	if in.InsuranceAccountId != 0 {
		item.InsuranceAccountId = in.InsuranceAccountId
	}
	if in.LiquidationDeficitPolicy != option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_UNKNOWN {
		item.LiquidationDeficitPolicy = int64(in.LiquidationDeficitPolicy)
	}
	if in.PhysicalDeliveryPolicy != option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_UNKNOWN {
		item.PhysicalDeliveryPolicy = int64(in.PhysicalDeliveryPolicy)
	}
	if in.PhysicalDeliveryCureSeconds != 0 {
		item.PhysicalDeliveryCureSeconds = in.PhysicalDeliveryCureSeconds
	}
	if in.TradingCalendarCode != "" {
		item.TradingCalendarCode = in.TradingCalendarCode
	}
	if in.SettlementType == option.SettlementType_SETTLEMENT_TYPE_CASH {
		item.PhysicalDeliveryPolicy = int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_UNKNOWN)
		item.PhysicalDeliveryCureSeconds = 0
	}
	if in.Status != 0 {
		item.Status = int64(in.Status)
	}
	if in.Sort != 0 {
		item.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	if in.IsDeleted != 0 {
		item.IsDeleted = int64(in.IsDeleted)
	}
	if !validateSupportedContract(item) {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if original.Status != item.Status {
		// Contract status is owned by lifecycle and audited halt/resume flows.
		// Admin parameter updates must never be an alternate listing path.
		return l.rejectGovernedContractMutation(&original, "status_bypass")
	}
	if _, seriesErr := l.svcCtx.OptionContractSeriesDetailModel.FindSeriesLaunchByContract(
		l.ctx, original.TenantId, original.Id,
	); seriesErr == nil {
		if !economicContractFieldsEqual(&original, item) ||
			original.Status != item.Status || original.IsDeleted != item.IsDeleted {
			// Series-generated economics and admission state are governed by
			// the immutable series and lifecycle launch gate.
			return l.rejectGovernedContractMutation(&original, "series_economics")
		}
	} else if !errors.Is(seriesErr, models.ErrNotFound) {
		return nil, seriesErr
	}
	if original.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) &&
		!economicContractFieldsEqual(&original, item) {
		return l.rejectGovernedContractMutation(&original, "listed_economics")
	}
	item.UpdateTimes = time.Now().Unix()
	tradingPolicyChanged := !original.MaxUserLongQty.Equal(item.MaxUserLongQty) ||
		!original.MaxUserShortQty.Equal(item.MaxUserShortQty) ||
		!original.MaxOpenInterest.Equal(item.MaxOpenInterest) ||
		!original.OrderPriceBandRatio.Equal(item.OrderPriceBandRatio) ||
		!original.CircuitBreakerRatio.Equal(item.CircuitBreakerRatio) ||
		original.GreeksMaxAgeSeconds != item.GreeksMaxAgeSeconds
	statusChanged := original.Status != item.Status
	operatorID := int64(0)
	if tradingPolicyChanged || statusChanged {
		operatorID, err = utils.GetUserIdFromMd(l.ctx)
		if err != nil {
			return nil, err
		}
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		locked, err := contractModel.FindOneForUpdate(ctx, item.Id)
		if err != nil {
			return err
		}
		if !economicContractFieldsEqual(locked, &original) ||
			locked.Status != original.Status || locked.Sort != original.Sort ||
			locked.Remark != original.Remark || locked.IsDeleted != original.IsDeleted ||
			!locked.MaxUserLongQty.Equal(original.MaxUserLongQty) ||
			!locked.MaxUserShortQty.Equal(original.MaxUserShortQty) ||
			!locked.MaxOpenInterest.Equal(original.MaxOpenInterest) ||
			!locked.OrderPriceBandRatio.Equal(original.OrderPriceBandRatio) ||
			!locked.CircuitBreakerRatio.Equal(original.CircuitBreakerRatio) ||
			locked.GreeksMaxAgeSeconds != original.GreeksMaxAgeSeconds {
			return errors.New("option contract was concurrently updated")
		}
		detailModel := models.NewTOptionContractSeriesDetailModel(conn, l.svcCtx.Config.CacheRedis)
		if _, seriesErr := detailModel.FindSeriesLaunchByContract(
			ctx, locked.TenantId, locked.Id,
		); seriesErr == nil {
			if !economicContractFieldsEqual(locked, item) ||
				locked.Status != item.Status || locked.IsDeleted != item.IsDeleted {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
		} else if !errors.Is(seriesErr, models.ErrNotFound) {
			return seriesErr
		}
		if err := contractModel.Update(ctx, item); err != nil {
			return err
		}
		if !tradingPolicyChanged && !statusChanged {
			return nil
		}
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		if tradingPolicyChanged {
			if _, err = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: item.TenantId, ContractId: item.Id,
				EventType: "TRADING_POLICY_UPDATED", Reason: "ADMIN_UPDATE",
				Detail: fmt.Sprintf(
					"long:%s->%s short:%s->%s oi:%s->%s band:%s->%s circuit:%s->%s greeksMaxAge:%d->%d",
					original.MaxUserLongQty, item.MaxUserLongQty,
					original.MaxUserShortQty, item.MaxUserShortQty,
					original.MaxOpenInterest, item.MaxOpenInterest,
					original.OrderPriceBandRatio, item.OrderPriceBandRatio,
					original.CircuitBreakerRatio, item.CircuitBreakerRatio,
					original.GreeksMaxAgeSeconds, item.GreeksMaxAgeSeconds,
				),
				OperatorId: operatorID, CreateTimes: item.UpdateTimes,
			}); err != nil {
				return err
			}
		}
		if statusChanged {
			_, err = eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
				TenantId: item.TenantId, ContractId: item.Id,
				EventType: "CONTRACT_STATUS_UPDATED", Reason: "ADMIN_UPDATE",
				Detail:     fmt.Sprintf("status:%d->%d", original.Status, item.Status),
				OperatorId: operatorID, CreateTimes: item.UpdateTimes,
			})
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, enqueueErr := range enqueueContractSchedules(l.svcCtx, item) {
		l.Errorf("enqueue option contract schedule failed, contractId=%d err=%v", item.Id, enqueueErr)
	}

	return &option.CommonResp{Base: helper.OkResp()}, nil
}

func (l *UpdateContractLogic) rejectGovernedContractMutation(
	contract *models.TOptionContract,
	reason string,
) (*option.CommonResp, error) {
	operatorID, _ := utils.GetUserIdFromMd(l.ctx)
	observability.RecordAdminRejectedMutation(contract.TenantId, "contract", reason)
	l.Errorw(
		"rejected governed option contract mutation",
		logx.Field("tenantId", contract.TenantId),
		logx.Field("contractId", contract.Id),
		logx.Field("operatorId", operatorID),
		logx.Field("reason", reason),
	)
	return &option.CommonResp{Base: helper.ErrResp(
		i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx),
	)}, nil
}

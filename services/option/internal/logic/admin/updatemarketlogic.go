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
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateMarketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMarketLogic {
	return &UpdateMarketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权行情数据
func (l *UpdateMarketLogic) UpdateMarket(in *option.UpdateMarketReq) (*option.CommonResp, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, contract.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}

	now := time.Now().Unix()
	operatorID, _ := utils.GetUserIdFromMd(l.ctx)
	market := &models.TOptionMarket{
		TenantId:    contract.TenantId,
		ContractId:  in.ContractId,
		CreateTimes: now,
	}
	previousMarkPrice := decimal.Zero
	underlyingUpdated := in.UnderlyingPrice != ""
	// Bid/ask/theoretical updates must not make a stale risk mark look fresh.
	markUpdated := in.MarkPrice != ""
	greeksUpdated := in.Iv != "" || in.Delta != "" || in.Gamma != "" ||
		in.Theta != "" || in.Vega != "" || in.Rho != "" ||
		in.RiskFreeRate != "" || in.PricingModel != ""
	if in.UnderlyingPrice != "" {
		value, err := conv.ParseDecimalField(in.UnderlyingPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.UnderlyingPriceFormatError, i18n.Translate(i18n.UnderlyingPriceFormatError, l.ctx))}, nil
		}
		market.UnderlyingPrice = value
	}
	if in.MarkPrice != "" {
		value, err := conv.ParseDecimalField(in.MarkPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MarkPriceFormatError, i18n.Translate(i18n.MarkPriceFormatError, l.ctx))}, nil
		}
		market.MarkPrice = value
	}
	if in.LastPrice != "" {
		value, err := conv.ParseDecimalField(in.LastPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.LastPriceFormatError, i18n.Translate(i18n.LastPriceFormatError, l.ctx))}, nil
		}
		market.LastPrice = value
	}
	if in.BidPrice != "" {
		value, err := conv.ParseDecimalField(in.BidPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.BidPriceFormatError, i18n.Translate(i18n.BidPriceFormatError, l.ctx))}, nil
		}
		market.BidPrice = value
	}
	if in.AskPrice != "" {
		value, err := conv.ParseDecimalField(in.AskPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.AskPriceFormatError, i18n.Translate(i18n.AskPriceFormatError, l.ctx))}, nil
		}
		market.AskPrice = value
	}
	if in.TheoreticalPrice != "" {
		value, err := conv.ParseDecimalField(in.TheoreticalPrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.TheoreticalPriceFormatError, i18n.Translate(i18n.TheoreticalPriceFormatError, l.ctx))}, nil
		}
		market.TheoreticalPrice = value
	}
	if in.IntrinsicValue != "" {
		value, err := conv.ParseDecimalField(in.IntrinsicValue)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.IntrinsicValueFormatError, i18n.Translate(i18n.IntrinsicValueFormatError, l.ctx))}, nil
		}
		market.IntrinsicValue = value
	}
	if in.TimeValue != "" {
		value, err := conv.ParseDecimalField(in.TimeValue)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.TimeValueFormatError, i18n.Translate(i18n.TimeValueFormatError, l.ctx))}, nil
		}
		market.TimeValue = value
	}
	if in.Iv != "" {
		value, err := conv.ParseDecimalField(in.Iv)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.IVFormatError, i18n.Translate(i18n.IVFormatError, l.ctx))}, nil
		}
		market.Iv = value
	}
	if in.Delta != "" {
		value, err := conv.ParseDecimalField(in.Delta)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.DeltaFormatError, i18n.Translate(i18n.DeltaFormatError, l.ctx))}, nil
		}
		market.Delta = value
	}
	if in.Gamma != "" {
		value, err := conv.ParseDecimalField(in.Gamma)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.GammaFormatError, i18n.Translate(i18n.GammaFormatError, l.ctx))}, nil
		}
		market.Gamma = value
	}
	if in.Theta != "" {
		value, err := conv.ParseDecimalField(in.Theta)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ThetaFormatError, i18n.Translate(i18n.ThetaFormatError, l.ctx))}, nil
		}
		market.Theta = value
	}
	if in.Vega != "" {
		value, err := conv.ParseDecimalField(in.Vega)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.VegaFormatError, i18n.Translate(i18n.VegaFormatError, l.ctx))}, nil
		}
		market.Vega = value
	}
	if in.Rho != "" {
		value, err := conv.ParseDecimalField(in.Rho)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.RhoFormatError, i18n.Translate(i18n.RhoFormatError, l.ctx))}, nil
		}
		market.Rho = value
	}
	if in.RiskFreeRate != "" {
		value, err := conv.ParseDecimalField(in.RiskFreeRate)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.RiskFreeRateFormatError, i18n.Translate(i18n.RiskFreeRateFormatError, l.ctx))}, nil
		}
		market.RiskFreeRate = value
	}
	if in.PricingModel != "" {
		market.PricingModel = in.PricingModel
	}
	eventTime := in.SnapshotTime
	if eventTime == 0 {
		eventTime = now
	}
	if underlyingUpdated {
		market.UnderlyingSnapshotTime = eventTime
		if in.UnderlyingSnapshotTime > 0 {
			market.UnderlyingSnapshotTime = in.UnderlyingSnapshotTime
		}
	}
	if markUpdated {
		market.MarkSnapshotTime = eventTime
		if in.MarkSnapshotTime > 0 {
			market.MarkSnapshotTime = in.MarkSnapshotTime
		}
	}
	if greeksUpdated {
		market.GreeksSnapshotTime = eventTime
		if in.GreeksSnapshotTime > 0 {
			market.GreeksSnapshotTime = in.GreeksSnapshotTime
		}
	}
	market.SnapshotTime = eventTime
	market.UpdateTimes = now

	circuitTripped := false
	var circuitHalt *models.TOptionTradingHalt
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)

		lockedContract, err := contractModel.FindOneForUpdate(ctx, contract.Id)
		if err != nil {
			return err
		}
		lockedMarket, findErr := marketModel.FindOneByTenantIdContractIdForUpdate(
			ctx, contract.TenantId, contract.Id,
		)
		switch {
		case findErr == nil:
			previousMarkPrice = lockedMarket.MarkPrice
			mergeOptionMarketPatch(lockedMarket, market, in)
			market = lockedMarket
		case errors.Is(findErr, models.ErrNotFound):
			previousMarkPrice = decimal.Zero
		default:
			return findErr
		}
		if in.MarkPrice != "" && previousMarkPrice.IsPositive() && market.MarkPrice.IsPositive() &&
			lockedContract.Status == int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
			lockedContract.CircuitBreakerRatio.IsPositive() {
			jumpRatio, trip := optionCircuitBreakerDecision(
				previousMarkPrice, market.MarkPrice, lockedContract.CircuitBreakerRatio,
			)
			if trip {
				haltNo := fmt.Sprintf("CB-%d-%d", lockedContract.Id, now)
				circuitHalt = &models.TOptionTradingHalt{
					TenantId: lockedContract.TenantId, HaltNo: haltNo,
					ActiveKey:  fmt.Sprintf("CONTRACT:%d", lockedContract.Id),
					ContractId: lockedContract.Id,
					Source:     int64(option.TradingHaltSource_TRADING_HALT_SOURCE_CIRCUIT_BREAKER),
					Status:     int64(option.TradingHaltStatus_TRADING_HALT_STATUS_ACTIVE),
					Reason:     "MARK_PRICE_JUMP", EvidenceRef: "OPTION_MARKET_UPDATE",
					StartedAt: now, CreatedBy: operatorID, CreateTimes: now, UpdateTimes: now,
				}
				result, insertErr := haltModel.Insert(ctx, circuitHalt)
				if insertErr != nil {
					return insertErr
				}
				circuitHalt.Id, insertErr = result.LastInsertId()
				if insertErr != nil {
					return insertErr
				}
				lockedContract.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
				lockedContract.UpdateTimes = now
				if err := contractModel.Update(ctx, lockedContract); err != nil {
					return err
				}
				if _, err := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
					TenantId: contract.TenantId, ContractId: contract.Id,
					EventType: "CIRCUIT_BREAKER", Reason: "MARK_PRICE_JUMP",
					Detail: fmt.Sprintf(
						"previous=%s current=%s ratio=%s threshold=%s",
						previousMarkPrice, market.MarkPrice, jumpRatio, lockedContract.CircuitBreakerRatio,
					),
					OperatorId: operatorID, CreateTimes: now,
				}); err != nil {
					return err
				}
				circuitTripped = true
			}
		}

		if market.Id == 0 {
			result, err := marketModel.Insert(ctx, market)
			if err != nil {
				return err
			}
			market.Id, _ = result.LastInsertId()
		} else if err := marketModel.Update(ctx, market); err != nil {
			return err
		}

		return helpers.InsertMarketSnapshot(
			ctx, snapshotModel, market,
			helpers.MarketSnapshotSourceAdmin, "", now,
		)
	})
	if err != nil {
		return nil, err
	}
	if circuitTripped {
		l.Errorf(
			"option trading control metric event=CIRCUIT_BREAKER reason=MARK_PRICE_JUMP tenantId=%d contractId=%d",
			contract.TenantId, contract.Id,
		)
		orders, total, findErr := l.svcCtx.OptionOrderModel.FindPage(
			l.ctx,
			models.OptionOrderPageFilter{
				TenantId: contract.TenantId, ContractId: contract.Id,
				Statuses: []int64{
					int64(option.OrderStatus_ORDER_STATUS_FUNDING),
					int64(option.OrderStatus_ORDER_STATUS_PENDING),
					int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
				},
			},
			0, 1,
		)
		_ = orders
		cancelErr := findErr
		if cancelErr == nil {
			cancelErr = applogic.CancelContractOrdersByControl(
				l.ctx, l.svcCtx, contract.TenantId, contract.Id, "CIRCUIT_BREAKER",
			)
		}
		lastError := ""
		success, failed := total, int64(0)
		if cancelErr != nil {
			lastError = cancelErr.Error()
			if len(lastError) > 500 {
				lastError = lastError[:500]
			}
			success = 0
			failed = total
		}
		updateErr := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			conn := sqlx.NewSqlConnFromSession(session)
			haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
			halt, err := haltModel.FindOneForUpdate(ctx, circuitHalt.Id)
			if err != nil {
				return err
			}
			halt.CancelTotal = total
			halt.CancelSuccess = success
			halt.CancelFailed = failed
			halt.LastErrorMsg = lastError
			halt.UpdateTimes = time.Now().Unix()
			return haltModel.Update(ctx, halt)
		})
		if updateErr != nil {
			return nil, updateErr
		}
		if cancelErr != nil {
			// The contract is already paused, so failure is safe but must remain visible.
			return nil, cancelErr
		}
	}

	return &option.CommonResp{Base: helper.OkResp()}, nil
}

func mergeOptionMarketPatch(
	target, patch *models.TOptionMarket,
	in *option.UpdateMarketReq,
) {
	if in.UnderlyingPrice != "" {
		target.UnderlyingPrice = patch.UnderlyingPrice
		target.UnderlyingSnapshotTime = patch.UnderlyingSnapshotTime
	}
	if in.MarkPrice != "" {
		target.MarkPrice = patch.MarkPrice
		target.MarkSnapshotTime = patch.MarkSnapshotTime
	}
	if in.LastPrice != "" {
		target.LastPrice = patch.LastPrice
	}
	if in.BidPrice != "" {
		target.BidPrice = patch.BidPrice
	}
	if in.AskPrice != "" {
		target.AskPrice = patch.AskPrice
	}
	if in.TheoreticalPrice != "" {
		target.TheoreticalPrice = patch.TheoreticalPrice
	}
	if in.IntrinsicValue != "" {
		target.IntrinsicValue = patch.IntrinsicValue
	}
	if in.TimeValue != "" {
		target.TimeValue = patch.TimeValue
	}
	if in.Iv != "" {
		target.Iv = patch.Iv
	}
	if in.Delta != "" {
		target.Delta = patch.Delta
	}
	if in.Gamma != "" {
		target.Gamma = patch.Gamma
	}
	if in.Theta != "" {
		target.Theta = patch.Theta
	}
	if in.Vega != "" {
		target.Vega = patch.Vega
	}
	if in.Rho != "" {
		target.Rho = patch.Rho
	}
	if in.RiskFreeRate != "" {
		target.RiskFreeRate = patch.RiskFreeRate
	}
	if in.PricingModel != "" {
		target.PricingModel = patch.PricingModel
	}
	if in.Iv != "" || in.Delta != "" || in.Gamma != "" ||
		in.Theta != "" || in.Vega != "" || in.Rho != "" ||
		in.RiskFreeRate != "" || in.PricingModel != "" {
		target.GreeksSnapshotTime = patch.GreeksSnapshotTime
	}
	target.SnapshotTime = patch.SnapshotTime
	target.UpdateTimes = patch.UpdateTimes
}

func optionCircuitBreakerDecision(
	previous, current, threshold decimal.Decimal,
) (decimal.Decimal, bool) {
	if !previous.IsPositive() || !current.IsPositive() || !threshold.IsPositive() ||
		threshold.GreaterThan(decimal.NewFromInt(1)) {
		return decimal.Zero, false
	}
	ratio := current.Sub(previous).Abs().Div(previous)
	return ratio, ratio.GreaterThanOrEqual(threshold)
}

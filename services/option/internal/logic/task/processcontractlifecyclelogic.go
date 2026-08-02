package tasklogic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"wklive/common/generate"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessContractLifecycleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessContractLifecycleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessContractLifecycleLogic {
	return &ProcessContractLifecycleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 期权合约生命周期处理（状态流转/订单过期/自动行权/到期结算）
func (l *ProcessContractLifecycleLogic) ProcessContractLifecycle(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_contract_lifecycle", func() (*option.OptionTaskResp, error) {
		now := time.Now().Unix()
		if err := l.syncContracts(option.ContractStatus_CONTRACT_STATUS_PENDING, now, 0, option.ContractStatus_CONTRACT_STATUS_TRADING, now); err != nil {
			return nil, err
		}
		if err := l.closeTradingContracts(now); err != nil {
			return nil, err
		}
		if err := l.expirePausedContracts(now); err != nil {
			return nil, err
		}
		if err := l.processExpiredContracts(now); err != nil {
			return nil, err
		}
		return helpers.OkTaskResp(), nil
	})
}

// closeTradingContracts applies the independent last-trade boundary before
// expiry processing. PENDING contracts that never passed launch gates are also
// closed, so they cannot become tradable after their approved window.
func (l *ProcessContractLifecycleLogic) closeTradingContracts(now int64) error {
	for _, status := range []option.ContractStatus{
		option.ContractStatus_CONTRACT_STATUS_PENDING,
		option.ContractStatus_CONTRACT_STATUS_TRADING,
		option.ContractStatus_CONTRACT_STATUS_PAUSED,
	} {
		cursor := int64(0)
		for {
			items, _, err := l.svcCtx.OptionContractModel.FindPage(
				l.ctx,
				models.OptionContractPageFilter{
					Status: int64(status), LastTradeTimeEnd: now,
				},
				cursor, 100,
			)
			if err != nil {
				return err
			}
			if len(items) == 0 {
				break
			}
			for _, item := range items {
				cursor = item.Id
				if err := l.closeContractTrading(
					item.Id, item.TenantId, item.LastTradeTime, now,
				); err != nil {
					return err
				}
			}
			if len(items) < 100 {
				break
			}
		}
	}
	return nil
}

// closeContractTrading is shared by the periodic scanner and delayed message.
// The row lock serializes the boundary with admission/matching transactions;
// cancellation then uses the same funding-release state machine as user and
// emergency-control cancellation. Replays are intentionally idempotent.
func (l *ProcessContractLifecycleLogic) closeContractTrading(
	contractID, tenantID, expectedLastTradeTime, now int64,
) error {
	eligible := false
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		contract, findErr := contractModel.FindOneForUpdate(ctx, contractID)
		if findErr != nil {
			return findErr
		}
		if contract.TenantId != tenantID ||
			contract.LastTradeTime != expectedLastTradeTime ||
			contract.LastTradeTime <= 0 || now < contract.LastTradeTime {
			return nil
		}
		switch option.ContractStatus(contract.Status) {
		case option.ContractStatus_CONTRACT_STATUS_PENDING,
			option.ContractStatus_CONTRACT_STATUS_TRADING:
			eligible = true
			contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
			contract.UpdateTimes = now
			return contractModel.Update(ctx, contract)
		case option.ContractStatus_CONTRACT_STATUS_PAUSED:
			eligible = true
			return nil
		default:
			return nil
		}
	})
	if err != nil {
		return err
	}
	if !eligible {
		return nil
	}
	_, _, failed, cancelErr := applogic.CancelContractOrdersByControlReport(
		l.ctx, l.svcCtx, tenantID, contractID, "CONTRACT_LAST_TRADE_ENDED", true,
	)
	if failed > 0 {
		return cancelErr
	}
	return cancelErr
}

func (l *ProcessContractLifecycleLogic) expirePausedContracts(now int64) error {
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.OptionContractModel.FindPage(
			l.ctx,
			models.OptionContractPageFilter{
				Status: int64(option.ContractStatus_CONTRACT_STATUS_PAUSED), ExpireTimeEnd: now,
			},
			cursor, 100,
		)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			// Last-trade cancellation can leave Asset release instructions in
			// flight. Keep the contract PAUSED until every order reaches a safe
			// terminal state; otherwise expiry settlement could start while an
			// order reservation is still frozen.
			unsafeOrders, checkErr := l.svcCtx.OptionOrderModel.HasUnsafeContractResumeOrders(
				l.ctx, item.TenantId, item.Id,
			)
			if checkErr != nil {
				return checkErr
			}
			if unsafeOrders {
				l.Infof(
					"keep option contract paused until last-trade order releases finish, tenantId=%d contractId=%d",
					item.TenantId, item.Id,
				)
				continue
			}
			blocked, checkErr := l.svcCtx.OptionCorporateActionContractModel.IsContractMigrationActive(
				l.ctx, item.TenantId, item.Id,
			)
			if checkErr != nil {
				return checkErr
			}
			if blocked {
				l.Infof(
					"keep paused option contract unchanged while corporate action is active, tenantId=%d contractId=%d",
					item.TenantId, item.Id,
				)
				continue
			}
			if err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
				haltModel := models.NewTOptionTradingHaltModel(conn, l.svcCtx.Config.CacheRedis)
				actionContractModel := models.NewTOptionCorporateActionContractModel(
					conn, l.svcCtx.Config.CacheRedis,
				)
				contract, findErr := contractModel.FindOneForUpdate(ctx, item.Id)
				if findErr != nil {
					return findErr
				}
				if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) ||
					contract.ExpireTime <= 0 || contract.ExpireTime > now {
					return nil
				}
				blocked, findErr = actionContractModel.IsContractMigrationActive(
					ctx, contract.TenantId, contract.Id,
				)
				if findErr != nil {
					return findErr
				}
				if blocked {
					return nil
				}
				halt, findErr := haltModel.FindActiveByContract(ctx, contract.TenantId, contract.Id)
				if findErr == nil {
					halt, findErr = haltModel.FindOneForUpdate(ctx, halt.Id)
					if findErr != nil {
						return findErr
					}
					halt.Status = int64(option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED)
					halt.ActiveKey = "HALT:" + halt.HaltNo
					halt.LiftedAt = now
					halt.LiftedBy = 0
					halt.LiftReason = "CONTRACT_EXPIRED"
					halt.UpdateTimes = now
					if updateErr := haltModel.Update(ctx, halt); updateErr != nil {
						return updateErr
					}
				} else if !errors.Is(findErr, models.ErrNotFound) {
					return findErr
				}
				contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED)
				contract.UpdateTimes = now
				return contractModel.Update(ctx, contract)
			}); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) syncContracts(from option.ContractStatus, listEnd, expireEnd int64, to option.ContractStatus, now int64) error {
	cursor := int64(0)
	for {
		items, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status:        int64(from),
			ListTimeEnd:   listEnd,
			ExpireTimeEnd: expireEnd,
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		for _, item := range items {
			cursor = item.Id
			if to == option.ContractStatus_CONTRACT_STATUS_TRADING {
				if _, err := l.listContractIfEligible(
					item.Id, item.TenantId, item.ListTime, now,
				); err != nil {
					return err
				}
				continue
			}
			item.Status = int64(to)
			item.UpdateTimes = now
			if err := l.svcCtx.OptionContractModel.Update(l.ctx, item); err != nil {
				return err
			}
		}
		if len(items) < 100 {
			return nil
		}
	}
}

// listContractIfEligible is the only PENDING -> TRADING transition used by
// both the periodic scanner and delayed list messages. Keeping the checks in
// the same row-locked transaction prevents the delay queue from bypassing
// series approval, corporate-action, market, risk, or calendar gates.
func (l *ProcessContractLifecycleLogic) listContractIfEligible(
	contractID, tenantID, expectedListTime, now int64,
) (bool, error) {
	listed := false
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		seriesDetailModel := models.NewTOptionContractSeriesDetailModel(conn, l.svcCtx.Config.CacheRedis)
		actionContractModel := models.NewTOptionCorporateActionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, l.svcCtx.Config.CacheRedis)

		contract, findErr := contractModel.FindOneForUpdate(ctx, contractID)
		if findErr != nil {
			return findErr
		}
		if contract.TenantId != tenantID || contract.ListTime != expectedListTime ||
			contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
			contract.ListTime <= 0 || now < contract.ListTime ||
			contract.LastTradeTime <= contract.ListTime || now >= contract.LastTradeTime {
			return nil
		}
		series, seriesErr := seriesDetailModel.FindSeriesLaunchByContract(
			ctx, contract.TenantId, contract.Id,
		)
		if seriesErr == nil {
			if !seriesContractLaunchApproved(series) {
				l.Infof(
					"keep series-generated option contract pending until launch approval, tenantId=%d contractId=%d seriesId=%d",
					contract.TenantId, contract.Id, series.Id,
				)
				return nil
			}
		} else if !errors.Is(seriesErr, models.ErrNotFound) {
			return seriesErr
		}
		blocked, findErr := actionContractModel.IsSuccessorBlocked(
			ctx, contract.TenantId, contract.Id,
		)
		if findErr != nil {
			return findErr
		}
		if blocked {
			l.Infof(
				"keep option successor contract pending until corporate action completes, tenantId=%d contractId=%d",
				contract.TenantId, contract.Id,
			)
			return nil
		}
		market, marketErr := marketModel.FindOneByTenantIdContractId(
			ctx, contract.TenantId, contract.Id,
		)
		if marketErr != nil {
			if errors.Is(marketErr, models.ErrNotFound) {
				l.Infof(
					"keep option contract pending because market is missing, tenantId=%d contractId=%d",
					contract.TenantId, contract.Id,
				)
				return nil
			}
			return marketErr
		}
		if !contractLaunchControlsReady(contract, market, now) {
			l.Infof(
				"keep option contract pending because risk controls or market freshness are incomplete, tenantId=%d contractId=%d",
				contract.TenantId, contract.Id,
			)
			return nil
		}
		code, valid := helpers.NormalizeTradingCalendarCode(contract.TradingCalendarCode)
		if !valid {
			l.Errorf(
				"keep option contract pending because trading calendar code is invalid, tenantId=%d contractId=%d code=%q",
				contract.TenantId, contract.Id, contract.TradingCalendarCode,
			)
			return nil
		}
		if _, findErr = calendarModel.FindEffective(ctx, contract.TenantId, code, now); findErr != nil {
			l.Errorf(
				"keep option contract pending because no unambiguous effective calendar exists, tenantId=%d contractId=%d code=%s err=%v",
				contract.TenantId, contract.Id, code, findErr,
			)
			return nil
		}
		contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_TRADING)
		contract.UpdateTimes = now
		if updateErr := contractModel.Update(ctx, contract); updateErr != nil {
			return updateErr
		}
		listed = true
		return nil
	})
	return listed, err
}

func seriesContractLaunchApproved(series *models.TOptionContractSeries) bool {
	return series != nil &&
		series.Status == int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED) &&
		series.LaunchStatus == int64(
			option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_APPROVED,
		)
}

func contractLaunchControlsReady(
	contract *models.TOptionContract, market *models.TOptionMarket, now int64,
) bool {
	return contract != nil &&
		contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_PENDING) &&
		contract.IsDeleted == int64(common.YesNo_YES_NO_NO) &&
		contract.MaxUserLongQty.IsPositive() &&
		contract.MaxUserShortQty.IsPositive() &&
		contract.MaxOpenInterest.IsPositive() &&
		contract.OrderPriceBandRatio.IsPositive() &&
		contract.OrderPriceBandRatio.LessThanOrEqual(decimal.NewFromInt(1)) &&
		contract.CircuitBreakerRatio.IsPositive() &&
		contract.CircuitBreakerRatio.LessThanOrEqual(decimal.NewFromInt(1)) &&
		contract.GreeksMaxAgeSeconds > 0 &&
		contract.ListTime > 0 &&
		contract.LastTradeTime > contract.ListTime &&
		now < contract.LastTradeTime &&
		contract.ExerciseCutoffTime >= contract.LastTradeTime &&
		contract.ExpireTime >= contract.ExerciseCutoffTime &&
		contract.DeliverTime >= contract.ExpireTime &&
		helpers.IsRiskMarketFresh(market, now, 30) &&
		helpers.IsGreeksFresh(market, now, contract.GreeksMaxAgeSeconds)
}

func (l *ProcessContractLifecycleLogic) processExpiredContracts(now int64) error {
	cursor := int64(0)
	for {
		contracts, _, err := l.svcCtx.OptionContractModel.FindPage(l.ctx, models.OptionContractPageFilter{
			Status: int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(contracts) == 0 {
			return nil
		}
		for _, contract := range contracts {
			cursor = contract.Id
			if err := l.expireContractOrders(contract, now); err != nil {
				return err
			}
			incomplete, err := l.svcCtx.OptionOutboxModel.HasIncomplete(l.ctx, contract.TenantId, contract.Id)
			if err != nil {
				return err
			}
			if incomplete {
				l.Infof("delay option settlement for incomplete trade events, contractId=%d", contract.Id)
				continue
			}
			incomplete, err = l.svcCtx.OptionAssetInstructionModel.HasIncompleteMarginForContract(
				l.ctx, contract.TenantId, contract.Id,
			)
			if err != nil {
				return err
			}
			if incomplete {
				l.Infof("delay option settlement for incomplete margin instructions, contractId=%d", contract.Id)
				continue
			}
			pendingExercise, err := l.svcCtx.OptionExerciseModel.HasPendingByContract(
				l.ctx, contract.TenantId, contract.Id,
			)
			if err != nil {
				return err
			}
			if pendingExercise {
				l.Infof("delay option settlement for pending early exercises, contractId=%d", contract.Id)
				continue
			}
			settlementPrice, err := l.lockSettlementPrice(contract, now)
			if err != nil {
				return err
			}
			if settlementPrice == nil ||
				settlementPrice.Status != int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED) {
				l.Errorf("option settlement waiting for authoritative price, contractId=%d", contract.Id)
				continue
			}
			if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) {
				if err := l.autoExerciseContract(contract, settlementPrice.DeliveryPrice, now); err != nil {
					return err
				}
			}
			if contract.DeliverTime > now {
				continue
			}
			if err := l.settleContract(contract, settlementPrice, now); err != nil {
				return err
			}
		}
		if len(contracts) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) expireContractOrders(contract *models.TOptionContract, now int64) error {
	cursor := int64(0)
	for {
		orders, _, err := l.svcCtx.OptionOrderModel.FindPage(l.ctx, models.OptionOrderPageFilter{
			ContractId: contract.Id,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(orders) == 0 {
			return nil
		}
		for _, order := range orders {
			cursor = order.Id
			order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRED)
			if order.MarginAmount.IsPositive() {
				order.Status = int64(option.OrderStatus_ORDER_STATUS_EXPIRING)
			}
			order.CancelReason = "CONTRACT_EXPIRED"
			order.CancelTime = now
			order.UpdateTimes = now
			err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
				conn := sqlx.NewSqlConnFromSession(session)
				orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
				positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
				instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
				if err := releaseClosePositionFrozenQty(ctx, positionModel, order, order.UnfilledQty, now); err != nil {
					return err
				}
				if order.MarginAmount.IsPositive() {
					if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
						TenantId: order.TenantId, InstructionNo: order.OrderNo + "-EXPIRE-RELEASE",
						BizNo: order.OrderNo, OrderId: order.Id, UserId: order.UserId, AccountId: order.AccountId,
						Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
						TargetBizNo: order.OrderNo, Coin: applogic.OptionOrderMarginCoin(order), Amount: order.MarginAmount,
						StepNo: 2, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
						ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
						CreateTimes:          now, UpdateTimes: now,
					}); err != nil {
						return err
					}
				}
				return orderModel.Update(ctx, order)
			})
			if err != nil {
				return err
			}
			publishOptionOrderChanged(l.ctx, l.svcCtx, order)
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) lockSettlementPrice(contract *models.TOptionContract, now int64) (*models.TOptionSettlementPrice, error) {
	item, err := l.svcCtx.OptionSettlementPriceModel.FindLatest(l.ctx, contract.TenantId, contract.Id)
	if err == nil {
		switch option.SettlementPriceStatus(item.Status) {
		case option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING:
			if validateErr := l.validateSettlementPriceForUse(contract, item, false); validateErr != nil {
				return nil, fmt.Errorf("pending option settlement price evidence is invalid: %w", validateErr)
			}
			return item, nil
		case option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED:
			if validateErr := l.validateSettlementPriceForUse(contract, item, true); validateErr != nil {
				return nil, fmt.Errorf("confirmed option settlement price evidence is invalid: %w", validateErr)
			}
			return item, nil
		case option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_REJECTED,
			option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_SUPERSEDED:
			confirmed, confirmedErr := l.svcCtx.OptionSettlementPriceModel.FindLatestConfirmed(
				l.ctx, contract.TenantId, contract.Id,
			)
			if confirmedErr == nil {
				if validateErr := l.validateSettlementPriceForUse(contract, confirmed, true); validateErr != nil {
					return nil, fmt.Errorf("confirmed option settlement price evidence is invalid: %w", validateErr)
				}
				return confirmed, nil
			}
			if !errors.Is(confirmedErr, models.ErrNotFound) {
				return nil, confirmedErr
			}
			err = models.ErrNotFound
		}
	}
	if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if contract.ExpireTime <= 0 || now < contract.ExpireTime {
		return nil, nil
	}
	priceSource := strings.ToLower(strings.TrimSpace(contract.SettlementPriceSource))
	if priceSource == "" {
		priceSource = helpers.SettlementPriceSourceAutomatic
	}
	method := strings.ToUpper(strings.TrimSpace(contract.SettlementPriceMethod))
	if method == "" {
		method = helpers.SettlementPriceMethodAutomatic
	}
	windowSeconds := contract.SettlementWindowSeconds
	if windowSeconds == 0 {
		windowSeconds = 60
	}
	minSamples := contract.SettlementMinSamples
	if minSamples == 0 {
		minSamples = 3
	}
	if priceSource != helpers.SettlementPriceSourceAutomatic || method != helpers.SettlementPriceMethodAutomatic ||
		windowSeconds < 1 || minSamples < 1 {
		return nil, fmt.Errorf(
			"unsupported option settlement price rule, contractId=%d source=%s method=%s window=%d minSamples=%d",
			contract.Id, priceSource, method, windowSeconds, minSamples,
		)
	}
	windowStart := contract.ExpireTime - windowSeconds
	windowEnd := contract.ExpireTime
	samples, err := l.loadSettlementPriceSamples(contract, windowStart, windowEnd)
	if err != nil {
		return nil, err
	}
	if int64(len(samples)) < minSamples {
		return nil, nil
	}
	deliveryPrice := calculateSettlementMedian(samples)
	if !deliveryPrice.IsPositive() {
		return nil, nil
	}
	sourceIDs := make([]string, 0, len(samples))
	for _, sample := range samples {
		sourceIDs = append(sourceIDs, sample.sourceSnapshotID)
	}
	sourceSnapshotJSON, err := json.Marshal(sourceIDs)
	if err != nil {
		return nil, err
	}
	version := int64(1)
	if item != nil {
		version = item.Version + 1
	}
	latest := item
	item = &models.TOptionSettlementPrice{
		TenantId: contract.TenantId, ContractId: contract.Id,
		PriceSource: priceSource, WindowStart: windowStart,
		WindowEnd: windowEnd, SampleCount: int64(len(samples)), CalculationMethod: method,
		DeliveryPrice:     deliveryPrice,
		SourceSnapshotIds: string(sourceSnapshotJSON),
		Version:           version,
		Status:            int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING),
		ChangeReason:      "system calculation awaiting independent review",
		CreateTimes:       now, UpdateTimes: now,
	}
	_, canonicalSourceIDs, normalizeErr := helpers.NormalizeSettlementPriceSourceIDs(item.SourceSnapshotIds)
	if normalizeErr != nil {
		return nil, normalizeErr
	}
	item.SourceSnapshotIds = canonicalSourceIDs
	if validateErr := helpers.ValidateSettlementPriceEvidence(contract, item, false); validateErr != nil {
		return nil, fmt.Errorf("calculated option settlement price evidence is invalid: %w", validateErr)
	}
	if shouldSuppressRejectedSettlementPrice(latest, item) {
		// A reviewer rejection must not be undone by recreating the identical
		// automatic candidate every task cycle. New immutable evidence can still
		// produce a new version, while an operator may create a manual correction.
		return latest, nil
	}
	result, err := l.svcCtx.OptionSettlementPriceModel.Insert(l.ctx, item)
	if err != nil {
		existing, findErr := l.svcCtx.OptionSettlementPriceModel.FindLatest(l.ctx, contract.TenantId, contract.Id)
		if findErr == nil {
			requireConfirmed := existing.Status == int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_CONFIRMED)
			if existing.Status == int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_PENDING) || requireConfirmed {
				if validateErr := l.validateSettlementPriceForUse(contract, existing, requireConfirmed); validateErr != nil {
					return nil, fmt.Errorf("concurrent option settlement price evidence is invalid: %w", validateErr)
				}
			}
			return existing, nil
		}
		return nil, err
	}
	item.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return item, nil
}

func shouldSuppressRejectedSettlementPrice(
	latest, candidate *models.TOptionSettlementPrice,
) bool {
	return latest != nil &&
		latest.Status == int64(option.SettlementPriceStatus_SETTLEMENT_PRICE_STATUS_REJECTED) &&
		helpers.SameSettlementPriceEvidence(latest, candidate)
}

type settlementPriceSample struct {
	id               int64
	snapshotTime     int64
	sourceSnapshotID string
	price            decimal.Decimal
}

func (l *ProcessContractLifecycleLogic) loadSettlementPriceSamples(
	contract *models.TOptionContract,
	windowStart, windowEnd int64,
) ([]settlementPriceSample, error) {
	var samples []settlementPriceSample
	cursor := int64(0)
	for {
		rows, _, err := l.svcCtx.OptionMarketSnapshotModel.FindPage(
			l.ctx,
			models.OptionMarketSnapshotPageFilter{
				TenantId: contract.TenantId, ContractId: contract.Id,
				SnapshotStart: windowStart, SnapshotEnd: windowEnd,
				SourceType: helpers.MarketSnapshotSourceUnderlying,
			},
			cursor, 100,
		)
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			cursor = row.Id
			if !row.UnderlyingPrice.IsPositive() || strings.TrimSpace(row.SourceSnapshotId) == "" {
				continue
			}
			samples = append(samples, settlementPriceSample{
				id: row.Id, snapshotTime: row.SnapshotTime,
				sourceSnapshotID: row.SourceSnapshotId,
				price:            row.UnderlyingPrice,
			})
		}
		if len(rows) < 100 {
			break
		}
	}
	sort.Slice(samples, func(i, j int) bool {
		if samples[i].snapshotTime == samples[j].snapshotTime {
			return samples[i].id < samples[j].id
		}
		return samples[i].snapshotTime < samples[j].snapshotTime
	})
	return samples, nil
}

func (l *ProcessContractLifecycleLogic) validateSettlementPriceForUse(
	contract *models.TOptionContract,
	price *models.TOptionSettlementPrice,
	requireConfirmed bool,
) error {
	if err := helpers.ValidateSettlementPriceEvidence(contract, price, requireConfirmed); err != nil {
		return err
	}
	if price.PriceSource != helpers.SettlementPriceSourceAutomatic ||
		price.CalculationMethod != helpers.SettlementPriceMethodAutomatic {
		return nil
	}
	sourceIDs, _, err := helpers.NormalizeSettlementPriceSourceIDs(price.SourceSnapshotIds)
	if err != nil {
		return err
	}
	samples, err := l.loadSettlementPriceSamples(contract, price.WindowStart, price.WindowEnd)
	if err != nil {
		return err
	}
	bySourceID := make(map[string]settlementPriceSample, len(samples))
	for _, sample := range samples {
		sourceID := strings.TrimSpace(sample.sourceSnapshotID)
		if _, duplicate := bySourceID[sourceID]; duplicate {
			return fmt.Errorf("duplicate authoritative settlement snapshot id: %s", sourceID)
		}
		bySourceID[sourceID] = sample
	}
	selected := make([]settlementPriceSample, 0, len(sourceIDs))
	for _, sourceID := range sourceIDs {
		sample, exists := bySourceID[sourceID]
		if !exists {
			return fmt.Errorf("settlement price snapshot evidence is missing: %s", sourceID)
		}
		selected = append(selected, sample)
	}
	calculated := calculateSettlementMedian(selected)
	if !calculated.Equal(price.DeliveryPrice) {
		return fmt.Errorf(
			"settlement price does not match immutable evidence: stored=%s calculated=%s",
			price.DeliveryPrice, calculated,
		)
	}
	return nil
}

func calculateSettlementMedian(samples []settlementPriceSample) decimal.Decimal {
	if len(samples) == 0 {
		return decimal.Zero
	}
	prices := make([]decimal.Decimal, 0, len(samples))
	for _, sample := range samples {
		if sample.price.IsPositive() {
			prices = append(prices, sample.price)
		}
	}
	if len(prices) == 0 {
		return decimal.Zero
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LessThan(prices[j]) })
	middle := len(prices) / 2
	if len(prices)%2 == 1 {
		return prices[middle].Round(16)
	}
	return prices[middle-1].Add(prices[middle]).Div(decimal.NewFromInt(2)).Round(16)
}

func (l *ProcessContractLifecycleLogic) autoExerciseContract(contract *models.TOptionContract, deliveryPrice decimal.Decimal, now int64) error {
	intrinsicValue := optionIntrinsicValue(contract, deliveryPrice)
	cursor := int64(0)
	for {
		positions, _, err := l.svcCtx.OptionPositionModel.FindPage(l.ctx, models.OptionPositionPageFilter{
			ContractId: contract.Id,
			Status:     int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		}, cursor, 100)
		if err != nil {
			return err
		}
		if len(positions) == 0 {
			return nil
		}
		for _, position := range positions {
			cursor = position.Id
			if err := l.autoExercisePosition(contract, position.Id, deliveryPrice, intrinsicValue, now); err != nil {
				return err
			}
		}
		if len(positions) < 100 {
			return nil
		}
	}
}

func (l *ProcessContractLifecycleLogic) autoExercisePosition(contract *models.TOptionContract, positionId int64, deliveryPrice, intrinsicValue decimal.Decimal, now int64) error {
	// Generate outside the database transaction so a Redis/network round-trip
	// cannot hold the position row lock.
	exerciseNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "EX", "")
	if err != nil {
		return err
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		exerciseModel := models.NewTOptionExerciseModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionExerciseInstructionModel(conn, l.svcCtx.Config.CacheRedis)

		position, err := positionModel.FindOneForUpdate(ctx, positionId)
		if err != nil {
			return err
		}
		if position.Status != int64(option.PositionStatus_POSITION_STATUS_HOLDING) {
			return nil
		}
		instructionType := option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO
		instruction, instructionErr := instructionModel.FindLatestByPosition(
			ctx, position.TenantId, position.Id,
		)
		if instructionErr == nil {
			if instruction.Status != int64(option.ExerciseInstructionStatus_EXERCISE_INSTRUCTION_STATUS_ACTIVE) {
				return fmt.Errorf("latest exercise instruction is not active: positionId=%d", position.Id)
			}
			instructionType = option.ExerciseInstructionType(instruction.InstructionType)
		} else if !errors.Is(instructionErr, models.ErrNotFound) {
			return instructionErr
		}
		payoff := optionSettlementPayoff(contract, deliveryPrice, position.ExerciseableQty)
		fee := payoff.Mul(contract.ExerciseFeeRate).Round(16)
		if position.Side != int64(common.PositionSide_POSITION_SIDE_LONG) ||
			!position.ExerciseableQty.IsPositive() ||
			!shouldExerciseAtExpiry(contract, intrinsicValue, payoff, fee, instructionType) {
			position.Status = int64(option.PositionStatus_POSITION_STATUS_EXPIRED)
			position.ExerciseableQty = decimal.Zero
			position.UpdateTimes = now
			return positionModel.Update(ctx, position)
		}

		exists, _, err := exerciseModel.FindPage(ctx, models.OptionExercisePageFilter{
			TenantId: position.TenantId, PositionId: position.Id,
			ExerciseType: int64(option.ExerciseType_EXERCISE_TYPE_AUTO),
		}, 0, 1)
		if err != nil {
			return err
		}
		if len(exists) > 0 {
			return nil
		}
		if _, err = exerciseModel.Insert(ctx, &models.TOptionExercise{
			TenantId: position.TenantId, ExerciseNo: exerciseNo,
			ClientExerciseId: fmt.Sprintf("AUTO-%d-%d", position.Id, contract.ExpireTime),
			UserId:           position.UserId,
			AccountId:        position.AccountId, ContractId: contract.Id, PositionId: position.Id,
			ExerciseType: int64(option.ExerciseType_EXERCISE_TYPE_AUTO), ExerciseQty: position.ExerciseableQty,
			StrikePrice: contract.StrikePrice, SettlementPrice: deliveryPrice,
			ExerciseAmount: optionExerciseAmount(contract, position.ExerciseableQty),
			ProfitAmount:   payoff,
			Fee:            fee,
			FeeCoin:        contract.SettleCoin, Status: int64(option.ExerciseStatus_EXERCISE_STATUS_DONE),
			Remark:       fmt.Sprintf("option expiry exercise instruction=%s", instructionType.String()),
			ExerciseTime: now, FinishTime: now,
			CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		position.Status = int64(option.PositionStatus_POSITION_STATUS_EXERCISED)
		position.ExerciseableQty = decimal.Zero
		position.UpdateTimes = now
		return positionModel.Update(ctx, position)
	})
}

func shouldExerciseAtExpiry(
	contract *models.TOptionContract,
	intrinsicValue, payoff, fee decimal.Decimal,
	instructionType option.ExerciseInstructionType,
) bool {
	if contract == nil || !intrinsicValue.IsPositive() || !payoff.Sub(fee).IsPositive() {
		return false
	}
	switch instructionType {
	case option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_DO_NOT_EXERCISE:
		return false
	case option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_EXERCISE:
		return true
	case option.ExerciseInstructionType_EXERCISE_INSTRUCTION_TYPE_AUTO:
		return intrinsicValue.GreaterThanOrEqual(contract.AutoExerciseThreshold)
	default:
		return false
	}
}

func (l *ProcessContractLifecycleLogic) settleContract(contract *models.TOptionContract, settlementPrice *models.TOptionSettlementPrice, now int64) error {
	_, err := l.svcCtx.OptionSettlementModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err == nil {
		return nil
	}
	if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, contract.Id)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	deliveryPrice := settlementPrice.DeliveryPrice
	theoreticalPrice := decimal.Zero
	iv := decimal.Zero
	isITM := int64(common.YesNo_YES_NO_NO)
	if market != nil {
		theoreticalPrice = market.TheoreticalPrice
		iv = market.Iv
	}
	if (contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) && deliveryPrice.GreaterThan(contract.StrikePrice)) ||
		(contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) && deliveryPrice.LessThan(contract.StrikePrice)) {
		isITM = int64(common.YesNo_YES_NO_YES)
	}
	exerciseResult := int64(option.ExerciseResult_EXERCISE_RESULT_NONE)
	if contract.IsAutoExercise == int64(common.YesNo_YES_NO_YES) {
		if isITM == int64(common.YesNo_YES_NO_YES) {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_EXERCISE)
		} else {
			exerciseResult = int64(option.ExerciseResult_EXERCISE_RESULT_AUTO_ABANDON)
		}
	}
	settlementNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OPS", "")
	if err != nil {
		return err
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		settlementModel := models.NewTOptionSettlementModel(conn, l.svcCtx.Config.CacheRedis)
		batchModel := models.NewTOptionSettlementBatchModel(conn, l.svcCtx.Config.CacheRedis)
		detailModel := models.NewTOptionSettlementDetailModel(conn, l.svcCtx.Config.CacheRedis)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		physicalUnitModel := models.NewTOptionPhysicalDeliveryUnitModel(conn, l.svcCtx.Config.CacheRedis)

		batchResult, err := batchModel.Insert(ctx, &models.TOptionSettlementBatch{
			TenantId: contract.TenantId, BatchNo: settlementNo, ContractId: contract.Id,
			SettlementPriceId: settlementPrice.Id,
			Status:            int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_CALCULATING),
			CreateTimes:       now, UpdateTimes: now,
		})
		if err != nil {
			return err
		}
		batchId, err := batchResult.LastInsertId()
		if err != nil {
			return err
		}
		result, err := settlementModel.Insert(ctx, &models.TOptionSettlement{
			TenantId:         contract.TenantId,
			SettlementNo:     settlementNo,
			ContractId:       contract.Id,
			UnderlyingSymbol: contract.UnderlyingSymbol,
			ExpireTime:       contract.ExpireTime,
			SettlementTime:   now,
			DeliveryPrice:    deliveryPrice,
			TheoreticalPrice: theoreticalPrice,
			Iv:               iv,
			IsItm:            isITM,
			ExerciseResult:   exerciseResult,
			Status:           int64(option.SettlementStatus_SETTLEMENT_STATUS_PROCESSING),
			Remark:           "option settlement task",
			CreateTimes:      now,
			UpdateTimes:      now,
		})
		if err != nil {
			return err
		}
		settlementId, err := result.LastInsertId()
		if err != nil {
			return err
		}
		summary, err := settleContractPositions(
			ctx, positionModel, detailModel, instructionModel, marginLotModel,
			physicalUnitModel, contract, batchId, settlementNo, deliveryPrice, now,
		)
		if err != nil {
			return err
		}
		batch, err := batchModel.FindOne(ctx, batchId)
		if err != nil {
			return err
		}
		batch.TotalCredit = summary.totalCredit
		batch.TotalDebit = summary.totalDebit
		batch.InstructionCount = summary.instructionCount
		batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_INSTRUCTIONS_CREATED)
		if summary.instructionCount == 0 {
			batch.Status = int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE)
		}
		batch.UpdateTimes = now
		if err := batchModel.Update(ctx, batch); err != nil {
			return err
		}
		if err := helpers.InsertMarketSnapshot(
			ctx, snapshotModel, market,
			helpers.MarketSnapshotSourceSettlement, settlementNo, now,
		); err != nil {
			return err
		}
		if summary.instructionCount == 0 {
			settlement, err := settlementModel.FindOne(ctx, settlementId)
			if err != nil {
				return err
			}
			settlement.Status = int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE)
			settlement.UpdateTimes = now
			if err := settlementModel.Update(ctx, settlement); err != nil {
				return err
			}
			contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_SETTLED)
			contract.UpdateTimes = now
			return contractModel.Update(ctx, contract)
		}
		return nil
	})
}

type optionSettlementSummary struct {
	totalCredit      decimal.Decimal
	totalDebit       decimal.Decimal
	instructionCount int64
}

func settleContractPositions(ctx context.Context, positionModel models.TOptionPositionModel, detailModel models.TOptionSettlementDetailModel, instructionModel models.TOptionAssetInstructionModel, marginLotModel models.TOptionMarginLotModel, physicalUnitModel models.TOptionPhysicalDeliveryUnitModel, contract *models.TOptionContract, batchId int64, settlementNo string, deliveryPrice decimal.Decimal, now int64) (optionSettlementSummary, error) {
	if contract.SettlementType == int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL) {
		return settlePhysicalContractPositions(
			ctx, positionModel, detailModel, instructionModel, marginLotModel, physicalUnitModel,
			contract, batchId, settlementNo, deliveryPrice, now,
		)
	}
	positions, err := findCashSettlementPositions(ctx, positionModel, contract.Id)
	if err != nil {
		return optionSettlementSummary{}, err
	}
	shortExerciseQty, err := allocateExpiryShortQuantity(positions)
	if err != nil {
		return optionSettlementSummary{}, fmt.Errorf(
			"option expiry assignment failed: contractId=%d: %w", contract.Id, err,
		)
	}
	summary := optionSettlementSummary{}
	for _, position := range positions {
		qty := position.PositionQty
		exerciseQty := decimal.Zero
		switch position.Side {
		case int64(common.PositionSide_POSITION_SIDE_LONG):
			if position.Status == int64(option.PositionStatus_POSITION_STATUS_EXERCISED) {
				exerciseQty = qty
			}
		case int64(common.PositionSide_POSITION_SIDE_SHORT):
			exerciseQty = shortExerciseQty[position.Id]
		}
		payoff := optionSettlementPayoff(contract, deliveryPrice, exerciseQty)
		changeAmount := decimal.Zero
		exerciseFee := decimal.Zero
		if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) && payoff.IsPositive() {
			exerciseFee = payoff.Mul(contract.ExerciseFeeRate).Round(16)
			changeAmount = payoff.Sub(exerciseFee)
			summary.totalCredit = summary.totalCredit.Add(payoff)
		} else if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) && payoff.IsPositive() {
			changeAmount = payoff.Neg()
			summary.totalDebit = summary.totalDebit.Add(payoff)
		}

		applyPositionSettlementReturn(
			position, payoff, exerciseFee, qty, optionMultiplier(contract),
		)
		position.PositionQty = decimal.Zero
		position.AvailableQty = decimal.Zero
		position.FrozenQty = decimal.Zero
		position.PositionValue = decimal.Zero
		position.MarginAmount = decimal.Zero
		position.MaintenanceMargin = decimal.Zero
		position.UnrealizedPnl = decimal.Zero
		position.ExerciseableQty = decimal.Zero
		position.Status = int64(option.PositionStatus_POSITION_STATUS_SETTLED)
		position.LastCalcTime = now
		position.UpdateTimes = now
		if err := positionModel.Update(ctx, position); err != nil {
			return optionSettlementSummary{}, err
		}
		direction := option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_ABANDON
		instructionNo := ""
		if changeAmount.IsPositive() {
			direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_CREDIT
			instructionNo = fmt.Sprintf("%s-P%d-CREDIT", settlementNo, position.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: instructionNo,
				BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
				Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
				Coin:   contract.SettleCoin, Amount: changeAmount, StepNo: 2,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return optionSettlementSummary{}, err
			}
			summary.instructionCount++
		} else if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
			if payoff.IsPositive() {
				direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_DEBIT
			}
			var count int64
			instructionNo, count, err = createShortSettlementInstructions(
				ctx, instructionModel, marginLotModel, contract, position, settlementNo, payoff, now,
			)
			if err != nil {
				return optionSettlementSummary{}, err
			}
			summary.instructionCount += count
		}
		if exerciseFee.IsPositive() {
			if contract.FeeUserId <= 0 || contract.FeeAccountId <= 0 {
				return optionSettlementSummary{}, fmt.Errorf("option exercise fee account is missing: contractId=%d", contract.Id)
			}
			feeInstructionNo := fmt.Sprintf("%s-P%d-FEE", settlementNo, position.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: feeInstructionNo,
				BizNo: settlementNo, PositionId: position.Id,
				UserId: contract.FeeUserId, AccountId: contract.FeeAccountId,
				Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
				Coin:   contract.SettleCoin, Amount: exerciseFee, StepNo: 2,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return optionSettlementSummary{}, err
			}
			summary.instructionCount++
			if instructionNo == "" {
				instructionNo = feeInstructionNo
			}
		}
		if _, err := detailModel.Insert(ctx, &models.TOptionSettlementDetail{
			TenantId: position.TenantId, BatchId: batchId, BatchNo: settlementNo,
			ContractId: contract.Id, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
			Side: position.Side, Quantity: exerciseQty, Payoff: payoff, Direction: int64(direction),
			InstructionNo: instructionNo, CreateTimes: now,
		}); err != nil {
			return optionSettlementSummary{}, err
		}
	}
	return summary, validateOptionSettlementBalance(contract.Id, summary)
}

func findCashSettlementPositions(
	ctx context.Context,
	positionModel models.TOptionPositionModel,
	contractID int64,
) ([]*models.TOptionPosition, error) {
	const pageSize int64 = 100
	cursor := int64(0)
	var result []*models.TOptionPosition
	for {
		items, _, err := positionModel.FindPage(ctx, models.OptionPositionPageFilter{
			ContractId: contractID,
			Statuses: []int64{
				int64(option.PositionStatus_POSITION_STATUS_HOLDING),
				int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
				int64(option.PositionStatus_POSITION_STATUS_EXPIRED),
			},
		}, cursor, pageSize)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
		if int64(len(items)) < pageSize {
			return result, nil
		}
		cursor = items[len(items)-1].Id
	}
}

func allocateExpiryShortQuantity(
	positions []*models.TOptionPosition,
) (map[int64]decimal.Decimal, error) {
	required := decimal.Zero
	var shorts []*models.TOptionPosition
	for _, position := range positions {
		if position == nil || !position.PositionQty.IsPositive() {
			continue
		}
		if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) &&
			position.Status == int64(option.PositionStatus_POSITION_STATUS_EXERCISED) {
			required = required.Add(position.PositionQty)
		}
		if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
			shorts = append(shorts, position)
		}
	}
	sort.Slice(shorts, func(i, j int) bool {
		if shorts[i].CreateTimes == shorts[j].CreateTimes {
			return shorts[i].Id < shorts[j].Id
		}
		return shorts[i].CreateTimes < shorts[j].CreateTimes
	})
	result := make(map[int64]decimal.Decimal, len(shorts))
	remaining := required
	for _, short := range shorts {
		assigned := decimal.Min(short.PositionQty, remaining)
		result[short.Id] = assigned
		remaining = remaining.Sub(assigned)
		if !remaining.IsPositive() {
			remaining = decimal.Zero
		}
	}
	if remaining.IsPositive() {
		return nil, fmt.Errorf(
			"exercised long quantity exceeds remaining short quantity: missing=%s",
			remaining,
		)
	}
	return result, nil
}

func settlePhysicalContractPositions(
	ctx context.Context,
	positionModel models.TOptionPositionModel,
	detailModel models.TOptionSettlementDetailModel,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	physicalUnitModel models.TOptionPhysicalDeliveryUnitModel,
	contract *models.TOptionContract,
	batchId int64,
	settlementNo string,
	deliveryPrice decimal.Decimal,
	now int64,
) (optionSettlementSummary, error) {
	if contract.PhysicalDeliveryPolicy != int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT) ||
		contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY) ||
		contract.UnderlyingCoin == "" || contract.PhysicalDeliveryCureSeconds <= 0 {
		return optionSettlementSummary{}, errors.New("physical delivery contract is not strictly collateralized")
	}
	itm := optionIntrinsicValue(contract, deliveryPrice).IsPositive()
	positions, err := findPhysicalSettlementPositions(ctx, positionModel, contract.Id)
	if err != nil {
		return optionSettlementSummary{}, err
	}
	summary := optionSettlementSummary{}
	firstInstruction := make(map[int64]string)
	if itm {
		allocations, allocationErr := allocatePhysicalDeliveryPositions(positions)
		if allocationErr != nil {
			return optionSettlementSummary{}, fmt.Errorf(
				"physical delivery allocation failed: contractId=%d: %w",
				contract.Id, allocationErr,
			)
		}
		for i, allocation := range allocations {
			longInstruction, shortInstruction, count, createErr := createPhysicalDeliveryUnit(
				ctx, instructionModel, marginLotModel, physicalUnitModel,
				contract, allocation, batchId, settlementNo, int64(i+1), now,
			)
			if createErr != nil {
				return optionSettlementSummary{}, createErr
			}
			if firstInstruction[allocation.long.Id] == "" {
				firstInstruction[allocation.long.Id] = longInstruction
			}
			if firstInstruction[allocation.short.Id] == "" {
				firstInstruction[allocation.short.Id] = shortInstruction
			}
			paymentAmount := optionExerciseAmount(contract, allocation.quantity)
			summary.totalCredit = summary.totalCredit.Add(paymentAmount)
			summary.totalDebit = summary.totalDebit.Add(paymentAmount)
			summary.instructionCount += count
		}
	}
	for _, position := range positions {
		qty := position.PositionQty
		deliveryQty := qty.Mul(optionMultiplier(contract)).Round(16)
		paymentAmount := optionExerciseAmount(contract, qty)
		direction := option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_ABANDON
		if itm && position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_CREDIT
		} else if itm && position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
			direction = option.SettlementDetailDirection_SETTLEMENT_DETAIL_DIRECTION_DEBIT
		} else if !itm && position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
			instructionNo, count, releaseErr := createCoveredShortSettlementInstructions(
				ctx, instructionModel, marginLotModel, contract, position,
				settlementNo, deliveryQty, paymentAmount, false, now,
			)
			if releaseErr != nil {
				return optionSettlementSummary{}, releaseErr
			}
			firstInstruction[position.Id] = instructionNo
			summary.instructionCount += count
		}
		settledPayoff := decimal.Zero
		if itm {
			settledPayoff = optionSettlementPayoff(contract, deliveryPrice, qty)
		}
		applyPositionSettlementReturn(
			position, settledPayoff, decimal.Zero, qty, optionMultiplier(contract),
		)
		position.PositionQty = decimal.Zero
		position.AvailableQty = decimal.Zero
		position.FrozenQty = decimal.Zero
		position.PositionValue = decimal.Zero
		position.MarginAmount = decimal.Zero
		position.MaintenanceMargin = decimal.Zero
		position.UnrealizedPnl = decimal.Zero
		position.ExerciseableQty = decimal.Zero
		position.Status = int64(option.PositionStatus_POSITION_STATUS_SETTLED)
		position.LastCalcTime = now
		position.UpdateTimes = now
		if err := positionModel.Update(ctx, position); err != nil {
			return optionSettlementSummary{}, err
		}
		auditDeliveryQty, auditPaymentAmount := deliveryQty, paymentAmount
		if !itm {
			auditDeliveryQty = decimal.Zero
			auditPaymentAmount = decimal.Zero
		}
		if _, err := detailModel.Insert(ctx, &models.TOptionSettlementDetail{
			TenantId: position.TenantId, BatchId: batchId, BatchNo: settlementNo,
			ContractId: contract.Id, PositionId: position.Id, UserId: position.UserId,
			AccountId: position.AccountId, Side: position.Side, Quantity: qty,
			Payoff: settledPayoff, Direction: int64(direction),
			InstructionNo: firstInstruction[position.Id],
			DeliveryCoin:  contract.UnderlyingCoin, DeliveryQuantity: auditDeliveryQty,
			PaymentCoin: contract.SettleCoin, PaymentAmount: auditPaymentAmount,
			CreateTimes: now,
		}); err != nil {
			return optionSettlementSummary{}, err
		}
	}
	return summary, validateOptionSettlementBalance(contract.Id, summary)
}

func findPhysicalSettlementPositions(
	ctx context.Context,
	positionModel models.TOptionPositionModel,
	contractId int64,
) ([]*models.TOptionPosition, error) {
	var result []*models.TOptionPosition
	cursor := int64(0)
	for {
		items, _, err := positionModel.FindPage(ctx, models.OptionPositionPageFilter{
			ContractId: contractId,
			Statuses: []int64{
				int64(option.PositionStatus_POSITION_STATUS_HOLDING),
				int64(option.PositionStatus_POSITION_STATUS_EXPIRED),
			},
		}, cursor, 100)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
		if len(items) < 100 {
			sort.SliceStable(result, func(i, j int) bool {
				if result[i].CreateTimes == result[j].CreateTimes {
					return result[i].Id < result[j].Id
				}
				return result[i].CreateTimes < result[j].CreateTimes
			})
			return result, nil
		}
		cursor = items[len(items)-1].Id
	}
}

type physicalDeliveryAllocation struct {
	long     *models.TOptionPosition
	short    *models.TOptionPosition
	quantity decimal.Decimal
}

func allocatePhysicalDeliveryPositions(
	positions []*models.TOptionPosition,
) ([]physicalDeliveryAllocation, error) {
	var longs, shorts []*models.TOptionPosition
	longTotal, shortTotal := decimal.Zero, decimal.Zero
	for _, position := range positions {
		if position == nil || !position.PositionQty.IsPositive() {
			continue
		}
		switch position.Side {
		case int64(common.PositionSide_POSITION_SIDE_LONG):
			longs = append(longs, position)
			longTotal = longTotal.Add(position.PositionQty)
		case int64(common.PositionSide_POSITION_SIDE_SHORT):
			shorts = append(shorts, position)
			shortTotal = shortTotal.Add(position.PositionQty)
		default:
			return nil, fmt.Errorf("unsupported physical position side: positionId=%d", position.Id)
		}
	}
	if !longTotal.Equal(shortTotal) {
		return nil, fmt.Errorf("unbalanced quantity: long=%s short=%s", longTotal, shortTotal)
	}
	if longTotal.IsZero() {
		return nil, nil
	}
	result := make([]physicalDeliveryAllocation, 0, len(longs)+len(shorts))
	longIndex, shortIndex := 0, 0
	longRemaining, shortRemaining := longs[0].PositionQty, shorts[0].PositionQty
	for longIndex < len(longs) && shortIndex < len(shorts) {
		quantity := decimal.Min(longRemaining, shortRemaining)
		if !quantity.IsPositive() {
			return nil, errors.New("physical allocation produced non-positive quantity")
		}
		result = append(result, physicalDeliveryAllocation{
			long: longs[longIndex], short: shorts[shortIndex], quantity: quantity,
		})
		longRemaining = longRemaining.Sub(quantity)
		shortRemaining = shortRemaining.Sub(quantity)
		if longRemaining.IsZero() {
			longIndex++
			if longIndex < len(longs) {
				longRemaining = longs[longIndex].PositionQty
			}
		}
		if shortRemaining.IsZero() {
			shortIndex++
			if shortIndex < len(shorts) {
				shortRemaining = shorts[shortIndex].PositionQty
			}
		}
	}
	if longIndex != len(longs) || shortIndex != len(shorts) {
		return nil, errors.New("physical allocation did not consume both sides")
	}
	return result, nil
}

func createPhysicalDeliveryUnit(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	unitModel models.TOptionPhysicalDeliveryUnitModel,
	contract *models.TOptionContract,
	allocation physicalDeliveryAllocation,
	batchId int64,
	settlementNo string,
	sequence int64,
	now int64,
) (string, string, int64, error) {
	deliveryQty := allocation.quantity.Mul(optionMultiplier(contract)).Round(16)
	paymentAmount := optionExerciseAmount(contract, allocation.quantity)
	unitNo := fmt.Sprintf("%s-DU%06d", settlementNo, sequence)
	result, err := unitModel.Insert(ctx, &models.TOptionPhysicalDeliveryUnit{
		TenantId: allocation.long.TenantId, DeliveryUnitNo: unitNo,
		BatchId: batchId, BatchNo: settlementNo, ContractId: contract.Id,
		LongPositionId: allocation.long.Id, LongUserId: allocation.long.UserId,
		LongAccountId: allocation.long.AccountId, ShortPositionId: allocation.short.Id,
		ShortUserId: allocation.short.UserId, ShortAccountId: allocation.short.AccountId,
		Quantity: allocation.quantity, DeliveryCoin: contract.UnderlyingCoin,
		DeliveryQuantity: deliveryQty, PaymentCoin: contract.SettleCoin,
		PaymentAmount: paymentAmount,
		Status:        int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_ASSET_PROCESSING),
		CureDeadline:  now + contract.PhysicalDeliveryCureSeconds,
		CreateTimes:   now, UpdateTimes: now,
	})
	if err != nil {
		return "", "", 0, err
	}
	unitId, err := result.LastInsertId()
	if err != nil {
		return "", "", 0, err
	}
	longDebitCoin, longDebitAmount, longCreditCoin, longCreditAmount :=
		physicalLongAssetLegs(contract, deliveryQty, paymentAmount)
	longDebitNo := unitNo + "-LONG-DEBIT"
	if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
		TenantId: allocation.long.TenantId, InstructionNo: longDebitNo,
		BizNo: settlementNo, PositionId: allocation.long.Id, DeliveryUnitId: unitId,
		ExecutionGroup: unitNo, UserId: allocation.long.UserId, AccountId: allocation.long.AccountId,
		Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
		Coin:   longDebitCoin, Amount: longDebitAmount, StepNo: 1,
		Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	}); err != nil {
		return "", "", 0, err
	}
	shortDebitNo, shortDebitCount, err := createPhysicalUnitShortDebits(
		ctx, instructionModel, marginLotModel, contract, allocation.short,
		unitId, unitNo, settlementNo, allocation.quantity, longCreditAmount, longCreditCoin, now,
	)
	if err != nil {
		return "", "", 0, err
	}
	longCreditNo := unitNo + "-LONG-CREDIT"
	if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
		TenantId: allocation.long.TenantId, InstructionNo: longCreditNo,
		BizNo: settlementNo, PositionId: allocation.long.Id, DeliveryUnitId: unitId,
		ExecutionGroup: unitNo, UserId: allocation.long.UserId, AccountId: allocation.long.AccountId,
		Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
		Coin:   longCreditCoin, Amount: longCreditAmount, StepNo: 3,
		Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	}); err != nil {
		return "", "", 0, err
	}
	shortCreditNo := unitNo + "-SHORT-CREDIT"
	if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
		TenantId: allocation.short.TenantId, InstructionNo: shortCreditNo,
		BizNo: settlementNo, PositionId: allocation.short.Id, DeliveryUnitId: unitId,
		ExecutionGroup: unitNo, UserId: allocation.short.UserId, AccountId: allocation.short.AccountId,
		Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
		Coin:   longDebitCoin, Amount: longDebitAmount, StepNo: 3,
		Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	}); err != nil {
		return "", "", 0, err
	}
	return longDebitNo, shortDebitNo, shortDebitCount + 3, nil
}

func createPhysicalUnitShortDebits(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	_ *models.TOptionContract,
	short *models.TOptionPosition,
	unitId int64,
	unitNo, settlementNo string,
	quantity, expectedAmount decimal.Decimal,
	expectedCoin string,
	now int64,
) (string, int64, error) {
	lots, err := marginLotModel.FindClosableByPosition(ctx, short.TenantId, short.Id)
	if err != nil {
		return "", 0, err
	}
	remainingQty, allocatedAmount := quantity, decimal.Zero
	firstInstruction := ""
	count := int64(0)
	for _, lot := range lots {
		if !remainingQty.IsPositive() {
			break
		}
		availableQty := lot.RemainingQuantity
		availableAmount := decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero)
		if !availableQty.IsPositive() || !availableAmount.IsPositive() {
			continue
		}
		coin := lot.CollateralCoin
		if coin == "" {
			return "", 0, fmt.Errorf("physical delivery collateral coin evidence is missing: lotId=%d", lot.Id)
		}
		if coin != expectedCoin {
			return "", 0, fmt.Errorf("physical delivery collateral coin mismatch: lotId=%d", lot.Id)
		}
		takeQty := decimal.Min(availableQty, remainingQty)
		takeAmount := takeQty.Mul(expectedAmount).Div(quantity).Round(16)
		if takeQty.Equal(availableQty) {
			takeAmount = availableAmount
		}
		if takeAmount.GreaterThan(availableAmount) || !takeAmount.IsPositive() {
			return "", 0, fmt.Errorf("physical delivery collateral amount mismatch: lotId=%d", lot.Id)
		}
		instructionNo := fmt.Sprintf("%s-SHORT-L%d-DEBIT", unitNo, lot.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: short.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: short.Id, MarginLotId: lot.Id,
			DeliveryUnitId: unitId, ExecutionGroup: unitNo,
			UserId: short.UserId, AccountId: short.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			TargetBizNo: lot.FreezeBizNo, Coin: coin, Amount: takeAmount, StepNo: 2,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstruction == "" {
			firstInstruction = instructionNo
		}
		count++
		remainingQty = remainingQty.Sub(takeQty)
		allocatedAmount = allocatedAmount.Add(takeAmount)
		lot.PendingMargin = lot.PendingMargin.Add(takeAmount)
		lot.RemainingQuantity = decimal.Max(lot.RemainingQuantity.Sub(takeQty), decimal.Zero)
		lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", 0, err
		}
	}
	if remainingQty.IsPositive() || !allocatedAmount.Equal(expectedAmount) {
		return "", 0, fmt.Errorf(
			"physical delivery is not fully collateralized: positionId=%d quantityMissing=%s expected=%s actual=%s",
			short.Id, remainingQty, expectedAmount, allocatedAmount,
		)
	}
	return firstInstruction, count, nil
}

func physicalLongAssetLegs(
	contract *models.TOptionContract,
	deliveryQty, paymentAmount decimal.Decimal,
) (debitCoin string, debitAmount decimal.Decimal, creditCoin string, creditAmount decimal.Decimal) {
	debitCoin, debitAmount = contract.SettleCoin, paymentAmount
	creditCoin, creditAmount = contract.UnderlyingCoin, deliveryQty
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
		debitCoin, debitAmount = contract.UnderlyingCoin, deliveryQty
		creditCoin, creditAmount = contract.SettleCoin, paymentAmount
	}
	return
}

func createCoveredShortSettlementInstructions(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	contract *models.TOptionContract,
	position *models.TOptionPosition,
	settlementNo string,
	deliveryQty, paymentAmount decimal.Decimal,
	itm bool,
	now int64,
) (string, int64, error) {
	lots, err := marginLotModel.FindClosableByPosition(ctx, position.TenantId, position.Id)
	if err != nil {
		return "", 0, err
	}
	expectedCoin, expectedAmount := contract.UnderlyingCoin, deliveryQty
	creditCoin, creditAmount := contract.SettleCoin, paymentAmount
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_PUT) {
		expectedCoin, expectedAmount = contract.SettleCoin, paymentAmount
		creditCoin, creditAmount = contract.UnderlyingCoin, deliveryQty
	}
	firstInstruction := ""
	count := int64(0)
	totalCollateral := decimal.Zero
	for _, lot := range lots {
		if lot.PendingMargin.IsPositive() {
			return "", 0, errors.New("physical delivery margin lot has pending amount")
		}
		coin := lot.CollateralCoin
		if coin == "" {
			return "", 0, fmt.Errorf("physical delivery collateral coin evidence is missing: lotId=%d", lot.Id)
		}
		if coin != expectedCoin {
			return "", 0, fmt.Errorf("physical delivery collateral coin mismatch: lotId=%d", lot.Id)
		}
		amount := lot.RemainingMargin
		totalCollateral = totalCollateral.Add(amount)
		if !amount.IsPositive() {
			continue
		}
		action := option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN
		suffix := "RELEASE"
		if itm {
			action = option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN
			suffix = "DEDUCT"
		}
		instructionNo := fmt.Sprintf("%s-P%d-L%d-%s", settlementNo, position.Id, lot.Id, suffix)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
			UserId: position.UserId, AccountId: position.AccountId,
			Action: int64(action), TargetBizNo: lot.FreezeBizNo,
			Coin: coin, Amount: amount, StepNo: 1,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstruction == "" {
			firstInstruction = instructionNo
		}
		count++
		lot.PendingMargin = lot.PendingMargin.Add(amount)
		lot.RemainingQuantity = decimal.Zero
		if itm {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		} else {
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
		}
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", 0, err
		}
	}
	if itm && !totalCollateral.Equal(expectedAmount) {
		return "", 0, fmt.Errorf(
			"physical delivery is not fully collateralized: positionId=%d expected=%s actual=%s",
			position.Id, expectedAmount, totalCollateral,
		)
	}
	if itm {
		creditNo := fmt.Sprintf("%s-P%d-PHYSICAL-CREDIT", settlementNo, position.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: creditNo,
			BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE),
			Coin:   creditCoin, Amount: creditAmount, StepNo: 2,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstruction == "" {
			firstInstruction = creditNo
		}
		count++
	}
	return firstInstruction, count, nil
}

func createShortSettlementInstructions(ctx context.Context, instructionModel models.TOptionAssetInstructionModel, marginLotModel models.TOptionMarginLotModel, contract *models.TOptionContract, position *models.TOptionPosition, settlementNo string, payoff decimal.Decimal, now int64) (string, int64, error) {
	if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
		return createPortfolioShortSettlementInstructions(
			ctx, instructionModel, marginLotModel, contract, position, settlementNo, payoff, now,
		)
	}
	lots, err := marginLotModel.FindActiveByPosition(ctx, position.TenantId, position.Id)
	if err != nil {
		return "", 0, err
	}
	remainingPayoff := payoff
	firstInstructionNo := ""
	count := int64(0)
	for _, lot := range lots {
		availableMargin := decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero)
		deduct := decimal.Min(availableMargin, remainingPayoff)
		if deduct.IsPositive() {
			instructionNo := fmt.Sprintf("%s-P%d-L%d-MARGIN-DEDUCT", settlementNo, position.Id, lot.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: instructionNo,
				BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
				UserId: position.UserId, AccountId: position.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: deduct, StepNo: 1,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return "", 0, err
			}
			if firstInstructionNo == "" {
				firstInstructionNo = instructionNo
			}
			count++
			remainingPayoff = remainingPayoff.Sub(deduct)
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		}
		release := availableMargin.Sub(deduct)
		if release.IsPositive() {
			instructionNo := fmt.Sprintf("%s-P%d-L%d-MARGIN-RELEASE", settlementNo, position.Id, lot.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: position.TenantId, InstructionNo: instructionNo,
				BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
				UserId: position.UserId, AccountId: position.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: release, StepNo: 2,
				Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return "", 0, err
			}
			if firstInstructionNo == "" {
				firstInstructionNo = instructionNo
			}
			count++
			if !deduct.IsPositive() {
				lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
			}
		}
		lot.PendingMargin = lot.PendingMargin.Add(deduct).Add(release)
		lot.RemainingQuantity = decimal.Zero
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", 0, err
		}
	}
	if remainingPayoff.IsPositive() {
		instructionNo := fmt.Sprintf("%s-P%d-DEBIT-AVAILABLE", settlementNo, position.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: remainingPayoff, StepNo: 1,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstructionNo == "" {
			firstInstructionNo = instructionNo
		}
		count++
	}
	return firstInstructionNo, count, nil
}

func createPortfolioShortSettlementInstructions(
	ctx context.Context,
	instructionModel models.TOptionAssetInstructionModel,
	marginLotModel models.TOptionMarginLotModel,
	contract *models.TOptionContract,
	position *models.TOptionPosition,
	settlementNo string,
	payoff decimal.Decimal,
	now int64,
) (string, int64, error) {
	lots, err := marginLotModel.FindPortfolioActiveByAccount(
		ctx, position.TenantId, position.UserId, position.AccountId, contract.SettleCoin,
	)
	if err != nil {
		return "", 0, err
	}
	remainingPayoff := payoff
	firstInstructionNo := ""
	count := int64(0)
	for _, lot := range lots {
		if !remainingPayoff.IsPositive() {
			break
		}
		if lot.PendingMargin.IsPositive() {
			return "", 0, errors.New("portfolio settlement margin lot has pending amount")
		}
		deduct := decimal.Min(lot.RemainingMargin, remainingPayoff)
		if !deduct.IsPositive() {
			continue
		}
		instructionNo := fmt.Sprintf("%s-P%d-PL%d-MARGIN-DEDUCT", settlementNo, position.Id, lot.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: position.Id, MarginLotId: lot.Id,
			UserId: position.UserId, AccountId: position.AccountId,
			Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN),
			TargetBizNo: lot.FreezeBizNo, Coin: contract.SettleCoin, Amount: deduct, StepNo: 1,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstructionNo == "" {
			firstInstructionNo = instructionNo
		}
		count++
		remainingPayoff = remainingPayoff.Sub(deduct)
		lot.PendingMargin = lot.PendingMargin.Add(deduct)
		lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_CONSUMING)
		lot.UpdateTimes = now
		if err := marginLotModel.Update(ctx, lot); err != nil {
			return "", 0, err
		}
	}
	if remainingPayoff.IsPositive() {
		instructionNo := fmt.Sprintf("%s-P%d-DEBIT-AVAILABLE", settlementNo, position.Id)
		if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
			TenantId: position.TenantId, InstructionNo: instructionNo,
			BizNo: settlementNo, PositionId: position.Id, UserId: position.UserId, AccountId: position.AccountId,
			Action: int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE),
			Coin:   contract.SettleCoin, Amount: remainingPayoff, StepNo: 1,
			Status:               int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
			ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
			CreateTimes:          now, UpdateTimes: now,
		}); err != nil {
			return "", 0, err
		}
		if firstInstructionNo == "" {
			firstInstructionNo = instructionNo
		}
		count++
	}
	return firstInstructionNo, count, nil
}

func validateOptionSettlementBalance(contractId int64, summary optionSettlementSummary) error {
	if !summary.totalCredit.Equal(summary.totalDebit) {
		return fmt.Errorf("option settlement is not balanced: contractId=%d credit=%s debit=%s", contractId, summary.totalCredit, summary.totalDebit)
	}
	return nil
}

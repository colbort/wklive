package tasklogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/observability"
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ProcessRiskAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessRiskAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessRiskAccountsLogic {
	return &ProcessRiskAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 聚合卖方持仓、资产权益与行情，刷新风险账户
func (l *ProcessRiskAccountsLogic) ProcessRiskAccounts(in *option.OptionTaskReq) (*option.OptionTaskResp, error) {
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "process_risk_accounts", func() (*option.OptionTaskResp, error) {
		groups, err := l.collectRiskGroups(in.TenantId)
		if err != nil {
			observability.RecordRiskScanExecutionFailure(in.TenantId, "collect")
			return nil, err
		}
		var groupErrors []error
		results := make(map[int64]*observability.RiskScanTenantResult)
		if in.TenantId > 0 {
			// Each invocation is tenant scoped. Publish an explicit zero result when
			// this tenant has no active wallet groups without clearing other tenants.
			results[in.TenantId] = &observability.RiskScanTenantResult{TenantID: in.TenantId}
		}
		for _, group := range groups {
			result := results[group.key.tenantID]
			if result == nil {
				result = &observability.RiskScanTenantResult{TenantID: group.key.tenantID}
				results[group.key.tenantID] = result
			}
			result.TotalGroups++
			if group.scanErr != nil {
				result.FailedGroups++
				restrictErr := l.restrictRiskGroup(group)
				groupErr := fmt.Errorf(
					"restrict option risk account after market validation failure, tenantId=%d userId=%d coin=%s: %w",
					group.key.tenantID, group.key.userID, group.key.coin, group.scanErr,
				)
				if restrictErr != nil {
					groupErr = errors.Join(groupErr, fmt.Errorf("persist restricted status: %w", restrictErr))
				}
				l.Error(groupErr)
				groupErrors = append(groupErrors, groupErr)
				continue
			}
			if refreshErr := l.refreshRiskGroup(group); refreshErr != nil {
				result.FailedGroups++
				restrictErr := l.restrictRiskGroup(group)
				groupErr := fmt.Errorf(
					"refresh option risk account failed, tenantId=%d userId=%d coin=%s: %w",
					group.key.tenantID, group.key.userID, group.key.coin, refreshErr,
				)
				if restrictErr != nil {
					groupErr = errors.Join(groupErr, fmt.Errorf("persist restricted status: %w", restrictErr))
				}
				l.Error(groupErr)
				groupErrors = append(groupErrors, groupErr)
			}
		}
		if err := l.resetInactiveRiskAccounts(in.TenantId, groups); err != nil {
			observability.RecordRiskScanExecutionFailure(in.TenantId, "reset_inactive")
			groupErrors = append(groupErrors, fmt.Errorf("reset inactive option risk accounts: %w", err))
			// A scan that failed after wallet evaluation must not advance the
			// completed timestamp or replace the last complete denominator.
			return nil, errors.Join(groupErrors...)
		}
		published := make([]observability.RiskScanTenantResult, 0, len(results))
		for _, result := range results {
			published = append(published, *result)
		}
		observability.PublishRiskScanResults(published, time.Now().Unix(), in.TenantId)
		if len(groupErrors) > 0 {
			return nil, errors.Join(groupErrors...)
		}
		return helpers.OkTaskResp(), nil
	})
}

type optionRiskKey struct {
	tenantID int64
	userID   int64
	// accountID is always zero because Asset stores one OPTION wallet per
	// tenant/user/coin. Business account IDs remain on orders and positions,
	// but they must not split or duplicate wallet equity.
	accountID int64
	coin      string
}

type optionRiskPosition struct {
	position *models.TOptionPosition
	contract *models.TOptionContract
	market   *models.TOptionMarket
}

type optionRiskGroup struct {
	key       optionRiskKey
	positions []optionRiskPosition
	// A wallet may have thousands of positions in the same migrating contract.
	// Record one fail-closed reason per contract instead of constructing an
	// unbounded repeated error chain.
	migrationContracts map[int64]struct{}
	// scanErr prevents this whole wallet risk group from being calculated from
	// a partial or stale market set. Other wallets can still be refreshed.
	scanErr error
}

type portfolioRiskSnapshot struct {
	configItem  *models.TOptionPortfolioRiskConfig
	config      optionrisk.PortfolioConfig
	legs        map[int64]optionrisk.PortfolioLeg
	initial     optionrisk.PortfolioResult
	maintenance optionrisk.PortfolioResult
}

func (l *ProcessRiskAccountsLogic) collectRiskGroups(tenantID int64) ([]*optionRiskGroup, error) {
	groupMap := make(map[optionRiskKey]*optionRiskGroup)
	migrationCache := make(map[[2]int64]bool)
	migrationChecked := make(map[[2]int64]struct{})
	cursor := int64(0)
	now := time.Now().Unix()
	for {
		positions, _, err := l.svcCtx.OptionPositionModel.FindPage(l.ctx, models.OptionPositionPageFilter{
			TenantId: tenantID,
			Status:   int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		}, cursor, 100)
		if err != nil {
			return nil, err
		}
		for _, position := range positions {
			cursor = position.Id
			contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, position.ContractId)
			if err != nil {
				return nil, err
			}
			// Cash-settled long positions remain part of wallet net liquidation
			// value even when sell-to-open is disabled for their contract.
			// Physical delivery has different asset legs and is deliberately not
			// admitted as collateral to a cash-settled seller risk pool.
			if contract.SettlementType != int64(option.SettlementType_SETTLEMENT_TYPE_CASH) {
				continue
			}
			key := optionRiskKey{
				tenantID: position.TenantId, userID: position.UserId,
				accountID: 0, coin: contract.SettleCoin,
			}
			group := groupMap[key]
			if group == nil {
				group = &optionRiskGroup{key: key, migrationContracts: make(map[int64]struct{})}
				groupMap[key] = group
			}
			migrationKey := [2]int64{position.TenantId, position.ContractId}
			migrationActive := migrationCache[migrationKey]
			if _, checked := migrationChecked[migrationKey]; !checked {
				migrationActive, err = l.svcCtx.OptionCorporateActionContractModel.IsContractMigrationActive(
					l.ctx, position.TenantId, position.ContractId,
				)
				if err != nil {
					return nil, err
				}
				migrationCache[migrationKey] = migrationActive
				migrationChecked[migrationKey] = struct{}{}
			}
			if migrationActive {
				if _, recorded := group.migrationContracts[contract.Id]; !recorded {
					group.scanErr = errors.Join(
						group.scanErr,
						fmt.Errorf("corporate action migration active, contractId=%d", contract.Id),
					)
					group.migrationContracts[contract.Id] = struct{}{}
				}
				continue
			}
			market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(
				l.ctx, position.TenantId, position.ContractId,
			)
			if err != nil {
				group.scanErr = errors.Join(
					group.scanErr,
					fmt.Errorf("load option risk market, contractId=%d: %w", contract.Id, err),
				)
				continue
			}
			group.positions = append(group.positions, optionRiskPosition{
				position: position, contract: contract, market: market,
			})
			if !helpers.IsRiskMarketFresh(market, now, 30) {
				group.scanErr = errors.Join(
					group.scanErr,
					fmt.Errorf(
						"stale option risk market, contractId=%d underlyingSnapshotTime=%d markSnapshotTime=%d",
						contract.Id, market.UnderlyingSnapshotTime, market.MarkSnapshotTime,
					),
				)
			}
		}
		if len(positions) < 100 {
			break
		}
	}
	groups := make([]*optionRiskGroup, 0, len(groupMap))
	for _, group := range groupMap {
		groups = append(groups, group)
	}
	return groups, nil
}

func (l *ProcessRiskAccountsLogic) restrictRiskGroup(group *optionRiskGroup) error {
	if group == nil {
		return nil
	}
	now := time.Now().Unix()
	current, err := l.svcCtx.OptionRiskAccountModel.FindOneByTenantIdUserIdAccountIdSettleCoin(
		l.ctx, group.key.tenantID, group.key.userID, group.key.accountID, group.key.coin,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if errors.Is(err, models.ErrNotFound) {
		_, err = l.svcCtx.OptionRiskAccountModel.Insert(l.ctx, &models.TOptionRiskAccount{
			TenantId: group.key.tenantID, UserId: group.key.userID, AccountId: group.key.accountID,
			SettleCoin:  group.key.coin,
			Status:      int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED),
			CreateTimes: now, UpdateTimes: now,
		})
		return err
	}
	current.Status = int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED)
	// LastCalcTime deliberately remains the time of the last successful risk
	// calculation; UpdateTimes records when the restriction was applied.
	current.UpdateTimes = now
	return l.svcCtx.OptionRiskAccountModel.Update(l.ctx, current)
}

func (l *ProcessRiskAccountsLogic) refreshRiskGroup(group *optionRiskGroup) error {
	resp, err := l.svcCtx.AssetClient.GetAssetBalance(l.ctx, &asset.GetUserAssetDetailReq{
		TenantId:   group.key.tenantID,
		UserId:     group.key.userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION,
		Coin:       group.key.coin,
	})
	if err != nil {
		return err
	}
	if resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return fmt.Errorf("get option asset balance rejected, tenantId=%d userId=%d coin=%s",
			group.key.tenantID, group.key.userID, group.key.coin)
	}
	totalAsset, err := decimal.NewFromString(resp.GetData().GetTotalAmount())
	if err != nil {
		return fmt.Errorf("invalid option asset total amount: %w", err)
	}
	now := time.Now().Unix()
	positionMargin := decimal.Zero
	maintenanceMargin := decimal.Zero
	unrealizedPnL := decimal.Zero
	netOptionValue := decimal.Zero
	portfolioLegs := make(map[int64]optionrisk.PortfolioLeg)
	for _, item := range group.positions {
		position := item.position
		position.MarkPrice = item.market.MarkPrice
		position.PositionValue = item.market.MarkPrice.Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
		if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			position.UnrealizedPnl = item.market.MarkPrice.Sub(position.OpenAvgPrice).
				Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
			netOptionValue = netOptionValue.Add(position.PositionValue)
		} else {
			position.UnrealizedPnl = position.OpenAvgPrice.Sub(item.market.MarkPrice).
				Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
			netOptionValue = netOptionValue.Sub(position.PositionValue)
		}
		switch option.SellerMarginMode(item.contract.SellerMarginMode) {
		case option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED:
			if position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) {
				position.MaintenanceMargin = optionTaskSellerMargin(
					item.contract, item.market.UnderlyingPrice, item.market.MarkPrice, position.PositionQty,
				)
				positionMargin = positionMargin.Add(position.MarginAmount)
				maintenanceMargin = maintenanceMargin.Add(position.MaintenanceMargin)
			} else {
				position.MaintenanceMargin = decimal.Zero
			}
		case option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO:
			position.MaintenanceMargin = decimal.Zero
			leg := portfolioLegs[item.contract.Id]
			leg.Contract = item.contract
			leg.Market = item.market
			if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
				leg.LongQuantity = leg.LongQuantity.Add(position.PositionQty)
			} else {
				leg.ShortQuantity = leg.ShortQuantity.Add(position.PositionQty)
			}
			portfolioLegs[item.contract.Id] = leg
		default:
			position.MaintenanceMargin = decimal.Zero
		}
		position.LastCalcTime = now
		position.UpdateTimes = now
		if err := l.svcCtx.OptionPositionModel.Update(l.ctx, position); err != nil {
			return err
		}
		unrealizedPnL = unrealizedPnL.Add(position.UnrealizedPnl)
	}
	portfolioMethod := option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_UNKNOWN
	portfolioScenarioLoss := decimal.Zero
	portfolioShortFloor := decimal.Zero
	portfolioConcentrationAddon := decimal.Zero
	portfolioLiquidityAddon := decimal.Zero
	portfolioRiskConfigID := int64(0)
	portfolioRiskConfigVersion := int64(0)
	portfolioInitialRequirement := decimal.Zero
	var portfolioSnapshot *portfolioRiskSnapshot
	if len(portfolioLegs) > 0 {
		configItem, err := l.svcCtx.OptionPortfolioRiskConfigModel.FindActive(
			l.ctx, group.key.tenantID, group.key.coin, now,
		)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return errors.New("no approved active portfolio risk config")
			}
			return err
		}
		config, err := optionrisk.PortfolioConfigFromModel(configItem)
		if err != nil {
			return fmt.Errorf("invalid active portfolio risk config %d: %w", configItem.Id, err)
		}
		legs := make([]optionrisk.PortfolioLeg, 0, len(portfolioLegs))
		for _, leg := range portfolioLegs {
			legs = append(legs, leg)
		}
		initialResult, err := optionrisk.EvaluatePortfolio(legs, false, config)
		if err != nil {
			return err
		}
		maintenanceResult, err := optionrisk.EvaluatePortfolio(legs, true, config)
		if err != nil {
			return err
		}
		positionMargin = positionMargin.Add(initialResult.Requirement)
		portfolioInitialRequirement = initialResult.Requirement
		maintenanceMargin = maintenanceMargin.Add(maintenanceResult.Requirement)
		portfolioMethod = option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1
		portfolioRiskConfigID = configItem.Id
		portfolioRiskConfigVersion = configItem.Version
		portfolioScenarioLoss = maintenanceResult.ScenarioLoss
		portfolioShortFloor = maintenanceResult.ShortFloor
		portfolioConcentrationAddon = maintenanceResult.ConcentrationAddon
		portfolioLiquidityAddon = maintenanceResult.LiquidityAddon
		portfolioSnapshot = &portfolioRiskSnapshot{
			configItem: configItem, config: config, legs: portfolioLegs,
			initial: initialResult, maintenance: maintenanceResult,
		}
	}
	equity := optionRiskEquity(totalAsset, netOptionValue)
	riskRate := decimal.Zero
	if equity.IsPositive() {
		riskRate = maintenanceMargin.Div(equity).Round(10)
	} else if maintenanceMargin.IsPositive() {
		riskRate = decimal.NewFromInt(999999)
	}
	status := option.RiskAccountStatus_RISK_ACCOUNT_STATUS_NORMAL
	switch {
	case !equity.IsPositive():
		status = option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT
	case equity.LessThanOrEqual(maintenanceMargin):
		status = option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING
	case equity.LessThanOrEqual(maintenanceMargin.Mul(decimal.NewFromFloat(1.1))):
		status = option.RiskAccountStatus_RISK_ACCOUNT_STATUS_MARGIN_CALL
	}
	current, err := l.svcCtx.OptionRiskAccountModel.FindOneByTenantIdUserIdAccountIdSettleCoin(
		l.ctx, group.key.tenantID, group.key.userID, group.key.accountID, group.key.coin,
	)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return err
	}
	if errors.Is(err, models.ErrNotFound) {
		_, err = l.svcCtx.OptionRiskAccountModel.Insert(l.ctx, &models.TOptionRiskAccount{
			TenantId: group.key.tenantID, UserId: group.key.userID, AccountId: group.key.accountID,
			SettleCoin: group.key.coin, Equity: equity, PositionMargin: positionMargin,
			NetOptionValue:    netOptionValue,
			MaintenanceMargin: maintenanceMargin, UnrealizedPnl: unrealizedPnL, RiskRate: riskRate,
			PortfolioRiskMethod:   int64(portfolioMethod),
			PortfolioRiskConfigId: portfolioRiskConfigID, PortfolioRiskConfigVersion: portfolioRiskConfigVersion,
			PortfolioScenarioLoss: portfolioScenarioLoss, PortfolioShortFloor: portfolioShortFloor,
			PortfolioConcentrationAddon: portfolioConcentrationAddon,
			PortfolioLiquidityAddon:     portfolioLiquidityAddon,
			Status:                      int64(status), LastCalcTime: now, CreateTimes: now, UpdateTimes: now,
		})
		if err != nil {
			return err
		}
		if status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING &&
			status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT {
			if err := l.rebalancePortfolioCollateral(group, portfolioInitialRequirement, now); err != nil {
				return err
			}
		}
		return l.ensureLiquidations(group, status, equity, maintenanceMargin, now, portfolioSnapshot)
	}
	current.Equity = equity
	current.NetOptionValue = netOptionValue
	current.PositionMargin = positionMargin
	current.MaintenanceMargin = maintenanceMargin
	current.UnrealizedPnl = unrealizedPnL
	current.RiskRate = riskRate
	current.PortfolioRiskMethod = int64(portfolioMethod)
	current.PortfolioRiskConfigId = portfolioRiskConfigID
	current.PortfolioRiskConfigVersion = portfolioRiskConfigVersion
	current.PortfolioScenarioLoss = portfolioScenarioLoss
	current.PortfolioShortFloor = portfolioShortFloor
	current.PortfolioConcentrationAddon = portfolioConcentrationAddon
	current.PortfolioLiquidityAddon = portfolioLiquidityAddon
	current.Status = int64(status)
	current.LastCalcTime = now
	current.UpdateTimes = now
	if err := l.svcCtx.OptionRiskAccountModel.Update(l.ctx, current); err != nil {
		return err
	}
	if status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING &&
		status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT {
		if err := l.rebalancePortfolioCollateral(group, portfolioInitialRequirement, now); err != nil {
			return err
		}
	}
	return l.ensureLiquidations(group, status, equity, maintenanceMargin, now, portfolioSnapshot)
}

func (l *ProcessRiskAccountsLogic) rebalancePortfolioCollateral(
	group *optionRiskGroup,
	required decimal.Decimal,
	now int64,
) error {
	if group == nil {
		return nil
	}
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		riskAccountModel := models.NewTOptionRiskAccountModel(conn, l.svcCtx.Config.CacheRedis)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)
		marginLotModel := models.NewTOptionMarginLotModel(conn, l.svcCtx.Config.CacheRedis)
		instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
		if _, err := riskAccountModel.EnsureAndFindOneForUpdate(
			ctx, group.key.tenantID, group.key.userID, group.key.accountID, group.key.coin, now,
		); err != nil {
			return err
		}
		orders, err := orderModel.FindPortfolioRiskOrders(
			ctx, group.key.tenantID, group.key.userID, group.key.accountID,
		)
		if err != nil {
			return err
		}
		for _, order := range orders {
			contract, err := contractModel.FindOne(ctx, order.ContractId)
			if err != nil {
				return err
			}
			if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) &&
				contract.SettleCoin == group.key.coin {
				// Do not release pool collateral while a risk-increasing order
				// remains live; its full-fill scenario is evaluated at admission.
				return nil
			}
		}
		lots, err := marginLotModel.FindPortfolioActiveByAccount(
			ctx, group.key.tenantID, group.key.userID, group.key.accountID, group.key.coin,
		)
		if err != nil {
			return err
		}
		available := decimal.Zero
		for _, lot := range lots {
			available = available.Add(decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero))
		}
		excess := decimal.Max(available.Sub(required), decimal.Zero)
		for _, lot := range lots {
			if !excess.IsPositive() {
				break
			}
			lotAvailable := decimal.Max(lot.RemainingMargin.Sub(lot.PendingMargin), decimal.Zero)
			release := decimal.Min(lotAvailable, excess)
			if !release.IsPositive() {
				continue
			}
			instructionNo := fmt.Sprintf("PMR-%d-%d-L%d", group.key.accountID, time.Now().UnixNano(), lot.Id)
			if _, err := instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
				TenantId: group.key.tenantID, InstructionNo: instructionNo,
				BizNo: instructionNo, MarginLotId: lot.Id,
				UserId: lot.UserId, AccountId: lot.AccountId,
				Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_RELEASE_FROZEN),
				TargetBizNo: lot.FreezeBizNo, Coin: group.key.coin, Amount: release,
				StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
				ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
				CreateTimes:          now, UpdateTimes: now,
			}); err != nil {
				return err
			}
			lot.PendingMargin = lot.PendingMargin.Add(release)
			lot.Status = int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RELEASING)
			lot.UpdateTimes = now
			if err := marginLotModel.Update(ctx, lot); err != nil {
				return err
			}
			excess = excess.Sub(release)
		}
		return nil
	})
}

func (l *ProcessRiskAccountsLogic) resetInactiveRiskAccounts(tenantID int64, groups []*optionRiskGroup) error {
	active := make(map[optionRiskKey]struct{}, len(groups))
	for _, group := range groups {
		active[group.key] = struct{}{}
	}
	accounts, err := l.svcCtx.OptionRiskAccountModel.FindByTenant(l.ctx, tenantID)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	for _, account := range accounts {
		key := optionRiskKey{
			tenantID: account.TenantId, userID: account.UserId,
			accountID: account.AccountId, coin: account.SettleCoin,
		}
		if _, ok := active[key]; ok {
			continue
		}
		if account.PortfolioRiskMethod != int64(option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_UNKNOWN) {
			if err := l.rebalancePortfolioCollateral(&optionRiskGroup{key: key}, decimal.Zero, now); err != nil {
				return err
			}
		}
		account.PositionMargin = decimal.Zero
		account.NetOptionValue = decimal.Zero
		account.MaintenanceMargin = decimal.Zero
		account.UnrealizedPnl = decimal.Zero
		account.RiskRate = decimal.Zero
		account.PortfolioRiskMethod = int64(option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_UNKNOWN)
		account.PortfolioRiskConfigId = 0
		account.PortfolioRiskConfigVersion = 0
		account.PortfolioScenarioLoss = decimal.Zero
		account.PortfolioShortFloor = decimal.Zero
		account.PortfolioConcentrationAddon = decimal.Zero
		account.PortfolioLiquidityAddon = decimal.Zero
		account.Status = int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_NORMAL)
		account.LastCalcTime = now
		account.UpdateTimes = now
		if err := l.svcCtx.OptionRiskAccountModel.Update(l.ctx, account); err != nil {
			return err
		}
	}
	return nil
}

func (l *ProcessRiskAccountsLogic) ensureLiquidations(
	group *optionRiskGroup,
	status option.RiskAccountStatus,
	equity, maintenanceMargin decimal.Decimal,
	now int64,
	portfolioSnapshot *portfolioRiskSnapshot,
) error {
	if status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING &&
		status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT {
		return nil
	}
	if !hasLiquidatableCustomerShort(group) {
		// Insurance takeover inventory remains visible in the risk account and
		// operational alerts, but must not enter the customer liquidation loop:
		// doing so would transfer it back to the same insurance account forever.
		return nil
	}
	if _, err := l.svcCtx.OptionLiquidationModel.FindOpenByWallet(
		l.ctx, group.key.tenantID, group.key.userID, group.key.coin,
	); err == nil {
		// A wallet is liquidated sequentially. The next quantity is selected only
		// after all position, collateral and Asset effects have converged and the
		// risk account has been recalculated.
		return nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	deficit := decimal.Max(maintenanceMargin.Sub(equity), decimal.Zero).Round(16)
	if portfolioSnapshot != nil {
		// Portfolio quantities are still selected as whole positions because its
		// scenario requirement is non-linear. Keep that path sequential and do
		// not create an isolated liquidation against the same wallet in parallel.
		return l.ensurePortfolioLiquidation(group, equity, maintenanceMargin, now, portfolioSnapshot)
	}
	candidate, err := selectIsolatedLiquidationCandidate(group, deficit)
	if err != nil || candidate == nil {
		return err
	}
	item := candidate.item
	liquidationNo := fmt.Sprintf("OLQ-%d-%d-%d", item.position.TenantId, item.position.Id, now)
	_, err = l.svcCtx.OptionLiquidationModel.Insert(l.ctx, &models.TOptionLiquidation{
		TenantId: item.position.TenantId, LiquidationNo: liquidationNo,
		UserId: item.position.UserId, AccountId: item.position.AccountId,
		ContractId: item.position.ContractId, PositionId: item.position.Id,
		Quantity: candidate.quantity, MarkPrice: item.market.MarkPrice,
		MaintenanceMargin: candidate.maintenance, Equity: equity,
		DeficitAmount: deficit, LiquidationFee: candidate.fee,
		LiquidationScope: int64(option.LiquidationScope_LIQUIDATION_SCOPE_ISOLATED_POSITION),
		Status:           int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution: int64(
			option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE,
		),
		CreateTimes: now, UpdateTimes: now,
	})
	return err
}

type isolatedLiquidationCandidate struct {
	item          *optionRiskPosition
	quantity      decimal.Decimal
	maintenance   decimal.Decimal
	fee           decimal.Decimal
	reliefPerUnit decimal.Decimal
}

// selectIsolatedLiquidationCandidate chooses one deterministic quantity that
// strictly restores equity above maintenance when a single position can do so.
// If it cannot, the whole selected position is taken and the next risk scan
// recalculates the residual wallet before another liquidation is created.
func selectIsolatedLiquidationCandidate(
	group *optionRiskGroup, deficit decimal.Decimal,
) (*isolatedLiquidationCandidate, error) {
	if group == nil {
		return nil, nil
	}
	var selected *isolatedLiquidationCandidate
	hasEligibleShort := false
	for index := range group.positions {
		item := &group.positions[index]
		if isInsuranceInventoryPosition(item) ||
			item.position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) ||
			item.contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) ||
			(item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
				item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)) {
			continue
		}
		hasEligibleShort = true
		quantity := item.position.PositionQty
		step := item.contract.QtyStep
		if !quantity.IsPositive() || !step.IsPositive() || !quantity.Mod(step).IsZero() {
			return nil, fmt.Errorf("invalid isolated liquidation quantity or step for position %d", item.position.Id)
		}
		maintenancePerUnit := item.position.MaintenanceMargin.Div(quantity)
		feePerUnit := item.market.MarkPrice.Mul(optionMultiplier(item.contract)).
			Mul(item.contract.LiquidationFeeRate)
		reliefPerUnit := maintenancePerUnit.Sub(feePerUnit).Round(16)
		if !reliefPerUnit.IsPositive() {
			continue
		}
		if selected == nil || reliefPerUnit.GreaterThan(selected.reliefPerUnit) ||
			(reliefPerUnit.Equal(selected.reliefPerUnit) && item.position.Id < selected.item.position.Id) {
			selected = &isolatedLiquidationCandidate{item: item, reliefPerUnit: reliefPerUnit}
		}
	}
	if selected == nil {
		if hasEligibleShort {
			return nil, errors.New("no isolated short position improves the wallet maintenance gap")
		}
		return nil, nil
	}
	step := selected.item.contract.QtyStep
	reliefPerStep := selected.reliefPerUnit.Mul(step)
	steps := decimal.Max(deficit, decimal.Zero).Div(reliefPerStep).Floor().Add(decimal.NewFromInt(1))
	selected.quantity = decimal.Min(
		steps.Mul(step), selected.item.position.PositionQty,
	).Round(16)
	maintenancePerUnit := selected.item.position.MaintenanceMargin.Div(selected.item.position.PositionQty)
	selected.maintenance = maintenancePerUnit.Mul(selected.quantity).Round(16)
	selected.fee = selected.item.market.MarkPrice.Mul(optionMultiplier(selected.item.contract)).
		Mul(selected.quantity).Mul(selected.item.contract.LiquidationFeeRate).Round(16)
	return selected, nil
}

func hasLiquidatableCustomerShort(group *optionRiskGroup) bool {
	if group == nil {
		return false
	}
	for index := range group.positions {
		item := &group.positions[index]
		if !isInsuranceInventoryPosition(item) &&
			item.position.Side == int64(common.PositionSide_POSITION_SIDE_SHORT) &&
			(item.contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
				item.contract.Status == int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)) {
			return true
		}
	}
	return false
}

func isInsuranceInventoryPosition(item *optionRiskPosition) bool {
	return item != nil && item.position != nil && item.contract != nil &&
		item.contract.InsuranceUserId > 0 && item.contract.InsuranceAccountId > 0 &&
		item.position.UserId == item.contract.InsuranceUserId &&
		item.position.AccountId == item.contract.InsuranceAccountId
}

func (l *ProcessRiskAccountsLogic) ensurePortfolioLiquidation(
	group *optionRiskGroup,
	equity, maintenanceMargin decimal.Decimal,
	now int64,
	snapshot *portfolioRiskSnapshot,
) error {
	if group == nil || snapshot == nil || snapshot.configItem == nil {
		return errors.New("portfolio liquidation risk snapshot is required")
	}
	if _, err := l.svcCtx.OptionLiquidationModel.FindOpenPortfolioByWallet(
		l.ctx, group.key.tenantID, group.key.userID, group.key.coin,
	); err == nil {
		return nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return err
	}
	waiting, err := l.cancelPortfolioRiskOrders(group)
	if err != nil {
		return err
	}
	if waiting {
		return nil
	}
	incomplete, err := l.svcCtx.OptionOutboxModel.HasIncompletePortfolioForWallet(
		l.ctx, group.key.tenantID, group.key.userID, group.key.coin,
	)
	if err != nil {
		return err
	}
	if incomplete {
		return nil
	}
	incomplete, err = l.svcCtx.OptionAssetInstructionModel.HasIncompleteForWallet(
		l.ctx, group.key.tenantID, group.key.userID, group.key.coin,
	)
	if err != nil {
		return err
	}
	if incomplete {
		return nil
	}
	pending, err := l.svcCtx.OptionMarginLotModel.HasPendingPortfolioByWallet(
		l.ctx, group.key.tenantID, group.key.userID, group.key.coin,
	)
	if err != nil {
		return err
	}
	if pending {
		return nil
	}
	candidate, _, _, err := selectPortfolioLiquidationCandidate(group, snapshot)
	if err != nil {
		return err
	}
	lots, err := l.svcCtx.OptionMarginLotModel.FindPortfolioActiveByAccount(
		l.ctx, group.key.tenantID, group.key.userID, 0, group.key.coin,
	)
	if err != nil {
		return err
	}
	collateralBefore := decimal.Zero
	for _, lot := range lots {
		if lot.PendingMargin.IsPositive() {
			return nil
		}
		collateralBefore = collateralBefore.Add(lot.RemainingMargin)
	}
	quantity, initialAfter, maintenanceAfter, err := selectPortfolioLiquidationQuantity(
		candidate, snapshot, equity, maintenanceMargin, collateralBefore,
	)
	if err != nil {
		return err
	}
	takeoverCost := candidate.market.MarkPrice.Mul(quantity).
		Mul(optionMultiplier(candidate.contract)).Round(16)
	fee := takeoverCost.Mul(candidate.contract.LiquidationFeeRate).Round(16)
	consumable := decimal.Max(collateralBefore.Sub(initialAfter.Requirement), decimal.Zero)
	collateralUse := decimal.Min(consumable, takeoverCost.Add(fee))
	collateralAfter := collateralBefore.Sub(collateralUse).Round(16)
	if collateralAfter.LessThan(initialAfter.Requirement) {
		return errors.New("portfolio liquidation would under-collateralize residual portfolio")
	}
	deficit := decimal.Max(maintenanceMargin.Sub(equity), decimal.Zero).Round(16)
	liquidationNo := fmt.Sprintf("OLQ-P-%d-%d-%d", candidate.position.TenantId, candidate.position.Id, now)
	_, err = l.svcCtx.OptionLiquidationModel.Insert(l.ctx, &models.TOptionLiquidation{
		TenantId: candidate.position.TenantId, LiquidationNo: liquidationNo,
		UserId: candidate.position.UserId, AccountId: 0,
		ContractId: candidate.position.ContractId, PositionId: candidate.position.Id,
		Quantity: quantity, MarkPrice: candidate.market.MarkPrice,
		MaintenanceMargin: snapshot.maintenance.Requirement, Equity: equity,
		DeficitAmount: deficit, LiquidationFee: fee,
		Status:                     int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
		DeficitResolution:          int64(option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE),
		LiquidationScope:           int64(option.LiquidationScope_LIQUIDATION_SCOPE_PORTFOLIO_WALLET),
		PortfolioRiskConfigId:      snapshot.configItem.Id,
		PortfolioRiskConfigVersion: snapshot.configItem.Version,
		PortfolioMaintenanceBefore: snapshot.maintenance.Requirement,
		PortfolioMaintenanceAfter:  maintenanceAfter.Requirement,
		PortfolioInitialAfter:      initialAfter.Requirement,
		PortfolioCollateralBefore:  collateralBefore,
		PortfolioCollateralAfter:   collateralAfter,
		CreateTimes:                now, UpdateTimes: now,
	})
	return err
}

const maxPortfolioLiquidationQuantityEvaluations int64 = 100000

func selectPortfolioLiquidationQuantity(
	candidate *optionRiskPosition,
	snapshot *portfolioRiskSnapshot,
	equity, totalMaintenance, collateralBefore decimal.Decimal,
) (decimal.Decimal, optionrisk.PortfolioResult, optionrisk.PortfolioResult, error) {
	if candidate == nil || candidate.position == nil || candidate.contract == nil || snapshot == nil {
		return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("portfolio liquidation candidate and snapshot are required")
	}
	positionQty := candidate.position.PositionQty
	step := candidate.contract.QtyStep
	if !positionQty.IsPositive() || !step.IsPositive() || !positionQty.Mod(step).IsZero() {
		return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			fmt.Errorf("invalid portfolio liquidation quantity or step for position %d", candidate.position.Id)
	}
	stepCount := positionQty.Div(step).IntPart()
	if stepCount <= 0 || stepCount > maxPortfolioLiquidationQuantityEvaluations {
		return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			fmt.Errorf("portfolio liquidation quantity grid for position %d has %d steps; limit is %d",
				candidate.position.Id, stepCount, maxPortfolioLiquidationQuantityEvaluations)
	}
	isolatedMaintenance := totalMaintenance.Sub(snapshot.maintenance.Requirement).Round(16)
	if isolatedMaintenance.IsNegative() {
		return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("portfolio maintenance exceeds total wallet maintenance")
	}
	var fullInitial optionrisk.PortfolioResult
	var fullMaintenance optionrisk.PortfolioResult
	for index := int64(1); index <= stepCount; index++ {
		quantity := step.Mul(decimal.NewFromInt(index)).Round(16)
		initialAfter, maintenanceAfter, err := evaluatePortfolioAfterShortReduction(
			snapshot, candidate, quantity,
		)
		if err != nil {
			return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{}, err
		}
		if index == stepCount {
			fullInitial = initialAfter
			fullMaintenance = maintenanceAfter
		}
		if !snapshot.maintenance.Requirement.GreaterThan(maintenanceAfter.Requirement) ||
			collateralBefore.LessThan(initialAfter.Requirement) {
			continue
		}
		fee := candidate.market.MarkPrice.Mul(quantity).
			Mul(optionMultiplier(candidate.contract)).Mul(candidate.contract.LiquidationFeeRate).Round(16)
		equityAfter := equity.Sub(fee).Round(16)
		maintenanceAfterTotal := isolatedMaintenance.Add(maintenanceAfter.Requirement).Round(16)
		if equityAfter.GreaterThan(maintenanceAfterTotal) {
			return quantity, initialAfter, maintenanceAfter, nil
		}
	}
	if !snapshot.maintenance.Requirement.GreaterThan(fullMaintenance.Requirement) {
		return decimal.Zero, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("selected full portfolio position no longer reduces maintenance requirement")
	}
	return positionQty, fullInitial, fullMaintenance, nil
}

func evaluatePortfolioAfterShortReduction(
	snapshot *portfolioRiskSnapshot,
	candidate *optionRiskPosition,
	quantity decimal.Decimal,
) (optionrisk.PortfolioResult, optionrisk.PortfolioResult, error) {
	if snapshot == nil || candidate == nil || candidate.position == nil ||
		!quantity.IsPositive() || quantity.GreaterThan(candidate.position.PositionQty) {
		return optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("invalid portfolio liquidation reduction quantity")
	}
	legs := make([]optionrisk.PortfolioLeg, 0, len(snapshot.legs))
	matched := false
	for contractID, original := range snapshot.legs {
		leg := original
		if contractID == candidate.contract.Id {
			matched = true
			leg.ShortQuantity = leg.ShortQuantity.Sub(quantity)
			if leg.ShortQuantity.IsNegative() {
				return optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
					errors.New("portfolio liquidation quantity exceeds current short leg")
			}
		}
		legs = append(legs, leg)
	}
	if !matched {
		return optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("portfolio liquidation contract is absent from risk snapshot")
	}
	initialAfter, err := optionrisk.EvaluatePortfolio(legs, false, snapshot.config)
	if err != nil {
		return optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{}, err
	}
	maintenanceAfter, err := optionrisk.EvaluatePortfolio(legs, true, snapshot.config)
	if err != nil {
		return optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{}, err
	}
	return initialAfter, maintenanceAfter, nil
}

func (l *ProcessRiskAccountsLogic) cancelPortfolioRiskOrders(group *optionRiskGroup) (bool, error) {
	orders, err := l.svcCtx.OptionOrderModel.FindPortfolioRiskOrders(
		l.ctx, group.key.tenantID, group.key.userID, 0,
	)
	if err != nil {
		return false, err
	}
	waiting := false
	for _, order := range orders {
		contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, order.ContractId)
		if err != nil {
			return false, err
		}
		if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
			contract.SettleCoin != group.key.coin {
			continue
		}
		waiting = true
		if order.Status == int64(option.OrderStatus_ORDER_STATUS_FUNDING) ||
			order.Status == int64(option.OrderStatus_ORDER_STATUS_PENDING) ||
			order.Status == int64(option.OrderStatus_ORDER_STATUS_PART_FILLED) {
			if err := cancelOptionSystemOrder(l.ctx, l.svcCtx, order.Id, "PORTFOLIO_RISK_LIQUIDATION"); err != nil {
				return false, err
			}
		}
	}
	return waiting, nil
}

func selectPortfolioLiquidationCandidate(
	group *optionRiskGroup,
	snapshot *portfolioRiskSnapshot,
) (*optionRiskPosition, optionrisk.PortfolioResult, optionrisk.PortfolioResult, error) {
	var selected *optionRiskPosition
	selectedRelief := decimal.Zero
	var selectedInitial optionrisk.PortfolioResult
	var selectedMaintenance optionrisk.PortfolioResult
	for i := range group.positions {
		item := &group.positions[i]
		if isInsuranceInventoryPosition(item) ||
			item.position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) ||
			item.contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) ||
			(item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
				item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)) {
			continue
		}
		initialAfter, maintenanceAfter, err := evaluatePortfolioAfterShortReduction(
			snapshot, item, item.position.PositionQty,
		)
		if err != nil {
			return nil, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{}, err
		}
		relief := snapshot.maintenance.Requirement.Sub(maintenanceAfter.Requirement)
		if !relief.IsPositive() {
			continue
		}
		if selected == nil || relief.GreaterThan(selectedRelief) ||
			(relief.Equal(selectedRelief) && item.position.Id < selected.position.Id) {
			selected = item
			selectedRelief = relief
			selectedInitial = initialAfter
			selectedMaintenance = maintenanceAfter
		}
	}
	if selected == nil {
		return nil, optionrisk.PortfolioResult{}, optionrisk.PortfolioResult{},
			errors.New("no full portfolio short position reduces maintenance requirement")
	}
	return selected, selectedInitial, selectedMaintenance, nil
}

func optionTaskSellerMargin(contract *models.TOptionContract, underlyingPrice, premiumPrice, qty decimal.Decimal) decimal.Decimal {
	if contract == nil || !underlyingPrice.IsPositive() || !premiumPrice.IsPositive() || !qty.IsPositive() {
		return decimal.Zero
	}
	multiplier := optionMultiplier(contract)
	underlyingNotional := underlyingPrice.Mul(qty).Mul(multiplier)
	strikeNotional := contract.StrikePrice.Mul(qty).Mul(multiplier)
	otm := decimal.Zero
	minimumBase := contract.MinMarginRate.Mul(underlyingNotional)
	if contract.OptionType == int64(option.OptionType_OPTION_TYPE_CALL) {
		otm = decimal.Max(strikeNotional.Sub(underlyingNotional), decimal.Zero)
	} else {
		otm = decimal.Max(underlyingNotional.Sub(strikeNotional), decimal.Zero)
		minimumBase = contract.MinMarginRate.Mul(strikeNotional)
	}
	return decimal.Max(contract.MaintenanceMarginRate.Mul(underlyingNotional).Sub(otm), minimumBase).Round(16)
}

// optionRiskEquity returns the net liquidation value of the wallet. Premium
// cash flows have already changed Asset total, so adding mark-to-open PnL here
// would count the opening premium twice. The current option asset/liability is
// instead added at its signed mark value.
func optionRiskEquity(totalAsset, netOptionValue decimal.Decimal) decimal.Decimal {
	return totalAsset.Add(netOptionValue).Round(16)
}

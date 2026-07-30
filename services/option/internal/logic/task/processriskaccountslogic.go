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
			return nil, err
		}
		for _, group := range groups {
			if err := l.refreshRiskGroup(group); err != nil {
				return nil, err
			}
		}
		if err := l.resetInactiveRiskAccounts(in.TenantId, groups); err != nil {
			return nil, err
		}
		return helpers.OkTaskResp(), nil
	})
}

type optionRiskKey struct {
	tenantID  int64
	userID    int64
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
}

func (l *ProcessRiskAccountsLogic) collectRiskGroups(tenantID int64) ([]*optionRiskGroup, error) {
	groupMap := make(map[optionRiskKey]*optionRiskGroup)
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
			if contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) &&
				contract.SellerMarginMode != int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO) {
				continue
			}
			if contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) &&
				position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) {
				continue
			}
			market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, position.TenantId, position.ContractId)
			if err != nil {
				return nil, err
			}
			if !market.UnderlyingPrice.IsPositive() || !market.MarkPrice.IsPositive() ||
				market.SnapshotTime <= 0 || market.SnapshotTime > now || now-market.SnapshotTime > 30 {
				return nil, fmt.Errorf("stale option risk market, contractId=%d snapshotTime=%d", contract.Id, market.SnapshotTime)
			}
			key := optionRiskKey{
				tenantID: position.TenantId, userID: position.UserId,
				accountID: position.AccountId, coin: contract.SettleCoin,
			}
			group := groupMap[key]
			if group == nil {
				group = &optionRiskGroup{key: key}
				groupMap[key] = group
			}
			group.positions = append(group.positions, optionRiskPosition{
				position: position, contract: contract, market: market,
			})
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
	portfolioLegs := make(map[int64]optionrisk.PortfolioLeg)
	for _, item := range group.positions {
		position := item.position
		position.MarkPrice = item.market.MarkPrice
		position.PositionValue = item.market.MarkPrice.Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
		if position.Side == int64(common.PositionSide_POSITION_SIDE_LONG) {
			position.UnrealizedPnl = item.market.MarkPrice.Sub(position.OpenAvgPrice).
				Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
		} else {
			position.UnrealizedPnl = position.OpenAvgPrice.Sub(item.market.MarkPrice).
				Mul(position.PositionQty).Mul(optionMultiplier(item.contract)).Round(16)
		}
		if item.contract.SellerMarginMode == int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED) {
			position.MaintenanceMargin = optionTaskSellerMargin(
				item.contract, item.market.UnderlyingPrice, item.market.MarkPrice, position.PositionQty,
			)
			positionMargin = positionMargin.Add(position.MarginAmount)
			maintenanceMargin = maintenanceMargin.Add(position.MaintenanceMargin)
		} else {
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
	portfolioInitialRequirement := decimal.Zero
	if len(portfolioLegs) > 0 {
		legs := make([]optionrisk.PortfolioLeg, 0, len(portfolioLegs))
		for _, leg := range portfolioLegs {
			legs = append(legs, leg)
		}
		initialResult, err := optionrisk.EvaluatePortfolio(legs, false)
		if err != nil {
			return err
		}
		maintenanceResult, err := optionrisk.EvaluatePortfolio(legs, true)
		if err != nil {
			return err
		}
		positionMargin = positionMargin.Add(initialResult.Requirement)
		portfolioInitialRequirement = initialResult.Requirement
		maintenanceMargin = maintenanceMargin.Add(maintenanceResult.Requirement)
		portfolioMethod = option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1
		portfolioScenarioLoss = maintenanceResult.ScenarioLoss
		portfolioShortFloor = maintenanceResult.ShortFloor
	}
	equity := totalAsset.Add(unrealizedPnL).Round(16)
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
			MaintenanceMargin: maintenanceMargin, UnrealizedPnl: unrealizedPnL, RiskRate: riskRate,
			PortfolioRiskMethod:   int64(portfolioMethod),
			PortfolioScenarioLoss: portfolioScenarioLoss, PortfolioShortFloor: portfolioShortFloor,
			Status: int64(status), LastCalcTime: now, CreateTimes: now, UpdateTimes: now,
		})
		if err != nil {
			return err
		}
		if err := l.rebalancePortfolioCollateral(group, portfolioInitialRequirement, now); err != nil {
			return err
		}
		return l.ensureLiquidations(group, status, equity, maintenanceMargin, now)
	}
	current.Equity = equity
	current.PositionMargin = positionMargin
	current.MaintenanceMargin = maintenanceMargin
	current.UnrealizedPnl = unrealizedPnL
	current.RiskRate = riskRate
	current.PortfolioRiskMethod = int64(portfolioMethod)
	current.PortfolioScenarioLoss = portfolioScenarioLoss
	current.PortfolioShortFloor = portfolioShortFloor
	current.Status = int64(status)
	current.LastCalcTime = now
	current.UpdateTimes = now
	if err := l.svcCtx.OptionRiskAccountModel.Update(l.ctx, current); err != nil {
		return err
	}
	if err := l.rebalancePortfolioCollateral(group, portfolioInitialRequirement, now); err != nil {
		return err
	}
	return l.ensureLiquidations(group, status, equity, maintenanceMargin, now)
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
				UserId: group.key.userID, AccountId: group.key.accountID,
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
		account.MaintenanceMargin = decimal.Zero
		account.UnrealizedPnl = decimal.Zero
		account.RiskRate = decimal.Zero
		account.PortfolioRiskMethod = int64(option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_UNKNOWN)
		account.PortfolioScenarioLoss = decimal.Zero
		account.PortfolioShortFloor = decimal.Zero
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
) error {
	if status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_LIQUIDATING &&
		status != option.RiskAccountStatus_RISK_ACCOUNT_STATUS_BANKRUPT {
		return nil
	}
	deficit := decimal.Max(maintenanceMargin.Sub(equity), decimal.Zero).Round(16)
	for _, item := range group.positions {
		if item.position.Side != int64(common.PositionSide_POSITION_SIDE_SHORT) {
			continue
		}
		if item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) &&
			item.contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) {
			continue
		}
		_, err := l.svcCtx.OptionLiquidationModel.FindOpenByPosition(
			l.ctx, item.position.TenantId, item.position.Id,
		)
		if err == nil {
			continue
		}
		if !errors.Is(err, models.ErrNotFound) {
			return err
		}
		liquidationNo := fmt.Sprintf("OLQ-%d-%d-%d", item.position.TenantId, item.position.Id, now)
		fee := item.position.PositionValue.Mul(item.contract.LiquidationFeeRate).Round(16)
		if _, err := l.svcCtx.OptionLiquidationModel.Insert(l.ctx, &models.TOptionLiquidation{
			TenantId: item.position.TenantId, LiquidationNo: liquidationNo,
			UserId: item.position.UserId, AccountId: item.position.AccountId,
			ContractId: item.position.ContractId, PositionId: item.position.Id,
			Quantity: item.position.PositionQty, MarkPrice: item.market.MarkPrice,
			MaintenanceMargin: item.position.MaintenanceMargin, Equity: equity,
			DeficitAmount: deficit, LiquidationFee: fee,
			Status: int64(option.LiquidationStatus_LIQUIDATION_STATUS_PENDING),
			DeficitResolution: int64(
				option.LiquidationDeficitResolution_LIQUIDATION_DEFICIT_RESOLUTION_NONE,
			),
			CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
	}
	return nil
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

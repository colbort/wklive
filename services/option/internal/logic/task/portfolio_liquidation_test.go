package tasklogic

import (
	"testing"

	"wklive/proto/common"
	"wklive/proto/option"
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func TestSelectPortfolioLiquidationCandidateDeterministicTie(t *testing.T) {
	config := optionrisk.PortfolioConfig{
		InitialShockRate:     decimal.RequireFromString("0.2"),
		MaintenanceShockRate: decimal.RequireFromString("0.1"),
		ScenarioShocks: []decimal.Decimal{
			decimal.NewFromInt(-1), decimal.RequireFromString("-0.2"), decimal.Zero,
			decimal.RequireFromString("0.2"), decimal.NewFromInt(4),
		},
		ConcentrationThreshold: decimal.NewFromInt(1000000),
	}
	contractA := portfolioLiquidationTestContract(101)
	contractB := portfolioLiquidationTestContract(102)
	marketA := portfolioLiquidationTestMarket(101)
	marketB := portfolioLiquidationTestMarket(102)
	positionHighID := &models.TOptionPosition{
		Id: 22, ContractId: contractA.Id, Side: int64(common.PositionSide_POSITION_SIDE_SHORT),
		PositionQty: decimal.NewFromInt(1),
	}
	positionLowID := &models.TOptionPosition{
		Id: 11, ContractId: contractB.Id, Side: int64(common.PositionSide_POSITION_SIDE_SHORT),
		PositionQty: decimal.NewFromInt(1),
	}
	legs := map[int64]optionrisk.PortfolioLeg{
		contractA.Id: {Contract: contractA, Market: marketA, ShortQuantity: decimal.NewFromInt(1)},
		contractB.Id: {Contract: contractB, Market: marketB, ShortQuantity: decimal.NewFromInt(1)},
	}
	legList := []optionrisk.PortfolioLeg{legs[contractA.Id], legs[contractB.Id]}
	initial, err := optionrisk.EvaluatePortfolio(legList, false, config)
	if err != nil {
		t.Fatal(err)
	}
	maintenance, err := optionrisk.EvaluatePortfolio(legList, true, config)
	if err != nil {
		t.Fatal(err)
	}
	group := &optionRiskGroup{positions: []optionRiskPosition{
		{position: positionHighID, contract: contractA, market: marketA},
		{position: positionLowID, contract: contractB, market: marketB},
	}}
	selected, initialAfter, maintenanceAfter, err := selectPortfolioLiquidationCandidate(
		group,
		&portfolioRiskSnapshot{config: config, legs: legs, initial: initial, maintenance: maintenance},
	)
	if err != nil {
		t.Fatal(err)
	}
	if selected.position.Id != positionLowID.Id {
		t.Fatalf("tie selected position %d want lower id %d", selected.position.Id, positionLowID.Id)
	}
	if !maintenanceAfter.Requirement.LessThan(maintenance.Requirement) || initialAfter.Requirement.IsNegative() {
		t.Fatalf("candidate did not reduce risk before=%s after=%s initialAfter=%s",
			maintenance.Requirement, maintenanceAfter.Requirement, initialAfter.Requirement)
	}
}

func TestSelectPortfolioLiquidationQuantityUsesScenarioModelAndStrictHealth(t *testing.T) {
	config := optionrisk.PortfolioConfig{
		InitialShockRate:     decimal.RequireFromString("0.2"),
		MaintenanceShockRate: decimal.RequireFromString("0.1"),
		ScenarioShocks: []decimal.Decimal{
			decimal.NewFromInt(-1), decimal.RequireFromString("-0.2"), decimal.Zero,
			decimal.RequireFromString("0.2"), decimal.NewFromInt(4),
		},
		ConcentrationThreshold: decimal.NewFromInt(1000000),
	}
	contract := portfolioLiquidationTestContract(201)
	contract.QtyStep = decimal.NewFromInt(1)
	contract.LiquidationFeeRate = decimal.RequireFromString("0.01")
	market := portfolioLiquidationTestMarket(contract.Id)
	position := &models.TOptionPosition{
		Id: 31, ContractId: contract.Id, Side: int64(common.PositionSide_POSITION_SIDE_SHORT),
		PositionQty: decimal.NewFromInt(2),
	}
	legs := map[int64]optionrisk.PortfolioLeg{
		contract.Id: {Contract: contract, Market: market, ShortQuantity: decimal.NewFromInt(2)},
	}
	initial, err := optionrisk.EvaluatePortfolio([]optionrisk.PortfolioLeg{legs[contract.Id]}, false, config)
	if err != nil {
		t.Fatal(err)
	}
	maintenance, err := optionrisk.EvaluatePortfolio([]optionrisk.PortfolioLeg{legs[contract.Id]}, true, config)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := &portfolioRiskSnapshot{
		config: config, legs: legs, initial: initial, maintenance: maintenance,
	}
	candidate := &optionRiskPosition{position: position, contract: contract, market: market}
	_, afterOne, err := evaluatePortfolioAfterShortReduction(snapshot, candidate, decimal.NewFromInt(1))
	if err != nil {
		t.Fatal(err)
	}
	feeOne := market.MarkPrice.Mul(contract.LiquidationFeeRate)
	equity := afterOne.Requirement.Add(feeOne).Add(decimal.NewFromInt(1))
	if !maintenance.Requirement.GreaterThanOrEqual(equity) {
		t.Fatalf("fixture does not trigger liquidation maintenance/equity=%s/%s", maintenance.Requirement, equity)
	}
	quantity, _, selectedMaintenance, err := selectPortfolioLiquidationQuantity(
		candidate, snapshot, equity, maintenance.Requirement, initial.Requirement.Add(decimal.NewFromInt(100)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !quantity.Equal(decimal.NewFromInt(1)) || !selectedMaintenance.Requirement.Equal(afterOne.Requirement) {
		t.Fatalf("selected quantity/maintenance=%s/%s want=1/%s",
			quantity, selectedMaintenance.Requirement, afterOne.Requirement)
	}

	boundaryEquity := afterOne.Requirement.Add(feeOne)
	quantity, _, _, err = selectPortfolioLiquidationQuantity(
		candidate, snapshot, boundaryEquity, maintenance.Requirement,
		initial.Requirement.Add(decimal.NewFromInt(100)),
	)
	if err != nil {
		t.Fatal(err)
	}
	if !quantity.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("exact maintenance boundary selected quantity=%s want=2", quantity)
	}
}

func TestSelectPortfolioLiquidationQuantityRejectsUnboundedGrid(t *testing.T) {
	contract := portfolioLiquidationTestContract(202)
	contract.QtyStep = decimal.RequireFromString("0.000001")
	position := &models.TOptionPosition{
		Id: 32, Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
	}
	_, _, _, err := selectPortfolioLiquidationQuantity(
		&optionRiskPosition{position: position, contract: contract, market: portfolioLiquidationTestMarket(contract.Id)},
		&portfolioRiskSnapshot{}, decimal.Zero, decimal.Zero, decimal.Zero,
	)
	if err == nil {
		t.Fatal("unbounded portfolio liquidation quantity grid unexpectedly accepted")
	}
}

func TestAllocateIsolatedLiquidationLotsProportionally(t *testing.T) {
	lots := []*models.TOptionMarginLot{
		{Id: 1, RemainingQuantity: decimal.NewFromInt(2), RemainingMargin: decimal.NewFromInt(100)},
		{Id: 2, RemainingQuantity: decimal.NewFromInt(1), RemainingMargin: decimal.NewFromInt(30)},
	}
	allocations, margin, err := allocateIsolatedLiquidationLots(
		lots, decimal.RequireFromString("2.5"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(allocations) != 2 ||
		!allocations[0].quantity.Equal(decimal.NewFromInt(2)) ||
		!allocations[0].margin.Equal(decimal.NewFromInt(100)) ||
		!allocations[1].quantity.Equal(decimal.RequireFromString("0.5")) ||
		!allocations[1].margin.Equal(decimal.NewFromInt(15)) ||
		!margin.Equal(decimal.NewFromInt(115)) {
		t.Fatalf("unexpected proportional allocations=%+v margin=%s", allocations, margin)
	}
}

func TestAllocateIsolatedLiquidationLotsRejectsPendingOrInsufficientQuantity(t *testing.T) {
	if _, _, err := allocateIsolatedLiquidationLots([]*models.TOptionMarginLot{{
		RemainingQuantity: decimal.NewFromInt(2), RemainingMargin: decimal.NewFromInt(100),
		PendingMargin: decimal.NewFromInt(1),
	}}, decimal.NewFromInt(1)); err == nil {
		t.Fatal("pending margin lot unexpectedly accepted")
	}
	if _, _, err := allocateIsolatedLiquidationLots([]*models.TOptionMarginLot{{
		RemainingQuantity: decimal.NewFromInt(1), RemainingMargin: decimal.NewFromInt(50),
	}}, decimal.NewFromInt(2)); err == nil {
		t.Fatal("insufficient margin-lot quantity unexpectedly accepted")
	}
}

func portfolioLiquidationTestContract(id int64) *models.TOptionContract {
	return &models.TOptionContract{
		Id: id, UnderlyingSymbol: "BTCUSDT", SettleCoin: "USDT",
		OptionType: int64(option.OptionType_OPTION_TYPE_CALL), StrikePrice: decimal.NewFromInt(100),
		ContractUnit: decimal.NewFromInt(1), Multiplier: decimal.NewFromInt(1), ExpireTime: 2000000000,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO),
		InitialMarginRate:     decimal.RequireFromString("0.2"),
		MaintenanceMarginRate: decimal.RequireFromString("0.1"),
		MinMarginRate:         decimal.RequireFromString("0.05"),
		Status:                int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
	}
}

func portfolioLiquidationTestMarket(contractID int64) *models.TOptionMarket {
	return &models.TOptionMarket{
		ContractId: contractID, UnderlyingPrice: decimal.NewFromInt(100), MarkPrice: decimal.NewFromInt(10),
	}
}

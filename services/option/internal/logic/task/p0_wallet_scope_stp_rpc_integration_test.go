package tasklogic

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func testP0WalletRestrictionAndCrossAccountSTP(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		restrictedUserID int64 = 880001
		stpUserID        int64 = 880002
	)
	now := time.Now().Unix()
	calendarCode := "P0_WALLET_SCOPE_STP_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	restrictedContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-WALLET-RESTRICTION-CALL", calendarCode, 880003, now,
	)
	stpContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-WALLET-SCOPE-STP-CALL", calendarCode, 880004, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, restrictedContract.Id, "100", "10", now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, stpContract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, restrictedUserID, "100", "P0-WALLET-RESTRICTION-SEED")
	creditAsset(t, ctx, assetClient, stpUserID, "100", "P0-WALLET-SCOPE-STP-SEED")

	insertP0RestrictedWalletRiskAccount(t, ctx, serviceCtx, restrictedUserID, now)
	for index, accountID := range []int64{7101, 7102} {
		resp, err := applogic.NewPlaceOrderLogic(
			p0OrderUserContext(ctx, restrictedUserID), serviceCtx,
		).PlaceOrder(&option.PlaceOrderReq{
			AccountId: accountID, ContractId: restrictedContract.Id,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
			ClientOrderId: "P0-WALLET-RESTRICTION-" + string(rune('A'+index)),
		})
		if err != nil {
			t.Fatalf("restricted wallet order account=%d: %v", accountID, err)
		}
		if resp == nil || resp.Base == nil || resp.Base.Code == 200 {
			t.Fatalf("restricted wallet account=%d unexpectedly increased risk: %+v", accountID, resp)
		}
	}
	assertP0WalletRestrictionEvidence(t, ctx, db, restrictedContract.Id, restrictedUserID)

	sellerResp := placeP0Order(t, ctx, serviceCtx, stpUserID, &option.PlaceOrderReq{
		AccountId: 8201, ContractId: stpContract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-WALLET-SCOPE-STP-MAKER",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	sellerOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerResp.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if sellerOrder.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
		t.Fatalf("cross-account STP maker status=%d want pending", sellerOrder.Status)
	}

	buyerResp := placeP0Order(t, ctx, serviceCtx, stpUserID, &option.PlaceOrderReq{
		AccountId: 8202, ContractId: stpContract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-WALLET-SCOPE-STP-TAKER",
	})
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0CrossAccountSTPBeforeMakerCancel(
		t, ctx, db, serviceCtx, stpContract.Id, sellerOrder.Id, buyerResp.Data.OrderId, stpUserID,
	)

	assertP0UserCancelOK(t, ctx, serviceCtx, stpUserID, 8201, sellerOrder.Id)
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0CrossAccountSTPFinal(
		t, ctx, db, stpContract.Id, sellerOrder.Id, buyerResp.Data.OrderId, stpUserID,
	)
	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0CrossAccountSTPFinal(
		t, ctx, db, stpContract.Id, sellerOrder.Id, buyerResp.Data.OrderId, stpUserID,
	)
	testP0ConcurrentPortfolioAdmission(t, ctx, db, assetClient, serviceCtx, calendarCode, now)
}

func testP0ConcurrentPortfolioAdmission(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	const portfolioUserID int64 = 880005
	contractA := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-PORTFOLIO-CROSS-ACCOUNT-A", calendarCode, 880006, now,
	)
	contractB := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-PORTFOLIO-CROSS-ACCOUNT-B", calendarCode, 880007, now,
	)
	contractB.StrikePrice = decimal.NewFromInt(120)
	for _, contract := range []*models.TOptionContract{contractA, contractB} {
		contract.SellerMarginMode = int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO)
		if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
			t.Fatalf("enable portfolio admission contract %d: %v", contract.Id, err)
		}
	}
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractA.Id, "100", "10", now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractB.Id, "100", "5", now)
	configItem, err := serviceCtx.OptionPortfolioRiskConfigModel.FindActive(
		ctx, p0AssetE2ETenantID, "USDT", now,
	)
	if err != nil {
		t.Fatalf("find active portfolio config: %v", err)
	}
	config, err := optionrisk.PortfolioConfigFromModel(configItem)
	if err != nil {
		t.Fatalf("parse active portfolio config: %v", err)
	}
	marketA, err := serviceCtx.OptionMarketModel.FindOneByTenantIdContractId(
		ctx, p0AssetE2ETenantID, contractA.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	marketB, err := serviceCtx.OptionMarketModel.FindOneByTenantIdContractId(
		ctx, p0AssetE2ETenantID, contractB.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	expected, err := optionrisk.EvaluatePortfolio([]optionrisk.PortfolioLeg{
		{Contract: contractA, Market: marketA, ShortQuantity: decimal.NewFromInt(1)},
		{Contract: contractB, Market: marketB, ShortQuantity: decimal.NewFromInt(1)},
	}, false, config)
	if err != nil {
		t.Fatalf("evaluate concurrent portfolio fixture: %v", err)
	}
	if !expected.Requirement.IsPositive() {
		t.Fatalf("concurrent portfolio fixture requirement=%s", expected.Requirement)
	}
	creditAsset(t, ctx, assetClient, portfolioUserID, "1000", "P0-PORTFOLIO-CROSS-ACCOUNT-SEED")

	type admissionResult struct {
		response *option.PlaceOrderResp
		err      error
	}
	requests := []*option.PlaceOrderReq{
		{
			AccountId: 8301, ContractId: contractA.Id,
			Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
			ClientOrderId: "P0-PORTFOLIO-CROSS-ACCOUNT-A",
		},
		{
			AccountId: 8302, ContractId: contractB.Id,
			Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "5", Qty: "1",
			ClientOrderId: "P0-PORTFOLIO-CROSS-ACCOUNT-B",
		},
	}
	start := make(chan struct{})
	results := make(chan admissionResult, len(requests))
	var workers sync.WaitGroup
	for _, request := range requests {
		request := request
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			response, placeErr := applogic.NewPlaceOrderLogic(
				p0OrderUserContext(ctx, portfolioUserID), serviceCtx,
			).PlaceOrder(request)
			results <- admissionResult{response: response, err: placeErr}
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	orderIDs := make([]int64, 0, len(requests))
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent portfolio admission: %v", result.err)
		}
		if result.response == nil || result.response.Base == nil || result.response.Base.Code != 200 ||
			result.response.Data == nil {
			t.Fatalf("concurrent portfolio admission response: %+v", result.response)
		}
		orderIDs = append(orderIDs, result.response.Data.OrderId)
	}
	if len(orderIDs) != 2 {
		t.Fatalf("concurrent portfolio admission orders=%d", len(orderIDs))
	}
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0ConcurrentPortfolioAdmissionFunded(
		t, ctx, db, portfolioUserID, orderIDs, configItem.Id, configItem.Version, expected.Requirement,
	)
	for _, orderID := range orderIDs {
		order, findErr := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
		if findErr != nil {
			t.Fatal(findErr)
		}
		assertP0UserCancelOK(t, ctx, serviceCtx, portfolioUserID, order.AccountId, orderID)
	}
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0ConcurrentPortfolioAdmissionFinal(t, ctx, db, portfolioUserID, orderIDs, expected.Requirement)
	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0ConcurrentPortfolioAdmissionFinal(t, ctx, db, portfolioUserID, orderIDs, expected.Requirement)
}

func assertP0ConcurrentPortfolioAdmissionFunded(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID int64,
	orderIDs []int64,
	configID, configVersion int64,
	expectedMargin decimal.Decimal,
) {
	t.Helper()
	var orders, pending, accounts, matchingConfig, instructions, success, reconciled, riskAccounts, walletAccounts int64
	var margin decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),COUNT(DISTINCT account_id),
		SUM(portfolio_risk_config_id=? AND portfolio_risk_config_version=?),SUM(margin_amount)
		FROM t_option_order WHERE tenant_id=? AND id IN (?,?)`,
		int64(option.OrderStatus_ORDER_STATUS_PENDING), configID, configVersion,
		p0AssetE2ETenantID, orderIDs[0], orderIDs[1],
	).Scan(&orders, &pending, &accounts, &matchingConfig, &margin); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, orderIDs[0], orderIDs[1],
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(account_id=0)
		FROM t_option_risk_account WHERE tenant_id=? AND user_id=? AND settle_coin='USDT'`,
		p0AssetE2ETenantID, userID,
	).Scan(&riskAccounts, &walletAccounts); err != nil {
		t.Fatal(err)
	}
	if orders != 2 || pending != 2 || accounts != 2 || matchingConfig != 2 ||
		!margin.Equal(expectedMargin) || instructions != 2 || success != 2 || reconciled != 2 ||
		riskAccounts != 1 || walletAccounts != 1 {
		t.Fatalf("portfolio concurrent funded orders/pending/accounts/config/margin/instructions/success/reconciled/risk/wallet=%d/%d/%d/%d/%s/%d/%d/%d/%d/%d want margin=%s",
			orders, pending, accounts, matchingConfig, margin, instructions, success, reconciled,
			riskAccounts, walletAccounts, expectedMargin)
	}
	assertWalletCoinAmounts(t, ctx, db, userID, "USDT", "1000.000000000000000000",
		decimal.NewFromInt(1000).Sub(expectedMargin).StringFixed(18), expectedMargin.StringFixed(18))
}

func assertP0ConcurrentPortfolioAdmissionFinal(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID int64,
	orderIDs []int64,
	expectedMargin decimal.Decimal,
) {
	t.Helper()
	var orders, canceled, instructions, success, reconciled, flows int64
	var freezeAmount, releaseAmount decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?) FROM t_option_order
		WHERE tenant_id=? AND id IN (?,?)`,
		int64(option.OrderStatus_ORDER_STATUS_CANCELED), p0AssetE2ETenantID, orderIDs[0], orderIDs[1],
	).Scan(&orders, &canceled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?),
		SUM(CASE WHEN action=1 THEN amount ELSE 0 END),SUM(CASE WHEN action=3 THEN amount ELSE 0 END)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, orderIDs[0], orderIDs[1],
	).Scan(&instructions, &success, &reconciled, &freezeAmount, &releaseAmount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id IN (?,?)`,
		p0AssetE2ETenantID, orderIDs[0], orderIDs[1],
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if orders != 2 || canceled != 2 || instructions != 4 || success != 4 || reconciled != 4 ||
		flows != 4 || !freezeAmount.Equal(expectedMargin) || !releaseAmount.Equal(expectedMargin) {
		t.Fatalf("portfolio concurrent final orders/canceled/instructions/success/reconciled/flows/freeze/release=%d/%d/%d/%d/%d/%d/%s/%s want=%s",
			orders, canceled, instructions, success, reconciled, flows, freezeAmount, releaseAmount, expectedMargin)
	}
	assertWalletCoinAmounts(t, ctx, db, userID, "USDT",
		"1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
}

func insertP0RestrictedWalletRiskAccount(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, now int64,
) {
	t.Helper()
	_, err := serviceCtx.OptionRiskAccountModel.Insert(ctx, &models.TOptionRiskAccount{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: 0, SettleCoin: "USDT",
		Equity: decimal.NewFromInt(100), Status: int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED),
		LastCalcTime: now, CreateTimes: now, UpdateTimes: now,
	})
	if err != nil {
		t.Fatalf("insert restricted wallet risk account: %v", err)
	}
}

func assertP0WalletRestrictionEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, userID int64,
) {
	t.Helper()
	var riskAccounts, restricted, accountID, orders, clientKeys, instructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),MAX(account_id)
		FROM t_option_risk_account
		WHERE tenant_id=? AND user_id=? AND settle_coin='USDT'`,
		int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED),
		p0AssetE2ETenantID, userID,
	).Scan(&riskAccounts, &restricted, &accountID); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_order
		WHERE tenant_id=? AND user_id=? AND contract_id=?`,
		p0AssetE2ETenantID, userID, contractID,
	).Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_client_order_key
		WHERE tenant_id=? AND user_id=?`, p0AssetE2ETenantID, userID).Scan(&clientKeys); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND user_id=?`, p0AssetE2ETenantID, userID).Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if riskAccounts != 1 || restricted != 1 || accountID != 0 || orders != 0 || clientKeys != 0 || instructions != 0 {
		t.Fatalf("wallet restriction risk/restricted/account/orders/keys/instructions=%d/%d/%d/%d/%d/%d",
			riskAccounts, restricted, accountID, orders, clientKeys, instructions)
	}
}

func assertP0CrossAccountSTPBeforeMakerCancel(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, sellerOrderID, buyerOrderID, userID int64,
) {
	t.Helper()
	seller, err := serviceCtx.OptionOrderModel.FindOne(ctx, sellerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	buyer, err := serviceCtx.OptionOrderModel.FindOne(ctx, buyerOrderID)
	if err != nil {
		t.Fatal(err)
	}
	if seller.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) ||
		buyer.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
		buyer.CancelReason != "SELF_TRADE_PREVENTED" || seller.AccountId == buyer.AccountId {
		t.Fatalf("cross-account STP maker/taker=%+v/%+v", seller, buyer)
	}
	var trades, events, positions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trade
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&trades); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event
		WHERE tenant_id=? AND user_id=? AND contract_id=?
		  AND event_type='STP_PREVENTED' AND reason='SELF_TRADE_PREVENTED'`,
		p0AssetE2ETenantID, userID, contractID,
	).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_position
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&positions); err != nil {
		t.Fatal(err)
	}
	if trades != 0 || events != 1 || positions != 0 {
		t.Fatalf("cross-account STP before maker cancel trades/events/positions=%d/%d/%d",
			trades, events, positions)
	}
	assertWalletCoinAmounts(t, ctx, db, userID, "USDT",
		"100.000000000000000000", "50.000000000000000000", "50.000000000000000000")
}

func assertP0CrossAccountSTPFinal(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, sellerOrderID, buyerOrderID, userID int64,
) {
	t.Helper()
	var orders, canceled, accountCount, trades, events, instructions, success, reconciled, flows, positions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),COUNT(DISTINCT account_id)
		FROM t_option_order WHERE tenant_id=? AND id IN (?,?)`,
		int64(option.OrderStatus_ORDER_STATUS_CANCELED),
		p0AssetE2ETenantID, sellerOrderID, buyerOrderID,
	).Scan(&orders, &canceled, &accountCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trade
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&trades); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event
		WHERE tenant_id=? AND user_id=? AND contract_id=?
		  AND event_type='STP_PREVENTED' AND reason='SELF_TRADE_PREVENTED'`,
		p0AssetE2ETenantID, userID, contractID,
	).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?),SUM(reconciliation_status=?)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?)`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED),
		p0AssetE2ETenantID, sellerOrderID, buyerOrderID,
	).Scan(&instructions, &success, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id IN (?,?)`,
		p0AssetE2ETenantID, sellerOrderID, buyerOrderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_position
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&positions); err != nil {
		t.Fatal(err)
	}
	if orders != 2 || canceled != 2 || accountCount != 2 || trades != 0 || events != 1 ||
		instructions != 4 || success != 4 || reconciled != 4 || flows != 4 || positions != 0 {
		t.Fatalf("cross-account STP orders/canceled/accounts/trades/events/instructions/success/reconciled/flows/positions=%d/%d/%d/%d/%d/%d/%d/%d/%d/%d",
			orders, canceled, accountCount, trades, events, instructions, success, reconciled, flows, positions)
	}
	assertWalletCoinAmounts(t, ctx, db, userID, "USDT",
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}

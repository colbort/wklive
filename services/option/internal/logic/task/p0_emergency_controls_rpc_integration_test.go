package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
)

func testP0EmergencyTradingControls(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_EMERGENCY_CONTROLS_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	testP0KillSwitchReleaseBarrier(
		t, ctx, db, assetClient, serviceCtx, calendarCode, now,
	)
	testP0KillSwitchMatchRace(
		t, ctx, db, assetClient, serviceCtx, calendarCode, now,
	)
	testP0CircuitBreakerBatchCancel(
		t, ctx, db, assetClient, serviceCtx, calendarCode, now,
	)
}

func testP0KillSwitchReleaseBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	const (
		userID     int64 = 96001
		sellerID   int64 = 96002
		feeUserID  int64 = 96000
		operatorID int64 = 96099
		accountID  int64 = 96001
	)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-KILL-SWITCH-RELEASE-BARRIER", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, userID, "500", "P0-KILL-BARRIER-USER-SEED")
	creditAsset(t, ctx, assetClient, sellerID, "200", "P0-KILL-BARRIER-SELLER-SEED")

	pending := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P0-KILL-BARRIER-PENDING",
	})
	processAssetInstructions(t, ctx, serviceCtx)

	partFilled := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "2",
		ClientOrderId: "P0-KILL-BARRIER-PART-FILLED",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	seller := placeP0Order(t, ctx, serviceCtx, sellerID, &option.PlaceOrderReq{
		AccountId: sellerID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-KILL-BARRIER-PARTIAL-TAKER",
	})
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	processP0TradeEvents(t, ctx, serviceCtx)

	funding := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "8", Qty: "1",
		ClientOrderId: "P0-KILL-BARRIER-FUNDING",
	})
	assertP0OrderStatus(t, ctx, serviceCtx, pending.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	assertP0OrderStatus(t, ctx, serviceCtx, partFilled.Data.OrderId, option.OrderStatus_ORDER_STATUS_PART_FILLED)
	assertP0OrderStatus(t, ctx, serviceCtx, funding.Data.OrderId, option.OrderStatus_ORDER_STATUS_FUNDING)
	assertP0OrderStatus(t, ctx, serviceCtx, seller.Data.OrderId, option.OrderStatus_ORDER_STATUS_FILLED)

	activated, err := applogic.NewActivateKillSwitchLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).ActivateKillSwitch(&option.ActivateKillSwitchReq{Reason: "P0_RELEASE_BARRIER"})
	if err != nil || activated == nil || activated.Base == nil || activated.Base.Code != 200 ||
		activated.Data == nil || activated.Data.KillSwitch != common.YesNo_YES_NO_YES {
		t.Fatalf("activate release-barrier kill switch resp=%+v err=%v", activated, err)
	}
	assertP0OrderStatus(t, ctx, serviceCtx, pending.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELING)
	assertP0OrderStatus(t, ctx, serviceCtx, partFilled.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELING)
	assertP0OrderStatus(t, ctx, serviceCtx, funding.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)

	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	blocked, err := adminlogic.NewReleaseUserKillSwitchLogic(adminCtx, serviceCtx).
		ReleaseUserKillSwitch(&option.ReleaseUserKillSwitchReq{
			TenantId: p0AssetE2ETenantID, UserId: userID, Reason: "TOO_EARLY",
		})
	if err != nil || blocked == nil || blocked.Base == nil || blocked.Base.Code == 200 {
		t.Fatalf("kill switch release did not block non-terminal orders resp=%+v err=%v", blocked, err)
	}

	rejected, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).PlaceOrder(&option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P0-KILL-BARRIER-REJECTED-WHILE-ACTIVE",
	})
	if err != nil || rejected == nil || rejected.Base == nil || rejected.Base.Code == 200 ||
		rejected.Base.Msg != "USER_KILL_SWITCH" {
		t.Fatalf("kill switch did not reject new order resp=%+v err=%v", rejected, err)
	}

	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0OrderStatus(t, ctx, serviceCtx, pending.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)
	assertP0OrderStatus(t, ctx, serviceCtx, partFilled.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)
	released, err := adminlogic.NewReleaseUserKillSwitchLogic(adminCtx, serviceCtx).
		ReleaseUserKillSwitch(&option.ReleaseUserKillSwitchReq{
			TenantId: p0AssetE2ETenantID, UserId: userID, Reason: "FUNDS_RECONCILED",
		})
	if err != nil || released == nil || released.Base == nil || released.Base.Code != 200 {
		t.Fatalf("release reconciled kill switch resp=%+v err=%v", released, err)
	}

	postRelease := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P0-KILL-BARRIER-POST-RELEASE",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0UserCancelOK(t, ctx, serviceCtx, userID, accountID, postRelease.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0KillSwitchReleaseEvidence(
		t, ctx, db, contract.Id, userID, sellerID, feeUserID,
		pending.Data.OrderId, partFilled.Data.OrderId, funding.Data.OrderId,
	)
}

type p0KillActivationResult struct {
	response *option.GetUserTradingControlResp
	err      error
}

type p0EmergencyTaskResult struct {
	response *option.OptionTaskResp
	err      error
}

func testP0KillSwitchMatchRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	const (
		makerID    int64 = 96101
		takerID    int64 = 96102
		feeUserID  int64 = 96100
		operatorID int64 = 96199
	)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-KILL-SWITCH-MATCH-RACE", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, makerID, "200", "P0-KILL-RACE-MAKER-SEED")
	creditAsset(t, ctx, assetClient, takerID, "200", "P0-KILL-RACE-TAKER-SEED")

	maker := placeP0Order(t, ctx, serviceCtx, makerID, &option.PlaceOrderReq{
		AccountId: makerID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-KILL-RACE-MAKER",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	taker := placeP0Order(t, ctx, serviceCtx, takerID, &option.PlaceOrderReq{
		AccountId: takerID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-KILL-RACE-TAKER",
	})

	lockTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	var lockedOrderID int64
	if err := lockTx.QueryRowContext(
		ctx, `SELECT id FROM t_option_order WHERE tenant_id=? AND id=? FOR UPDATE`,
		p0AssetE2ETenantID, maker.Data.OrderId,
	).Scan(&lockedOrderID); err != nil {
		_ = lockTx.Rollback()
		t.Fatal(err)
	}

	activationResult := make(chan p0KillActivationResult, 1)
	go func() {
		response, activateErr := applogic.NewActivateKillSwitchLogic(
			p0OrderUserContext(ctx, makerID), serviceCtx,
		).ActivateKillSwitch(&option.ActivateKillSwitchReq{Reason: "P0_MATCH_RACE"})
		activationResult <- p0KillActivationResult{response: response, err: activateErr}
	}()
	waitP0KillSwitchActive(t, ctx, db, makerID)

	taskResult := make(chan p0EmergencyTaskResult, 1)
	go func() {
		response, taskErr := NewProcessAssetInstructionsLogic(ctx, serviceCtx).
			ProcessAssetInstructions(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID})
		taskResult <- p0EmergencyTaskResult{response: response, err: taskErr}
	}()
	time.Sleep(100 * time.Millisecond)
	if err := lockTx.Commit(); err != nil {
		t.Fatal(err)
	}

	select {
	case result := <-activationResult:
		if result.err != nil || result.response == nil || result.response.Base == nil ||
			result.response.Base.Code != 200 {
			t.Fatalf("concurrent kill activation resp=%+v err=%v", result.response, result.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent kill activation timed out")
	}
	select {
	case result := <-taskResult:
		if result.err != nil || result.response == nil || result.response.Base == nil ||
			result.response.Base.Code != 200 {
			t.Fatalf("concurrent funding/match task resp=%+v err=%v", result.response, result.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("concurrent funding/match task timed out")
	}
	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}

	assertP0OrderStatus(t, ctx, serviceCtx, maker.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)
	assertP0OrderStatus(t, ctx, serviceCtx, taker.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	assertP0UserCancelOK(t, ctx, serviceCtx, takerID, takerID, taker.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	released, err := adminlogic.NewReleaseUserKillSwitchLogic(adminCtx, serviceCtx).
		ReleaseUserKillSwitch(&option.ReleaseUserKillSwitchReq{
			TenantId: p0AssetE2ETenantID, UserId: makerID, Reason: "RACE_RECONCILED",
		})
	if err != nil || released == nil || released.Base == nil || released.Base.Code != 200 {
		t.Fatalf("release race kill switch resp=%+v err=%v", released, err)
	}
	assertP0KillSwitchRaceEvidence(t, ctx, db, contract.Id, makerID, takerID)
}

func waitP0KillSwitchActive(t *testing.T, ctx context.Context, db *sql.DB, userID int64) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var active int64
		err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_user_trading_control
			WHERE tenant_id=? AND user_id=? AND kill_switch=1`,
			p0AssetE2ETenantID, userID,
		).Scan(&active)
		if err == nil && active == 1 {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("kill switch did not become active while order cancellation was blocked")
}

func testP0CircuitBreakerBatchCancel(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	const (
		userID     int64 = 96201
		feeUserID  int64 = 96200
		operatorID int64 = 96299
		batchSize        = 101
	)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-CIRCUIT-BREAKER-BATCH-101", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, userID, "2000", "P0-CIRCUIT-BATCH-SEED")

	for index := 0; index < batchSize; index++ {
		placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
			AccountId: userID, ContractId: contract.Id,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
			ClientOrderId: fmt.Sprintf("P0-CIRCUIT-BATCH-%03d", index+1),
		})
	}
	processAssetInstructions(t, ctx, serviceCtx)

	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	refresh, err := adminlogic.NewUpdateMarketLogic(adminCtx, serviceCtx).UpdateMarket(&option.UpdateMarketReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id,
		UnderlyingPrice: "100", MarkPrice: "10", SnapshotTime: time.Now().Unix(),
	})
	if err != nil || refresh == nil || refresh.Base == nil || refresh.Base.Code != 200 {
		t.Fatalf("refresh pre-circuit market resp=%+v err=%v", refresh, err)
	}
	tripped, err := adminlogic.NewUpdateMarketLogic(adminCtx, serviceCtx).UpdateMarket(&option.UpdateMarketReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id,
		UnderlyingPrice: "100", MarkPrice: "15", SnapshotTime: time.Now().Unix(),
	})
	if err != nil || tripped == nil || tripped.Base == nil || tripped.Base.Code != 200 {
		t.Fatalf("trip circuit breaker resp=%+v err=%v", tripped, err)
	}
	assertP0CircuitBreakerBatchEvidence(t, ctx, db, contract.Id, userID, batchSize, false)
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0CircuitBreakerBatchEvidence(t, ctx, db, contract.Id, userID, batchSize, true)
}

func assertP0OrderStatus(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	orderID int64,
	want option.OrderStatus,
) {
	t.Helper()
	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
	if err != nil {
		t.Fatal(err)
	}
	if order.Status != int64(want) {
		t.Fatalf("order %d status=%d want=%d", orderID, order.Status, want)
	}
}

func assertP0KillSwitchReleaseEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, userID, sellerID, feeUserID, pendingID, partFilledID, fundingID int64,
) {
	t.Helper()
	var active, activated, released, rejected, trades, canceled, canceledFreeze int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_user_trading_control
		WHERE tenant_id=? AND user_id=? AND kill_switch=1`, p0AssetE2ETenantID, userID).
		Scan(&active); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(event_type='KILL_SWITCH_ACTIVATED'),0),
		COALESCE(SUM(event_type='KILL_SWITCH_RELEASED'),0),
		COALESCE(SUM(event_type='ORDER_REJECTED' AND reason='USER_KILL_SWITCH'),0)
		FROM t_option_trading_control_event WHERE tenant_id=? AND user_id=? AND contract_id IN (0,?)`,
		p0AssetE2ETenantID, userID, contractID,
	).Scan(&activated, &released, &rejected); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trade
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&trades); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(status=4),0),
		COALESCE(SUM(id=? AND status=4),0)
		FROM t_option_order WHERE tenant_id=? AND id IN (?,?,?)`,
		fundingID, p0AssetE2ETenantID, pendingID, partFilledID, fundingID,
	).Scan(&canceled, &canceledFreeze); err != nil {
		t.Fatal(err)
	}
	var fundingCanceledInstructions int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND order_id=? AND action=1 AND status=6`,
		p0AssetE2ETenantID, fundingID,
	).Scan(&fundingCanceledInstructions); err != nil {
		t.Fatal(err)
	}
	if active != 0 || activated != 1 || released != 1 || rejected != 1 || trades != 1 ||
		canceled != 3 || canceledFreeze != 1 || fundingCanceledInstructions != 1 {
		t.Fatalf("kill barrier active/activated/released/rejected/trades/canceled/funding_order/funding_instruction=%d/%d/%d/%d/%d/%d/%d/%d",
			active, activated, released, rejected, trades, canceled, canceledFreeze, fundingCanceledInstructions)
	}
	assertWalletAmounts(t, ctx, db, userID,
		"489.800000000000000000", "489.800000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, sellerID,
		"209.600000000000000000", "159.600000000000000000", "50.000000000000000000")
	assertWalletAmounts(t, ctx, db, feeUserID,
		"0.600000000000000000", "0.600000000000000000", "0.000000000000000000")
}

func assertP0KillSwitchRaceEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, makerID, takerID int64,
) {
	t.Helper()
	var trades, activeOrders, activated, released, instructions, success, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_trade WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_order WHERE tenant_id=? AND contract_id=? AND status IN (1,2,7,8,9)),
		(SELECT COUNT(*) FROM t_option_trading_control_event WHERE tenant_id=? AND user_id=? AND event_type='KILL_SWITCH_ACTIVATED'),
		(SELECT COUNT(*) FROM t_option_trading_control_event WHERE tenant_id=? AND user_id=? AND event_type='KILL_SWITCH_RELEASED'),
		COUNT(DISTINCT instruction.id),
		COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
		COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE orders.tenant_id=? AND orders.contract_id=?`,
		p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, makerID,
		p0AssetE2ETenantID, makerID,
		p0AssetE2ETenantID, contractID,
	).Scan(&trades, &activeOrders, &activated, &released, &instructions, &success, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if trades != 0 || activeOrders != 0 || activated != 1 || released != 1 ||
		instructions != 4 || success != 4 || reconciled != 4 || flows != 4 {
		t.Fatalf("kill race trades/active/activated/released/instructions/success/reconciled/flows=%d/%d/%d/%d/%d/%d/%d/%d",
			trades, activeOrders, activated, released, instructions, success, reconciled, flows)
	}
	assertWalletAmounts(t, ctx, db, makerID,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, takerID,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
}

func assertP0CircuitBreakerBatchEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, userID int64,
	batchSize int,
	settled bool,
) {
	t.Helper()
	var contractStatus, orders, canceling, canceled, trades, positions int64
	var haltTotal, haltSuccess, haltFailed, events, instructions, success, reconciled, flows int64
	var lastError string
	if err := db.QueryRowContext(ctx, `SELECT
		contract.status,
		(SELECT COUNT(*) FROM t_option_order orders WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id),
		(SELECT COUNT(*) FROM t_option_order orders WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND orders.status=8),
		(SELECT COUNT(*) FROM t_option_order orders WHERE orders.tenant_id=contract.tenant_id AND orders.contract_id=contract.id AND orders.status=4),
		halt.cancel_total,halt.cancel_success,halt.cancel_failed,halt.last_error_msg,
		(SELECT COUNT(*) FROM t_option_trading_control_event event WHERE event.tenant_id=contract.tenant_id AND event.contract_id=contract.id AND event.event_type='CIRCUIT_BREAKER'),
		(SELECT COUNT(*) FROM t_option_trade trade WHERE trade.tenant_id=contract.tenant_id AND trade.contract_id=contract.id),
		(SELECT COUNT(*) FROM t_option_position position WHERE position.tenant_id=contract.tenant_id AND position.contract_id=contract.id)
		FROM t_option_contract contract
		JOIN t_option_trading_halt halt ON halt.tenant_id=contract.tenant_id AND halt.contract_id=contract.id AND halt.status=1
		WHERE contract.tenant_id=? AND contract.id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&contractStatus, &orders, &canceling, &canceled, &haltTotal, &haltSuccess, &haltFailed,
		&lastError, &events, &trades, &positions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		COUNT(DISTINCT instruction.id),
		COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
		COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE orders.tenant_id=? AND orders.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&instructions, &success, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	wantCanceling, wantCanceled := int64(batchSize), int64(0)
	wantSuccess := int64(batchSize)
	if settled {
		wantCanceling, wantCanceled = 0, int64(batchSize)
		wantSuccess = int64(batchSize * 2)
	}
	if contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) ||
		orders != int64(batchSize) || canceling != wantCanceling || canceled != wantCanceled ||
		haltTotal != int64(batchSize) || haltSuccess != int64(batchSize) || haltFailed != 0 ||
		lastError != "" || events != 1 || trades != 0 || positions != 0 ||
		instructions != int64(batchSize*2) || success != wantSuccess || reconciled != wantSuccess || flows != wantSuccess {
		t.Fatalf("circuit batch settled=%t contract/orders/canceling/canceled/halt_total/halt_success/halt_failed/events/trades/positions/instructions/success/reconciled/flows=%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d last_error=%q",
			settled, contractStatus, orders, canceling, canceled, haltTotal, haltSuccess, haltFailed,
			events, trades, positions, instructions, success, reconciled, flows, lastError)
	}
	if settled {
		assertWalletAmounts(t, ctx, db, userID,
			"2000.000000000000000000", "2000.000000000000000000", "0.000000000000000000")
	}
}

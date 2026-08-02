package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

func testP0TradingCalendarGovernance(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	testP0TradingCalendarFutureSwitch(t, ctx, db, assetClient, serviceCtx)
	testP0ManualTradingHaltReleaseBarrier(t, ctx, db, assetClient, serviceCtx)
	testP0TradingHaltOrderAdmissionRace(t, ctx, db, serviceCtx)
}

type p0TradingCalendarPlaceResult struct {
	response *option.PlaceOrderResp
	err      error
}

type p0TradingCalendarHaltResult struct {
	response *option.GetTradingHaltResp
	err      error
}

func testP0TradingHaltOrderAdmissionRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		calendarCode = "P2_HALT_ORDER_RACE_24_7"
		rounds       = 20
	)
	now := time.Now().Unix()
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	accepted := 0
	rejected := 0
	for round := 0; round < rounds; round++ {
		userID := int64(96501 + round)
		operatorID := int64(96599)
		accountID := userID
		contract := insertP0OrderTestContract(
			t, ctx, serviceCtx, fmt.Sprintf("P2-HALT-ORDER-RACE-%02d", round),
			calendarCode, 96500, now,
		)
		insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
		position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
			TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: accountID,
			ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
			Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: decimal.NewFromInt(1),
			AvailableQty: decimal.NewFromInt(1), FrozenQty: decimal.Zero,
			OpenAvgPrice: decimal.NewFromInt(10), MarkPrice: decimal.NewFromInt(10),
			PositionValue: decimal.NewFromInt(10), ExerciseableQty: decimal.NewFromInt(1),
			Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
			CreateTimes: now, UpdateTimes: now,
		})
		clientOrderID := fmt.Sprintf("P2-HALT-ORDER-RACE-%02d", round)
		start := make(chan struct{})
		placeResult := make(chan p0TradingCalendarPlaceResult, 1)
		haltResult := make(chan p0TradingCalendarHaltResult, 1)
		var ready sync.WaitGroup
		ready.Add(2)
		go func() {
			ready.Done()
			<-start
			// Alternate the head start so both legal serializations of the
			// admission/halt race are exercised deterministically.
			if round%2 == 1 {
				time.Sleep(150 * time.Millisecond)
			}
			response, placeErr := applogic.NewPlaceOrderLogic(
				p0OrderUserContext(ctx, userID), serviceCtx,
			).PlaceOrder(&option.PlaceOrderReq{
				AccountId: accountID, ContractId: contract.Id,
				Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_CLOSE,
				OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
				ReduceOnly: common.YesNo_YES_NO_YES, ClientOrderId: clientOrderID,
			})
			placeResult <- p0TradingCalendarPlaceResult{response: response, err: placeErr}
		}()
		go func() {
			ready.Done()
			<-start
			if round%2 == 0 {
				time.Sleep(150 * time.Millisecond)
			}
			response, haltErr := adminlogic.NewHaltContractTradingLogic(
				p0AdminContext(ctx, operatorID, p0AssetE2ETenantID), serviceCtx,
			).HaltContractTrading(&option.HaltContractTradingReq{
				TenantId: p0AssetE2ETenantID, ContractId: contract.Id,
				Reason: "P2_HALT_ORDER_ADMISSION_RACE", EvidenceRef: "P2-HALT-RACE-EVIDENCE",
			})
			haltResult <- p0TradingCalendarHaltResult{response: response, err: haltErr}
		}()
		ready.Wait()
		close(start)

		var placed p0TradingCalendarPlaceResult
		select {
		case placed = <-placeResult:
		case <-time.After(10 * time.Second):
			t.Fatalf("round %d concurrent place order timed out", round)
		}
		var halted p0TradingCalendarHaltResult
		select {
		case halted = <-haltResult:
		case <-time.After(10 * time.Second):
			t.Fatalf("round %d concurrent trading halt timed out", round)
		}
		if halted.err != nil || halted.response == nil || halted.response.Base == nil ||
			halted.response.Base.Code != 200 || halted.response.Data == nil {
			t.Fatalf("round %d concurrent halt resp=%+v err=%v", round, halted.response, halted.err)
		}
		if placed.err != nil || placed.response == nil || placed.response.Base == nil {
			t.Fatalf("round %d concurrent place resp=%+v err=%v", round, placed.response, placed.err)
		}

		var orders, canceledOrders, clientKeys, instructions int64
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=?),0)
			FROM t_option_order WHERE tenant_id=? AND user_id=? AND contract_id=? AND client_order_id=?`,
			int64(option.OrderStatus_ORDER_STATUS_CANCELED), p0AssetE2ETenantID,
			userID, contract.Id, clientOrderID,
		).Scan(&orders, &canceledOrders); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_client_order_key
			WHERE tenant_id=? AND user_id=? AND client_order_id=?`,
			p0AssetE2ETenantID, userID, clientOrderID,
		).Scan(&clientKeys); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
			WHERE tenant_id=? AND order_id IN (
				SELECT id FROM t_option_order WHERE tenant_id=? AND user_id=?
				 AND contract_id=? AND client_order_id=?
			)`, p0AssetE2ETenantID, p0AssetE2ETenantID, userID, contract.Id, clientOrderID,
		).Scan(&instructions); err != nil {
			t.Fatal(err)
		}
		storedPosition, err := serviceCtx.OptionPositionModel.FindOne(ctx, position.Id)
		if err != nil || !storedPosition.AvailableQty.Equal(decimal.NewFromInt(1)) ||
			!storedPosition.FrozenQty.IsZero() {
			t.Fatalf("round %d position after race=%+v err=%v", round, storedPosition, err)
		}
		unsafe, err := serviceCtx.OptionOrderModel.HasUnsafeContractResumeOrders(
			ctx, p0AssetE2ETenantID, contract.Id,
		)
		if err != nil || unsafe {
			t.Fatalf("round %d unsafe orders remain=%v err=%v", round, unsafe, err)
		}

		if placed.response.Base.Code == 200 {
			accepted++
			if placed.response.Data == nil || placed.response.Data.OrderId <= 0 ||
				orders != 1 || canceledOrders != 1 || clientKeys != 1 || instructions != 0 ||
				halted.response.Data.CancelTotal != 1 || halted.response.Data.CancelSuccess != 1 ||
				halted.response.Data.CancelFailed != 0 {
				t.Fatalf("round %d accepted race evidence order=%+v halt=%+v counts=%d/%d/%d/%d",
					round, placed.response.Data, halted.response.Data,
					orders, canceledOrders, clientKeys, instructions)
			}
		} else {
			rejected++
			if orders != 0 || canceledOrders != 0 || clientKeys != 0 || instructions != 0 ||
				halted.response.Data.CancelTotal != 0 || halted.response.Data.CancelSuccess != 0 ||
				halted.response.Data.CancelFailed != 0 {
				t.Fatalf("round %d rejected race left side effects halt=%+v counts=%d/%d/%d/%d",
					round, halted.response.Data, orders, canceledOrders, clientKeys, instructions)
			}
		}

		resumed, err := adminlogic.NewResumeContractTradingLogic(
			p0AdminContext(ctx, operatorID, p0AssetE2ETenantID), serviceCtx,
		).ResumeContractTrading(&option.ResumeContractTradingReq{
			TenantId: p0AssetE2ETenantID, HaltId: halted.response.Data.Id,
			Reason: "P2_HALT_ORDER_RACE_RECONCILED",
		})
		if err != nil || resumed == nil || resumed.Base == nil || resumed.Base.Code != 200 {
			t.Fatalf("round %d resume after race resp=%+v err=%v", round, resumed, err)
		}
	}
	if accepted+rejected != rounds || accepted == 0 || rejected == 0 {
		t.Fatalf("halt/order race accepted/rejected/rounds=%d/%d/%d", accepted, rejected, rounds)
	}
	t.Logf("trading halt/order admission race accepted=%d rejected=%d rounds=%d", accepted, rejected, rounds)
}

func testP0TradingCalendarFutureSwitch(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		calendarCode       = "P2_CALENDAR_FUTURE_SWITCH"
		userID       int64 = 96301
		feeUserID    int64 = 96300
		creatorID    int64 = 96398
		reviewerID   int64 = 96399
	)
	now := time.Now().Unix()
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-CALENDAR-FUTURE-SWITCH", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, userID, "100", "P2-CALENDAR-SWITCH-SEED")

	switchAt := time.Now().Unix() + 4
	created, err := adminlogic.NewCreateTradingCalendarLogic(
		p0AdminContext(ctx, creatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateTradingCalendar(&option.CreateTradingCalendarReq{
		TenantId: p0AssetE2ETenantID, CalendarCode: calendarCode, Timezone: "UTC",
		EffectiveFrom: switchAt, ChangeReason: "P2_CALENDAR_FUTURE_SWITCH",
		EvidenceRef: "P2-CAL-E2E-V2",
		Sessions:    p0ContinuousTradingCalendarSessions(),
		Exceptions: []*option.TradingCalendarExceptionInput{{
			ExceptionType: option.TradingCalendarExceptionType_TRADING_CALENDAR_EXCEPTION_TYPE_CLOSED,
			StartTime:     switchAt, EndTime: switchAt + 3600,
			Reason: "P2_CALENDAR_SWITCH_CLOSED", AnnouncementRef: "P2-CAL-CLOSED-EVIDENCE",
		}},
	})
	if err != nil || created == nil || created.Base == nil || created.Base.Code != 200 ||
		created.Data == nil || created.Data.Version != 2 {
		t.Fatalf("create future calendar resp=%+v err=%v", created, err)
	}
	reviewed, err := adminlogic.NewReviewTradingCalendarLogic(
		p0AdminContext(ctx, reviewerID, p0AssetE2ETenantID), serviceCtx,
	).ReviewTradingCalendar(&option.ReviewTradingCalendarReq{
		TenantId: p0AssetE2ETenantID, CalendarId: created.Data.Id,
		Approve: true, Reason: "P2_CALENDAR_INDEPENDENT_REVIEW",
	})
	if err != nil || reviewed == nil || reviewed.Base == nil || reviewed.Base.Code != 200 ||
		reviewed.Data == nil || reviewed.Data.Status != option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_APPROVED {
		t.Fatalf("review future calendar resp=%+v err=%v", reviewed, err)
	}

	oldVersion, err := serviceCtx.OptionTradingCalendarModel.FindEffective(
		ctx, p0AssetE2ETenantID, calendarCode, switchAt-1,
	)
	if err != nil || oldVersion.Version != 1 || oldVersion.Status != int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_SUPERSEDED) ||
		oldVersion.EffectiveUntil != switchAt {
		t.Fatalf("calendar before switch=%+v err=%v", oldVersion, err)
	}
	newVersion, err := serviceCtx.OptionTradingCalendarModel.FindEffective(
		ctx, p0AssetE2ETenantID, calendarCode, switchAt,
	)
	if err != nil || newVersion.Id != created.Data.Id || newVersion.Version != 2 {
		t.Fatalf("calendar at switch=%+v err=%v", newVersion, err)
	}
	beforeDecision, err := logichelpers.IsContractTradingOpenWithModels(
		ctx, serviceCtx.OptionTradingHaltModel, serviceCtx.OptionTradingCalendarModel,
		serviceCtx.OptionTradingCalendarSessionModel, serviceCtx.OptionTradingCalendarExceptionModel,
		contract, switchAt-1,
	)
	if err != nil || beforeDecision == nil || !beforeDecision.Open || beforeDecision.Version != 1 {
		t.Fatalf("calendar decision before switch=%+v err=%v", beforeDecision, err)
	}
	atDecision, err := logichelpers.IsContractTradingOpenWithModels(
		ctx, serviceCtx.OptionTradingHaltModel, serviceCtx.OptionTradingCalendarModel,
		serviceCtx.OptionTradingCalendarSessionModel, serviceCtx.OptionTradingCalendarExceptionModel,
		contract, switchAt,
	)
	if err != nil || atDecision == nil || atDecision.Open || atDecision.Version != 2 ||
		atDecision.Reason != "CALENDAR_CLOSED_EXCEPTION" {
		t.Fatalf("calendar decision at switch=%+v err=%v", atDecision, err)
	}

	preSwitch := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P2-CALENDAR-BEFORE-SWITCH",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0OrderStatus(t, ctx, serviceCtx, preSwitch.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	waitP0UnixBoundary(t, ctx, switchAt)
	rejected, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).PlaceOrder(&option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P2-CALENDAR-AT-SWITCH-REJECTED",
	})
	if err != nil || rejected == nil || rejected.Base == nil || rejected.Base.Code == 200 {
		t.Fatalf("calendar did not reject order at new closed version resp=%+v err=%v", rejected, err)
	}
	assertP0UserCancelOK(t, ctx, serviceCtx, userID, userID, preSwitch.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, userID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertP0TradingCalendarSwitchEvidence(
		t, ctx, db, contract.Id, created.Data.Id, switchAt, userID,
	)
}

func testP0ManualTradingHaltReleaseBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		calendarCode       = "P2_MANUAL_HALT_24_7"
		userID       int64 = 96401
		feeUserID    int64 = 96400
		operatorID   int64 = 96499
	)
	now := time.Now().Unix()
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-MANUAL-HALT-RELEASE-BARRIER", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, userID, "200", "P2-MANUAL-HALT-SEED")

	funded := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P2-MANUAL-HALT-FUNDED",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	funding := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "8", Qty: "1",
		ClientOrderId: "P2-MANUAL-HALT-FUNDING",
	})
	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	halted, err := adminlogic.NewHaltContractTradingLogic(adminCtx, serviceCtx).
		HaltContractTrading(&option.HaltContractTradingReq{
			TenantId: p0AssetE2ETenantID, ContractId: contract.Id,
			Reason: "P2_MANUAL_HALT_E2E", EvidenceRef: "P2-HALT-EVIDENCE",
		})
	if err != nil || halted == nil || halted.Base == nil || halted.Base.Code != 200 ||
		halted.Data == nil || halted.Data.CancelTotal != 2 || halted.Data.CancelSuccess != 2 ||
		halted.Data.CancelFailed != 0 {
		t.Fatalf("manual trading halt resp=%+v err=%v", halted, err)
	}
	assertP0OrderStatus(t, ctx, serviceCtx, funded.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELING)
	assertP0OrderStatus(t, ctx, serviceCtx, funding.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)

	rejected, err := applogic.NewPlaceOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).PlaceOrder(&option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P2-MANUAL-HALT-REJECTED",
	})
	if err != nil || rejected == nil || rejected.Base == nil || rejected.Base.Code == 200 {
		t.Fatalf("manual halt did not reject new order resp=%+v err=%v", rejected, err)
	}
	earlyResume, err := adminlogic.NewResumeContractTradingLogic(adminCtx, serviceCtx).
		ResumeContractTrading(&option.ResumeContractTradingReq{
			TenantId: p0AssetE2ETenantID, HaltId: halted.Data.Id, Reason: "TOO_EARLY",
		})
	if err == nil || earlyResume != nil {
		t.Fatalf("manual halt resumed before releases resp=%+v err=%v", earlyResume, err)
	}

	fundedOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, funded.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	faultClient := &failOnceUnfreezeAssetClient{
		AssetClient: assetClient, targetBizNo: fundedOrder.OrderNo,
	}
	serviceCtx.AssetClient = faultClient
	defer func() { serviceCtx.AssetClient = assetClient }()
	processAssetInstructions(t, ctx, serviceCtx)
	serviceCtx.AssetClient = assetClient
	if faultClient.failureCount() != 1 {
		t.Fatalf("manual halt committed unfreeze response losses=%d want=1", faultClient.failureCount())
	}
	releaseInstructionID := assertP0TradingHaltReleaseFaultEvidence(
		t, ctx, db, fundedOrder.Id, fundedOrder.OrderNo,
	)
	responseLossResume, err := adminlogic.NewResumeContractTradingLogic(adminCtx, serviceCtx).
		ResumeContractTrading(&option.ResumeContractTradingReq{
			TenantId: p0AssetE2ETenantID, HaltId: halted.Data.Id,
			Reason: "ASSET_RESPONSE_LOSS_NOT_RECONCILED",
		})
	if err == nil || responseLossResume != nil {
		t.Fatalf("manual halt resumed after committed release response loss resp=%+v err=%v",
			responseLossResume, err)
	}
	retried, err := adminlogic.NewRetryAssetInstructionLogic(adminCtx, serviceCtx).
		RetryAssetInstruction(&option.RetryAssetInstructionReq{
			TenantId: p0AssetE2ETenantID, InstructionId: releaseInstructionID,
			Reason: "TRADING_HALT_ASSET_RESPONSE_LOSS_CONFIRMED",
		})
	if err != nil || retried == nil || retried.Base == nil || retried.Base.Code != 200 {
		t.Fatalf("retry manual halt release instruction resp=%+v err=%v", retried, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0TradingHaltReleaseRecoveredOnce(
		t, ctx, db, releaseInstructionID, fundedOrder.OrderNo,
	)
	assertP0OrderStatus(t, ctx, serviceCtx, funded.Data.OrderId, option.OrderStatus_ORDER_STATUS_CANCELED)
	assertWalletAmounts(t, ctx, db, userID,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
	resumed, err := adminlogic.NewResumeContractTradingLogic(adminCtx, serviceCtx).
		ResumeContractTrading(&option.ResumeContractTradingReq{
			TenantId: p0AssetE2ETenantID, HaltId: halted.Data.Id, Reason: "FUNDS_RECONCILED",
		})
	if err != nil || resumed == nil || resumed.Base == nil || resumed.Base.Code != 200 ||
		resumed.Data == nil || resumed.Data.Status != option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED {
		t.Fatalf("resume manual halt resp=%+v err=%v", resumed, err)
	}

	postResume := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
		AccountId: userID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P2-MANUAL-HALT-POST-RESUME",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0UserCancelOK(t, ctx, serviceCtx, userID, userID, postResume.Data.OrderId)
	processAssetInstructions(t, ctx, serviceCtx)
	assertWalletAmounts(t, ctx, db, userID,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")
	assertP0ManualTradingHaltEvidence(t, ctx, db, contract.Id, userID, halted.Data.Id)
}

func assertP0TradingHaltReleaseFaultEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	orderID int64,
	orderNo string,
) int64 {
	t.Helper()
	var instructionID, instructionStatus, reconciliationStatus, orderStatus, flowCount int64
	if err := db.QueryRowContext(ctx, `SELECT id,status,reconciliation_status
		FROM t_option_asset_instruction WHERE tenant_id=? AND instruction_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&instructionID, &instructionStatus, &reconciliationStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_order
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, orderID,
	).Scan(&orderStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&flowCount); err != nil {
		t.Fatal(err)
	}
	if instructionStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) ||
		reconciliationStatus != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING) ||
		orderStatus != int64(option.OrderStatus_ORDER_STATUS_CANCELING) || flowCount != 1 {
		t.Fatalf("halt response-loss instruction/reconciliation/order/flows=%d/%d/%d/%d",
			instructionStatus, reconciliationStatus, orderStatus, flowCount)
	}
	return instructionID
}

func assertP0TradingHaltReleaseRecoveredOnce(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	instructionID int64,
	orderNo string,
) {
	t.Helper()
	var instructionStatus, reconciliationStatus, flowCount, retryEvents int64
	if err := db.QueryRowContext(ctx, `SELECT status,reconciliation_status
		FROM t_option_asset_instruction WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, instructionID,
	).Scan(&instructionStatus, &reconciliationStatus); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&flowCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event
		WHERE tenant_id=? AND event_type='ASSET_INSTRUCTION_MANUAL_RETRY'
		 AND detail LIKE CONCAT('%instructionId=',?,'%')`,
		p0AssetE2ETenantID, instructionID,
	).Scan(&retryEvents); err != nil {
		t.Fatal(err)
	}
	if instructionStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) ||
		reconciliationStatus != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED) ||
		flowCount != 1 || retryEvents != 1 {
		t.Fatalf("halt recovered instruction/reconciliation/flows/retry_events=%d/%d/%d/%d",
			instructionStatus, reconciliationStatus, flowCount, retryEvents)
	}
}

func p0ContinuousTradingCalendarSessions() []*option.TradingCalendarSessionInput {
	sessions := make([]*option.TradingCalendarSessionInput, 0, 7)
	for weekday := int32(0); weekday < 7; weekday++ {
		sessions = append(sessions, &option.TradingCalendarSessionInput{
			Weekday: weekday, OpenSecond: 0, CloseSecond: 86400,
		})
	}
	return sessions
}

func waitP0UnixBoundary(t *testing.T, ctx context.Context, boundary int64) {
	t.Helper()
	for time.Now().Unix() < boundary {
		select {
		case <-ctx.Done():
			t.Fatalf("wait for trading calendar boundary: %v", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func assertP0TradingCalendarSwitchEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, newCalendarID, switchAt, userID int64,
) {
	t.Helper()
	var versions, approved, superseded, exactBoundary, orders, rejectedKeys, instructions, successful, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=2),SUM(status=4),
		SUM((status=4 AND effective_until=?) OR (status=2 AND effective_from=?))
		FROM t_option_trading_calendar WHERE tenant_id=? AND calendar_code='P2_CALENDAR_FUTURE_SWITCH'`,
		switchAt, switchAt, p0AssetE2ETenantID,
	).Scan(&versions, &approved, &superseded, &exactBoundary); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		SUM(client_order_id='P2-CALENDAR-AT-SWITCH-REJECTED')
		FROM t_option_order WHERE tenant_id=? AND user_id=? AND contract_id=?`,
		p0AssetE2ETenantID, userID, contractID,
	).Scan(&orders, &rejectedKeys); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=3),SUM(reconciliation_status=2),
		COUNT(DISTINCT asset_flow_no) FROM t_option_asset_instruction
		WHERE tenant_id=? AND order_id IN (
			SELECT id FROM t_option_order WHERE tenant_id=? AND contract_id=?
		)`, p0AssetE2ETenantID, p0AssetE2ETenantID, contractID,
	).Scan(&instructions, &successful, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if versions != 2 || approved != 1 || superseded != 1 || exactBoundary != 2 ||
		orders != 1 || rejectedKeys != 0 || instructions != 2 || successful != 2 ||
		reconciled != 2 || flows != 2 {
		t.Fatalf("calendar switch evidence versions/approved/superseded/boundary/orders/rejected/instructions/success/reconciled/flows=%d/%d/%d/%d/%d/%d/%d/%d/%d/%d",
			versions, approved, superseded, exactBoundary, orders, rejectedKeys,
			instructions, successful, reconciled, flows)
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_trading_calendar
		SET evidence_ref='TAMPERED' WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, newCalendarID); err == nil {
		t.Fatal("approved trading calendar evidence mutation was not rejected")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_trading_calendar
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, newCalendarID); err == nil {
		t.Fatal("approved trading calendar deletion was not rejected")
	}
}

func assertP0ManualTradingHaltEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, userID, haltID int64,
) {
	t.Helper()
	var contractStatus, haltStatus, cancelTotal, cancelSuccess, cancelFailed int64
	var activeHalts, orders, canceledOrders, instructions, successful, canceled, reconciled, flows, events int64
	if err := db.QueryRowContext(ctx, `SELECT contract.status,halt.status,halt.cancel_total,
		halt.cancel_success,halt.cancel_failed
		FROM t_option_contract contract JOIN t_option_trading_halt halt
		 ON halt.tenant_id=contract.tenant_id AND halt.contract_id=contract.id
		WHERE contract.tenant_id=? AND contract.id=? AND halt.id=?`,
		p0AssetE2ETenantID, contractID, haltID,
	).Scan(&contractStatus, &haltStatus, &cancelTotal, &cancelSuccess, &cancelFailed); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT SUM(active_key=CONCAT('CONTRACT:',contract_id))
		FROM t_option_trading_halt WHERE tenant_id=? AND contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&activeHalts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=?)
		FROM t_option_order WHERE tenant_id=? AND user_id=? AND contract_id=?`,
		int64(option.OrderStatus_ORDER_STATUS_CANCELED), p0AssetE2ETenantID, userID, contractID,
	).Scan(&orders, &canceledOrders); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=3),SUM(status=6),
		SUM(reconciliation_status=2),COUNT(DISTINCT NULLIF(asset_flow_no,''))
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (
			SELECT id FROM t_option_order WHERE tenant_id=? AND user_id=? AND contract_id=?
		)`, p0AssetE2ETenantID, p0AssetE2ETenantID, userID, contractID,
	).Scan(&instructions, &successful, &canceled, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=?
		 AND event_type IN ('CONTRACT_TRADING_HALTED','CONTRACT_TRADING_RESUMED')`,
		p0AssetE2ETenantID, contractID,
	).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) ||
		haltStatus != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED) ||
		cancelTotal != 2 || cancelSuccess != 2 || cancelFailed != 0 || activeHalts != 0 ||
		orders != 3 || canceledOrders != 3 || instructions != 5 || successful != 4 || canceled != 1 ||
		reconciled != 4 || flows != 4 || events != 2 {
		t.Fatalf("manual halt evidence contract/halt/cancel/active/orders/canceled_orders/instructions/success/canceled/reconciled/flows/events=%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d/%d",
			contractStatus, haltStatus, cancelTotal, cancelSuccess, cancelFailed, activeHalts,
			orders, canceledOrders, instructions, successful, canceled, reconciled, flows, events)
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_trading_halt SET reason='TAMPERED'
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, haltID); err == nil {
		t.Fatal("trading halt identity mutation was not rejected")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_trading_halt
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, haltID); err == nil {
		t.Fatal("trading halt deletion was not rejected")
	}
}

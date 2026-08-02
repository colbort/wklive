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
	"wklive/services/option/internal/svc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type failOnceUnfreezeAssetClient struct {
	asset.AssetClient

	mu          sync.Mutex
	targetBizNo string
	failures    int
}

func (c *failOnceUnfreezeAssetClient) UnfreezeAssetByBizNo(
	ctx context.Context,
	in *asset.UnfreezeAssetByBizNoReq,
	opts ...grpc.CallOption,
) (*asset.ChangeAssetResp, error) {
	c.mu.Lock()
	shouldFail := in.TargetBizNo == c.targetBizNo && c.failures == 0
	if shouldFail {
		c.failures++
	}
	c.mu.Unlock()
	if !shouldFail {
		return c.AssetClient.UnfreezeAssetByBizNo(ctx, in, opts...)
	}
	if _, err := c.AssetClient.UnfreezeAssetByBizNo(ctx, in, opts...); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unavailable, "P1 MMP injected unfreeze response loss after commit")
}

func (c *failOnceUnfreezeAssetClient) failureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

func testP1MMPAssetRPC(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		makerID      int64 = 96301
		takerID      int64 = 96302
		feeUserID    int64 = 96300
		operatorID   int64 = 96399
		makerCount         = 102
		primaryGroup       = "desk-a"
		otherGroup         = "desk-b"
	)
	now := time.Now().Unix()
	calendarCode := "P1_MMP_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	contract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P1-MMP-REAL-ASSET-BATCH-102", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, makerID, "3000", "P1-MMP-MAKER-SEED")
	creditAsset(t, ctx, assetClient, takerID, "3000", "P1-MMP-TAKER-SEED")
	creditAsset(t, ctx, assetClient, feeUserID, "1", "P1-MMP-FEE-SEED")

	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	upsertP1MMPConfig(t, adminCtx, serviceCtx, makerID, contract.Id, primaryGroup, 1)
	upsertP1MMPConfig(t, adminCtx, serviceCtx, makerID, contract.Id, otherGroup, 1000)

	primaryOrders := make([]*option.PlaceOrderResp, 0, makerCount)
	for index := 0; index < makerCount; index++ {
		primaryOrders = append(primaryOrders, placeP0Order(t, ctx, serviceCtx, makerID, &option.PlaceOrderReq{
			AccountId: makerID, ContractId: contract.Id,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "9", Qty: "1",
			ClientOrderId: fmt.Sprintf("P1-MMP-PRIMARY-%03d", index+1),
			Mmp:           common.YesNo_YES_NO_YES, MmpGroup: primaryGroup,
		}))
	}
	other := placeP0Order(t, ctx, serviceCtx, makerID, &option.PlaceOrderReq{
		AccountId: makerID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "8", Qty: "1",
		ClientOrderId: "P1-MMP-OTHER-GROUP",
		Mmp:           common.YesNo_YES_NO_YES, MmpGroup: otherGroup,
	})
	ordinary := placeP0Order(t, ctx, serviceCtx, makerID, &option.PlaceOrderReq{
		AccountId: makerID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "8", Qty: "1",
		ClientOrderId: "P1-MMP-ORDINARY-ISOLATION",
	})
	for attempt := 0; attempt < 3; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1MMPOrderCounts(t, ctx, db, contract.Id, primaryGroup, makerCount, 0, makerCount, 0)
	assertP0OrderStatus(t, ctx, serviceCtx, other.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	assertP0OrderStatus(t, ctx, serviceCtx, ordinary.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)

	taker := placeP0Order(t, ctx, serviceCtx, takerID, &option.PlaceOrderReq{
		AccountId: takerID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "9", Qty: "1",
		ClientOrderId: "P1-MMP-TRIGGER-TAKER",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0OrderStatus(t, ctx, serviceCtx, taker.Data.OrderId, option.OrderStatus_ORDER_STATUS_FILLED)
	assertP1MMPOrderCounts(t, ctx, db, contract.Id, primaryGroup, makerCount, 1, 0, makerCount-1)
	assertP0OrderStatus(t, ctx, serviceCtx, other.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	assertP0OrderStatus(t, ctx, serviceCtx, ordinary.Data.OrderId, option.OrderStatus_ORDER_STATUS_PENDING)
	assertP1MMPConfigStatus(t, ctx, db, makerID, contract.Id, primaryGroup, option.MMPStatus_MMP_STATUS_TRIGGERED)

	if _, err := db.ExecContext(ctx, `UPDATE t_option_mmp_config
		SET cooldown_until=UNIX_TIMESTAMP()-1
		WHERE tenant_id=? AND user_id=? AND contract_id=? AND group_code=?`,
		p0AssetE2ETenantID, makerID, contract.Id, primaryGroup,
	); err != nil {
		t.Fatal(err)
	}
	assertP1MMPPlaceRejected(t, ctx, serviceCtx, makerID, contract.Id, primaryGroup, "AUTO-RELEASE-BARRIER")
	assertP1MMPManualResetBlocked(t, adminCtx, serviceCtx, makerID, contract.Id, primaryGroup, "MANUAL-RELEASE-BARRIER")
	assertP1MMPConfigStatus(t, ctx, db, makerID, contract.Id, primaryGroup, option.MMPStatus_MMP_STATUS_TRIGGERED)

	faultTarget := primaryOrders[len(primaryOrders)-1]
	faultOrder, err := serviceCtx.OptionOrderModel.FindOne(ctx, faultTarget.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	faultClient := &failOnceUnfreezeAssetClient{
		AssetClient: assetClient, targetBizNo: faultOrder.OrderNo,
	}
	serviceCtx.AssetClient = faultClient
	processAssetInstructions(t, ctx, serviceCtx)
	serviceCtx.AssetClient = assetClient
	if faultClient.failureCount() != 1 {
		t.Fatalf("MMP committed unfreeze response losses=%d want=1", faultClient.failureCount())
	}
	faultInstructionID := assertP1MMPReleaseFaultEvidence(
		t, ctx, db, contract.Id, faultOrder.Id, faultOrder.OrderNo,
	)
	assertP1MMPManualResetBlocked(t, adminCtx, serviceCtx, makerID, contract.Id, primaryGroup, "FAULT-RELEASE-BARRIER")
	assertP1MMPPlaceRejected(t, ctx, serviceCtx, makerID, contract.Id, primaryGroup, "FAULT-AUTO-BARRIER")

	retried, err := adminlogic.NewRetryAssetInstructionLogic(adminCtx, serviceCtx).
		RetryAssetInstruction(&option.RetryAssetInstructionReq{
			TenantId: p0AssetE2ETenantID, InstructionId: faultInstructionID,
			Reason: "MMP_ASSET_RESPONSE_LOSS_CONFIRMED",
		})
	if err != nil || retried == nil || retried.Base == nil || retried.Base.Code != 200 {
		t.Fatalf("retry MMP release instruction resp=%+v err=%v", retried, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0OrderStatus(t, ctx, serviceCtx, faultOrder.Id, option.OrderStatus_ORDER_STATUS_CANCELED)
	assertP1MMPReleaseRecoveredOnce(t, ctx, db, faultInstructionID, faultOrder.OrderNo)

	type resetResult struct {
		response *option.GetMMPConfigResp
		err      error
	}
	results := make(chan resetResult, 20)
	var wait sync.WaitGroup
	for index := 0; index < 20; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			response, resetErr := adminlogic.NewResetMMPConfigLogic(adminCtx, serviceCtx).
				ResetMMPConfig(&option.ResetMMPConfigReq{
					TenantId: p0AssetE2ETenantID, UserId: makerID,
					ContractId: contract.Id, GroupCode: primaryGroup,
					Reason: fmt.Sprintf("MMP_RECONCILED_REVIEW_%02d", index),
				})
			results <- resetResult{response: response, err: resetErr}
		}(index)
	}
	wait.Wait()
	close(results)
	successfulResets := 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent MMP reset transport error: %v", result.err)
		}
		if result.response != nil && result.response.Base != nil && result.response.Base.Code == 200 {
			successfulResets++
		}
	}
	if successfulResets != 1 {
		t.Fatalf("concurrent successful MMP resets=%d want=1", successfulResets)
	}
	assertP1MMPConfigStatus(t, ctx, db, makerID, contract.Id, primaryGroup, option.MMPStatus_MMP_STATUS_ACTIVE)

	postReset := placeP0Order(t, ctx, serviceCtx, makerID, &option.PlaceOrderReq{
		AccountId: makerID, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "8", Qty: "1",
		ClientOrderId: "P1-MMP-POST-RESET",
		Mmp:           common.YesNo_YES_NO_YES, MmpGroup: primaryGroup,
	})
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0UserCancelOK(t, ctx, serviceCtx, makerID, makerID, postReset.Data.OrderId)
	assertP0UserCancelOK(t, ctx, serviceCtx, makerID, makerID, other.Data.OrderId)
	assertP0UserCancelOK(t, ctx, serviceCtx, makerID, makerID, ordinary.Data.OrderId)
	processP0TradeEvents(t, ctx, serviceCtx)
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1MMPEvidence(t, ctx, db, contract.Id, makerID, primaryGroup, otherGroup, makerCount)
}

func upsertP1MMPConfig(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, contractID int64,
	groupCode string,
	tradeCountThreshold int64,
) {
	t.Helper()
	response, err := adminlogic.NewUpsertMMPConfigLogic(ctx, serviceCtx).
		UpsertMMPConfig(&option.UpsertMMPConfigReq{
			TenantId: p0AssetE2ETenantID, UserId: userID, ContractId: contractID,
			GroupCode: groupCode, Enabled: common.YesNo_YES_NO_YES,
			QtyThreshold: "0", TradeCountThreshold: tradeCountThreshold,
			LossThreshold: "0", WindowSeconds: 60, CooldownSeconds: 1,
			Reason: "P1_MMP_REAL_ASSET_ACCEPTANCE",
		})
	if err != nil || response == nil || response.Base == nil || response.Base.Code != 200 {
		t.Fatalf("upsert MMP group=%s resp=%+v err=%v", groupCode, response, err)
	}
}

func assertP1MMPPlaceRejected(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, contractID int64,
	groupCode, suffix string,
) {
	t.Helper()
	response, err := applogic.NewPlaceOrderLogic(p0OrderUserContext(ctx, userID), serviceCtx).
		PlaceOrder(&option.PlaceOrderReq{
			AccountId: userID, ContractId: contractID,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_POST_ONLY, Price: "8", Qty: "1",
			ClientOrderId: "P1-MMP-REJECTED-" + suffix,
			Mmp:           common.YesNo_YES_NO_YES, MmpGroup: groupCode,
		})
	if err != nil || response == nil || response.Base == nil || response.Base.Code == 200 ||
		response.Base.Msg != "MMP_TRIGGERED" {
		t.Fatalf("MMP order crossed recovery barrier resp=%+v err=%v", response, err)
	}
}

func assertP1MMPManualResetBlocked(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID, contractID int64,
	groupCode, reason string,
) {
	t.Helper()
	response, err := adminlogic.NewResetMMPConfigLogic(ctx, serviceCtx).
		ResetMMPConfig(&option.ResetMMPConfigReq{
			TenantId: p0AssetE2ETenantID, UserId: userID, ContractId: contractID,
			GroupCode: groupCode, Reason: reason,
		})
	if err != nil || response == nil || response.Base == nil || response.Base.Code == 200 {
		t.Fatalf("MMP manual reset crossed release barrier resp=%+v err=%v", response, err)
	}
}

func assertP1MMPConfigStatus(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	userID, contractID int64,
	groupCode string,
	want option.MMPStatus,
) {
	t.Helper()
	var statusValue int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_mmp_config
		WHERE tenant_id=? AND user_id=? AND contract_id=? AND group_code=?`,
		p0AssetE2ETenantID, userID, contractID, groupCode,
	).Scan(&statusValue); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(want) {
		t.Fatalf("MMP group=%s status=%d want=%d", groupCode, statusValue, want)
	}
}

func assertP1MMPOrderCounts(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	groupCode string,
	totalWant, filledWant, pendingWant, cancelingWant int,
) {
	t.Helper()
	var total, filled, pending, canceling int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		COALESCE(SUM(status=?),0),COALESCE(SUM(status=?),0),COALESCE(SUM(status=?),0)
		FROM t_option_order WHERE tenant_id=? AND contract_id=? AND mmp=1 AND mmp_group=?`,
		int64(option.OrderStatus_ORDER_STATUS_FILLED),
		int64(option.OrderStatus_ORDER_STATUS_PENDING),
		int64(option.OrderStatus_ORDER_STATUS_CANCELING),
		p0AssetE2ETenantID, contractID, groupCode,
	).Scan(&total, &filled, &pending, &canceling); err != nil {
		t.Fatal(err)
	}
	if total != totalWant || filled != filledWant || pending != pendingWant || canceling != cancelingWant {
		t.Fatalf("MMP group=%s total/filled/pending/canceling=%d/%d/%d/%d want=%d/%d/%d/%d",
			groupCode, total, filled, pending, canceling,
			totalWant, filledWant, pendingWant, cancelingWant,
		)
	}
}

func assertP1MMPReleaseFaultEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, orderID int64,
	orderNo string,
) int64 {
	t.Helper()
	var instructionID, statusValue, orderStatus, flowCount int64
	if err := db.QueryRowContext(ctx, `SELECT id,status FROM t_option_asset_instruction
		WHERE tenant_id=? AND instruction_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&instructionID, &statusValue); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) {
		t.Fatalf("MMP fault instruction status=%d want FAILED", statusValue)
	}
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_order
		WHERE tenant_id=? AND contract_id=? AND id=?`,
		p0AssetE2ETenantID, contractID, orderID,
	).Scan(&orderStatus); err != nil {
		t.Fatal(err)
	}
	if orderStatus != int64(option.OrderStatus_ORDER_STATUS_CANCELING) {
		t.Fatalf("MMP response-loss order status=%d want CANCELING", orderStatus)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&flowCount); err != nil {
		t.Fatal(err)
	}
	if flowCount != 1 {
		t.Fatalf("MMP committed release flows=%d want=1 contract=%d", flowCount, contractID)
	}
	return instructionID
}

func assertP1MMPReleaseRecoveredOnce(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	instructionID int64,
	orderNo string,
) {
	t.Helper()
	var statusValue, reconciled, flowCount int64
	if err := db.QueryRowContext(ctx, `SELECT status,reconciliation_status
		FROM t_option_asset_instruction WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, instructionID,
	).Scan(&statusValue, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_no=?`,
		p0AssetE2ETenantID, orderNo+"-CONTROL-RELEASE",
	).Scan(&flowCount); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) ||
		reconciled != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED) ||
		flowCount != 1 {
		t.Fatalf("MMP recovered instruction status/reconciled/flows=%d/%d/%d want=%d/%d/1",
			statusValue, reconciled, flowCount,
			option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS,
			option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED,
		)
	}
}

func assertP1MMPEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, makerID int64,
	primaryGroup, otherGroup string,
	makerCount int,
) {
	t.Helper()
	var primaryCanceled, otherCanceled, ordinaryCanceled int
	if err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(mmp=1 AND mmp_group=? AND status=?),0),
		COALESCE(SUM(mmp=1 AND mmp_group=? AND status=?),0),
		COALESCE(SUM(mmp<>1 AND status=?),0)
		FROM t_option_order WHERE tenant_id=? AND contract_id=? AND user_id=?`,
		primaryGroup, int64(option.OrderStatus_ORDER_STATUS_CANCELED),
		otherGroup, int64(option.OrderStatus_ORDER_STATUS_CANCELED),
		int64(option.OrderStatus_ORDER_STATUS_CANCELED),
		p0AssetE2ETenantID, contractID, makerID,
	).Scan(&primaryCanceled, &otherCanceled, &ordinaryCanceled); err != nil {
		t.Fatal(err)
	}
	var releaseInstructions, releaseSuccess, releaseFlows int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(instruction.status=?),0),
		COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order orders ON orders.tenant_id=instruction.tenant_id AND orders.id=instruction.order_id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		WHERE instruction.tenant_id=? AND orders.contract_id=? AND orders.mmp=1
		  AND orders.mmp_group=? AND instruction.instruction_no LIKE '%-CONTROL-RELEASE'`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, contractID, primaryGroup,
	).Scan(&releaseInstructions, &releaseSuccess, &releaseFlows); err != nil {
		t.Fatal(err)
	}
	var triggeredEvents, canceledEvents, resetEvents int
	if err := db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(event_type='MMP_TRIGGERED'),0),
		COALESCE(SUM(event_type='MMP_ORDER_CANCELED'),0),
		COALESCE(SUM(event_type='MMP_MANUAL_RESET'),0)
		FROM t_option_trading_control_event
		WHERE tenant_id=? AND user_id=? AND contract_id=?`,
		p0AssetE2ETenantID, makerID, contractID,
	).Scan(&triggeredEvents, &canceledEvents, &resetEvents); err != nil {
		t.Fatal(err)
	}
	if primaryCanceled != makerCount || otherCanceled != 1 || ordinaryCanceled != 1 ||
		releaseInstructions != makerCount-1 || releaseSuccess != makerCount-1 ||
		releaseFlows != makerCount-1 || triggeredEvents != 1 || canceledEvents != 1 || resetEvents != 1 {
		t.Fatalf("MMP evidence primary/other/ordinary/releases/success/flows/events=%d/%d/%d/%d/%d/%d/%d/%d/%d",
			primaryCanceled, otherCanceled, ordinaryCanceled,
			releaseInstructions, releaseSuccess, releaseFlows,
			triggeredEvents, canceledEvents, resetEvents,
		)
	}
}

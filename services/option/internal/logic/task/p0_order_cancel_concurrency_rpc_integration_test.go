package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"google.golang.org/grpc/metadata"
)

func testP0AdminForceCancelAndFundingRace(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_ADMIN_CANCEL_RACE_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	adminContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-ADMIN-FORCE-CANCEL-CALL", calendarCode, 194, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, adminContract.Id, "100", "10", now)
	testP0AdminForceCancelAudit(
		t, ctx, db, assetClient, serviceCtx, adminContract,
	)

	raceContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-CANCEL-FUNDING-RACE-CALL", calendarCode, 195, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, raceContract.Id, "100", "10", now)
	testP0ConcurrentCancelAndFunding(
		t, ctx, db, assetClient, serviceCtx, raceContract,
	)
}

func testP0AdminForceCancelAudit(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		fundedUserID     int64 = 191
		preFundingUserID int64 = 192
		operatorID       int64 = 9003
	)
	creditAsset(t, ctx, assetClient, fundedUserID, "100", "P0-ADMIN-CANCEL-FUNDED-SEED")
	creditAsset(t, ctx, assetClient, preFundingUserID, "100", "P0-ADMIN-CANCEL-PRE-FUNDING-SEED")

	funded := placeP0Order(t, ctx, serviceCtx, fundedUserID, &option.PlaceOrderReq{
		AccountId: 7066, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-ADMIN-CANCEL-FUNDED",
	})
	processAssetInstructions(t, ctx, serviceCtx)
	preFunding := placeP0Order(t, ctx, serviceCtx, preFundingUserID, &option.PlaceOrderReq{
		AccountId: 7067, ContractId: contract.Id,
		Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P0-ADMIN-CANCEL-PRE-FUNDING",
	})

	contract.Status = int64(option.ContractStatus_CONTRACT_STATUS_PAUSED)
	contract.UpdateTimes = time.Now().Unix()
	if err := serviceCtx.OptionContractModel.Update(ctx, contract); err != nil {
		t.Fatalf("pause admin-cancel contract: %v", err)
	}

	unauthorized, err := adminlogic.NewForceCancelContractOrdersLogic(
		p0AdminContext(ctx, operatorID, p0AssetE2ETenantID+1), serviceCtx,
	).ForceCancelContractOrders(&option.ForceCancelContractOrdersReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id, Reason: "P0_ADMIN_FORCE_CANCEL",
	})
	if err != nil || unauthorized == nil || unauthorized.Base == nil || unauthorized.Base.Code == 200 {
		t.Fatalf("cross-tenant force cancel was not rejected resp=%+v err=%v", unauthorized, err)
	}

	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	resp, err := adminlogic.NewForceCancelContractOrdersLogic(
		adminCtx, serviceCtx,
	).ForceCancelContractOrders(&option.ForceCancelContractOrdersReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id, Reason: "P0_ADMIN_FORCE_CANCEL",
	})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("admin force cancel resp=%+v err=%v", resp, err)
	}
	for i := 0; i < 3; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP0AdminForceCancelEvidence(
		t, ctx, db, serviceCtx, contract.Id, funded.Data.OrderId, preFunding.Data.OrderId,
		fundedUserID, preFundingUserID, operatorID,
	)

	replay, err := adminlogic.NewForceCancelContractOrdersLogic(
		adminCtx, serviceCtx,
	).ForceCancelContractOrders(&option.ForceCancelContractOrdersReq{
		TenantId: p0AssetE2ETenantID, ContractId: contract.Id, Reason: "P0_ADMIN_FORCE_CANCEL",
	})
	if err != nil || replay == nil || replay.Base == nil || replay.Base.Code != 200 {
		t.Fatalf("admin force-cancel replay resp=%+v err=%v", replay, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0AdminForceCancelEvidence(
		t, ctx, db, serviceCtx, contract.Id, funded.Data.OrderId, preFunding.Data.OrderId,
		fundedUserID, preFundingUserID, operatorID,
	)

	if _, err := db.ExecContext(ctx, `UPDATE t_option_trading_control_event
		SET reason='TAMPERED' WHERE tenant_id=? AND contract_id=?
		AND event_type='ADMIN_FORCE_CANCEL_ORDER'`, p0AssetE2ETenantID, contract.Id); err == nil {
		t.Fatal("admin force-cancel audit event update unexpectedly succeeded")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=? AND event_type='ADMIN_FORCE_CANCEL_ORDER'`,
		p0AssetE2ETenantID, contract.Id); err == nil {
		t.Fatal("admin force-cancel audit event delete unexpectedly succeeded")
	}
}

func assertP0AdminForceCancelEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractID, fundedOrderID, preFundingOrderID, fundedUserID, preFundingUserID, operatorID int64,
) {
	t.Helper()
	for _, orderID := range []int64{fundedOrderID, preFundingOrderID} {
		order, err := serviceCtx.OptionOrderModel.FindOne(ctx, orderID)
		if err != nil {
			t.Fatal(err)
		}
		if order.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
			order.CancelReason != "P0_ADMIN_FORCE_CANCEL" || !order.MarginAmount.IsZero() {
			t.Fatalf("unexpected admin-canceled order: %+v", order)
		}
	}
	var instructions, success, canceled, reconciled, flows, events, eventOperators, reasons int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=3),0),
		COALESCE(SUM(status=6),0),COALESCE(SUM(reconciliation_status=2),0)
		FROM t_option_asset_instruction WHERE tenant_id=? AND order_id IN (?,?)`,
		p0AssetE2ETenantID, fundedOrderID, preFundingOrderID,
	).Scan(&instructions, &success, &canceled, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction JOIN t_asset_flow flow
		 ON flow.tenant_id=instruction.tenant_id
		AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE instruction.tenant_id=? AND instruction.order_id IN (?,?)`,
		p0AssetE2ETenantID, fundedOrderID, preFundingOrderID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(operator_id=?),0),
		COALESCE(SUM(reason='P0_ADMIN_FORCE_CANCEL'),0)
		FROM t_option_trading_control_event WHERE tenant_id=? AND contract_id=?
		AND event_type='ADMIN_FORCE_CANCEL_ORDER'`,
		operatorID, p0AssetE2ETenantID, contractID,
	).Scan(&events, &eventOperators, &reasons); err != nil {
		t.Fatal(err)
	}
	if instructions != 3 || success != 2 || canceled != 1 || reconciled != 2 || flows != 2 ||
		events != 2 || eventOperators != 2 || reasons != 2 {
		t.Fatalf("admin cancel instructions/success/canceled/reconciled/flows/events/operators/reasons=%d/%d/%d/%d/%d/%d/%d/%d",
			instructions, success, canceled, reconciled, flows, events, eventOperators, reasons)
	}
	assertWalletAmounts(t, ctx, db, fundedUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletAmounts(t, ctx, db, preFundingUserID,
		"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}

func testP0ConcurrentCancelAndFunding(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
) {
	t.Helper()
	const (
		userID    int64 = 193
		accountID int64 = 7068
		rounds          = 20
	)
	creditAsset(t, ctx, assetClient, userID, "100", "P0-CANCEL-FUNDING-RACE-SEED")
	orderIDs := make([]int64, 0, rounds)
	for round := 0; round < rounds; round++ {
		placed := placeP0Order(t, ctx, serviceCtx, userID, &option.PlaceOrderReq{
			AccountId: accountID, ContractId: contract.Id,
			Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
			OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
			ClientOrderId: fmt.Sprintf("P0-CANCEL-FUNDING-RACE-%02d", round),
		})
		orderIDs = append(orderIDs, placed.Data.OrderId)

		start := make(chan struct{})
		taskResults := make(chan p0TaskCallResult, 1)
		cancelResults := make(chan p0CancelCallResult, 2)
		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			<-start
			resp, err := NewProcessAssetInstructionsLogic(ctx, serviceCtx).ProcessAssetInstructions(
				&option.OptionTaskReq{TenantId: p0AssetE2ETenantID},
			)
			code := int64(0)
			if resp != nil && resp.Base != nil {
				code = int64(resp.Base.Code)
			}
			taskResults <- p0TaskCallResult{code: code, err: err}
		}()
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				<-start
				resp, err := applogic.NewCancelOrderLogic(
					p0OrderUserContext(ctx, userID), serviceCtx,
				).CancelOrder(&option.CancelOrderReq{
					AccountId: accountID, OrderId: placed.Data.OrderId,
				})
				code := int64(0)
				if resp != nil && resp.Base != nil {
					code = int64(resp.Base.Code)
				}
				cancelResults <- p0CancelCallResult{code: code, err: err}
			}()
		}
		close(start)
		wg.Wait()
		close(taskResults)
		close(cancelResults)

		taskResult := <-taskResults
		if taskResult.err != nil || taskResult.code != 200 {
			t.Fatalf("round %d funding task code=%d err=%v", round, taskResult.code, taskResult.err)
		}
		cancelSuccess, cancelRejected := 0, 0
		for result := range cancelResults {
			if result.err != nil {
				t.Fatalf("round %d concurrent cancel error: %v", round, result.err)
			}
			if result.code == 200 {
				cancelSuccess++
			} else {
				cancelRejected++
			}
		}
		if cancelSuccess != 1 || cancelRejected != 1 {
			t.Fatalf("round %d cancel success/rejected=%d/%d", round, cancelSuccess, cancelRejected)
		}
		for i := 0; i < 3; i++ {
			processAssetInstructions(t, ctx, serviceCtx)
		}
		order, err := serviceCtx.OptionOrderModel.FindOne(ctx, placed.Data.OrderId)
		if err != nil {
			t.Fatal(err)
		}
		if order.Status != int64(option.OrderStatus_ORDER_STATUS_CANCELED) ||
			order.CancelReason != "USER_CANCEL" || !order.MarginAmount.IsZero() {
			t.Fatalf("round %d unexpected raced order: %+v", round, order)
		}
		assertWalletAmounts(t, ctx, db, userID,
			"100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	}
	assertP0CancelFundingRaceEvidence(t, ctx, db, contract.Id, orderIDs)
}

type p0TaskCallResult struct {
	code int64
	err  error
}

type p0CancelCallResult struct {
	code int64
	err  error
}

func assertP0CancelFundingRaceEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	orderIDs []int64,
) {
	t.Helper()
	if len(orderIDs) != 20 {
		t.Fatalf("unexpected race order IDs: %d", len(orderIDs))
	}
	var orders, canceledOrders, clientKeys, instructions, success, canceled, reconciled, flows, duplicateInstructionNos int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status=4),0)
		FROM t_option_order WHERE tenant_id=? AND contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&orders, &canceledOrders); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_client_order_key key_item
		JOIN t_option_order o ON o.tenant_id=key_item.tenant_id AND o.id=key_item.order_id
		WHERE o.tenant_id=? AND o.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&clientKeys); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(instruction.status=3),0),
		COALESCE(SUM(instruction.status=6),0),COALESCE(SUM(instruction.reconciliation_status=2),0)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order o ON o.tenant_id=instruction.tenant_id AND o.id=instruction.order_id
		WHERE o.tenant_id=? AND o.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&instructions, &success, &canceled, &reconciled); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order o ON o.tenant_id=instruction.tenant_id AND o.id=instruction.order_id
		JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1 THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE o.tenant_id=? AND o.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM (
		SELECT instruction.instruction_no FROM t_option_asset_instruction instruction
		JOIN t_option_order o ON o.tenant_id=instruction.tenant_id AND o.id=instruction.order_id
		WHERE o.tenant_id=? AND o.contract_id=?
		GROUP BY instruction.instruction_no HAVING COUNT(*)>1
	) duplicate_instruction`, p0AssetE2ETenantID, contractID).Scan(&duplicateInstructionNos); err != nil {
		t.Fatal(err)
	}
	if orders != 20 || canceledOrders != 20 || clientKeys != 20 || instructions != success+canceled ||
		reconciled != success || flows != success || canceled+success/2 != 20 ||
		duplicateInstructionNos != 0 {
		t.Fatalf("race orders/canceled_orders/keys/instructions/success/canceled/reconciled/flows/duplicates=%d/%d/%d/%d/%d/%d/%d/%d/%d",
			orders, canceledOrders, clientKeys, instructions, success, canceled, reconciled, flows, duplicateInstructionNos)
	}
}

func p0AdminContext(ctx context.Context, operatorID, tenantID int64) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUid, strconv.FormatInt(operatorID, 10),
		utils.CtxKeyTenantId, strconv.FormatInt(tenantID, 10),
		utils.CtxKeyUserType, strconv.FormatInt(utils.SysUserTypeTenantAdmin, 10),
	))
}

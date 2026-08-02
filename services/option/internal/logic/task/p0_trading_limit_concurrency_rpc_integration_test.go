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
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"

	"github.com/shopspring/decimal"
)

type p0LimitAdmissionRequest struct {
	userID    int64
	accountID int64
	request   *option.PlaceOrderReq
}

type p0LimitAdmissionResult struct {
	request  p0LimitAdmissionRequest
	response *option.PlaceOrderResp
	err      error
}

type p0AcceptedLimitOrder struct {
	userID, accountID, orderID int64
}

// testP0ConcurrentTradingLimits proves that projected exposure is serialized
// at the public order boundary. The first scenario races one user's twenty
// requests against a ten-contract long limit. The second races twenty users
// against a ten-contract OI limit. Both must admit exactly ten economic
// identities, reject ten with the intended control reason, and fully release
// every successful Asset freeze after cancellation.
func testP0ConcurrentTradingLimits(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P0_CONCURRENT_LIMITS_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)

	userContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-CONCURRENT-USER-LONG-LIMIT", calendarCode, 95100, now,
	)
	userContract.MaxUserLongQty = decimal.NewFromInt(10)
	userContract.MaxOpenInterest = decimal.NewFromInt(100)
	if err := serviceCtx.OptionContractModel.Update(ctx, userContract); err != nil {
		t.Fatalf("configure concurrent user limit: %v", err)
	}
	insertP0ExerciseMarket(t, ctx, serviceCtx, userContract.Id, "100", "10", now)
	creditAsset(t, ctx, assetClient, 95101, "200", "P0-CONCURRENT-USER-LIMIT-SEED")
	userRequests := make([]p0LimitAdmissionRequest, 0, 20)
	for index := 0; index < 20; index++ {
		userRequests = append(userRequests, p0LimitAdmissionRequest{
			userID: 95101, accountID: 95101,
			request: &option.PlaceOrderReq{
				AccountId: 95101, ContractId: userContract.Id,
				Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
				OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
				ClientOrderId: fmt.Sprintf("P0-CONCURRENT-USER-LIMIT-%02d", index+1),
			},
		})
	}
	userAccepted := runP0ConcurrentLimitAdmission(
		t, ctx, serviceCtx, userRequests, 10, 10,
	)
	finishP0LimitOrders(t, ctx, serviceCtx, userAccepted)
	assertP0ConcurrentLimitEvidence(
		t, ctx, db, userContract.Id, "USER_LONG_LIMIT", 20, 10, 10,
	)
	assertWalletAmounts(t, ctx, db, 95101,
		"200.000000000000000000", "200.000000000000000000", "0.000000000000000000")

	oiContract := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P0-CONCURRENT-OI-LIMIT", calendarCode, 95200, now,
	)
	oiContract.MaxUserLongQty = decimal.NewFromInt(100)
	oiContract.MaxOpenInterest = decimal.NewFromInt(10)
	if err := serviceCtx.OptionContractModel.Update(ctx, oiContract); err != nil {
		t.Fatalf("configure concurrent OI limit: %v", err)
	}
	insertP0ExerciseMarket(t, ctx, serviceCtx, oiContract.Id, "100", "10", now)
	oiRequests := make([]p0LimitAdmissionRequest, 0, 20)
	for index := 0; index < 20; index++ {
		userID := int64(95201 + index)
		creditAsset(t, ctx, assetClient, userID, "100",
			fmt.Sprintf("P0-CONCURRENT-OI-LIMIT-SEED-%02d", index+1))
		oiRequests = append(oiRequests, p0LimitAdmissionRequest{
			userID: userID, accountID: userID,
			request: &option.PlaceOrderReq{
				AccountId: userID, ContractId: oiContract.Id,
				Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
				OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
				ClientOrderId: fmt.Sprintf("P0-CONCURRENT-OI-LIMIT-%02d", index+1),
			},
		})
	}
	oiAccepted := runP0ConcurrentLimitAdmission(
		t, ctx, serviceCtx, oiRequests, 10, 10,
	)
	finishP0LimitOrders(t, ctx, serviceCtx, oiAccepted)
	assertP0ConcurrentLimitEvidence(
		t, ctx, db, oiContract.Id, "OPEN_INTEREST_LIMIT", 20, 10, 10,
	)
	assertP0LimitWalletRange(t, ctx, db, 95201, 95220,
		"2000.000000000000000000", "2000.000000000000000000", "0.000000000000000000")
}

func runP0ConcurrentLimitAdmission(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	requests []p0LimitAdmissionRequest,
	wantAccepted, wantRejected int,
) []p0AcceptedLimitOrder {
	t.Helper()
	start := make(chan struct{})
	results := make(chan p0LimitAdmissionResult, len(requests))
	var workers sync.WaitGroup
	for _, request := range requests {
		request := request
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			response, err := applogic.NewPlaceOrderLogic(
				p0OrderUserContext(ctx, request.userID), serviceCtx,
			).PlaceOrder(request.request)
			results <- p0LimitAdmissionResult{request: request, response: response, err: err}
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	accepted := make([]p0AcceptedLimitOrder, 0, wantAccepted)
	rejected := 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent limit admission user=%d: %v", result.request.userID, result.err)
		}
		if result.response == nil || result.response.Base == nil {
			t.Fatalf("concurrent limit admission user=%d response=%+v",
				result.request.userID, result.response)
		}
		if result.response.Base.Code != 200 {
			rejected++
			continue
		}
		if result.response.Data == nil || result.response.Data.OrderId <= 0 {
			t.Fatalf("accepted concurrent limit order missing identity user=%d response=%+v",
				result.request.userID, result.response)
		}
		accepted = append(accepted, p0AcceptedLimitOrder{
			userID: result.request.userID, accountID: result.request.accountID,
			orderID: result.response.Data.OrderId,
		})
	}
	if len(accepted) != wantAccepted || rejected != wantRejected {
		t.Fatalf("concurrent limit accepted/rejected=%d/%d want=%d/%d",
			len(accepted), rejected, wantAccepted, wantRejected)
	}
	return accepted
}

func finishP0LimitOrders(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	accepted []p0AcceptedLimitOrder,
) {
	t.Helper()
	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	for _, acceptedOrder := range accepted {
		order, err := serviceCtx.OptionOrderModel.FindOne(ctx, acceptedOrder.orderID)
		if err != nil {
			t.Fatal(err)
		}
		if order.Status != int64(option.OrderStatus_ORDER_STATUS_PENDING) {
			t.Fatalf("funded concurrent limit order %d status=%d", order.Id, order.Status)
		}
		assertP0UserCancelOK(
			t, ctx, serviceCtx, acceptedOrder.userID, acceptedOrder.accountID, acceptedOrder.orderID,
		)
	}
	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
}

func assertP0ConcurrentLimitEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	rejectionReason string,
	evaluations, accepted, rejected int64,
) {
	t.Helper()
	var orders, canceled, clientKeys, instructions, success, reconciled, flows, duplicateFlows int64
	var evaluatedEvents, rejectedEvents, wrongRejectEvents, trades, positions, activeOrders int64
	if err := db.QueryRowContext(ctx, `SELECT
		COUNT(DISTINCT orders.id),
		COUNT(DISTINCT IF(orders.status=4,orders.id,NULL)),
		COUNT(DISTINCT client_key.id),
		COUNT(DISTINCT instruction.id),
		COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
		COUNT(DISTINCT flow.id),
		COUNT(flow.id)-COUNT(DISTINCT flow.flow_no)
		FROM t_option_order orders
		LEFT JOIN t_option_client_order_key client_key
		  ON client_key.tenant_id=orders.tenant_id AND client_key.order_id=orders.id
		LEFT JOIN t_option_asset_instruction instruction
		  ON instruction.tenant_id=orders.tenant_id AND instruction.order_id=orders.id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id
		 AND flow.biz_no=CASE WHEN instruction.action=1
		   THEN instruction.target_biz_no ELSE instruction.instruction_no END
		WHERE orders.tenant_id=? AND orders.contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&orders, &canceled, &clientKeys, &instructions, &success, &reconciled, &flows, &duplicateFlows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		SUM(event_type='ORDER_CONTROL_EVALUATED'),
		SUM(event_type='ORDER_REJECTED' AND reason=?),
		SUM(event_type='ORDER_REJECTED' AND reason<>?)
		FROM t_option_trading_control_event
		WHERE tenant_id=? AND contract_id=?`, rejectionReason, rejectionReason,
		p0AssetE2ETenantID, contractID,
	).Scan(&evaluatedEvents, &rejectedEvents, &wrongRejectEvents); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_trade WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND contract_id=?),
		(SELECT COUNT(*) FROM t_option_order WHERE tenant_id=? AND contract_id=? AND status IN (1,2,7))`,
		p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID,
		p0AssetE2ETenantID, contractID,
	).Scan(&trades, &positions, &activeOrders); err != nil {
		t.Fatal(err)
	}
	if orders != accepted || canceled != accepted || clientKeys != accepted ||
		instructions != accepted*2 || success != instructions || reconciled != instructions ||
		flows != instructions || duplicateFlows != 0 || evaluatedEvents != evaluations ||
		rejectedEvents != rejected || wrongRejectEvents != 0 || trades != 0 || positions != 0 || activeOrders != 0 {
		t.Fatalf("concurrent limit contract=%d orders/canceled/keys=%d/%d/%d instructions=%d/%d/%d flows=%d duplicate=%d events=%d/%d/%d trades/positions/active=%d/%d/%d",
			contractID, orders, canceled, clientKeys, instructions, success, reconciled,
			flows, duplicateFlows, evaluatedEvents, rejectedEvents, wrongRejectEvents,
			trades, positions, activeOrders)
	}
}

func assertP0LimitWalletRange(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	firstUserID, lastUserID int64,
	total, available, frozen string,
) {
	t.Helper()
	var wallets int64
	var gotTotal, gotAvailable, gotFrozen string
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),CAST(SUM(total_amount) AS CHAR),
		CAST(SUM(available_amount) AS CHAR),CAST(SUM(frozen_amount) AS CHAR)
		FROM t_user_asset WHERE tenant_id=? AND wallet_type=? AND coin='USDT'
		  AND user_id BETWEEN ? AND ?`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION), firstUserID, lastUserID,
	).Scan(&wallets, &gotTotal, &gotAvailable, &gotFrozen); err != nil {
		t.Fatal(err)
	}
	if wallets != lastUserID-firstUserID+1 || gotTotal != total || gotAvailable != available || gotFrozen != frozen {
		t.Fatalf("concurrent limit wallets=%d amounts=%s/%s/%s want=%d %s/%s/%s",
			wallets, gotTotal, gotAvailable, gotFrozen, lastUserID-firstUserID+1,
			total, available, frozen)
	}
}

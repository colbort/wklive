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

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type failOnceComboFreezeAssetClient struct {
	asset.AssetClient

	mu          sync.Mutex
	targetBizNo string
	failures    int
}

func (c *failOnceComboFreezeAssetClient) FreezeAsset(
	ctx context.Context,
	in *asset.FreezeAssetReq,
	opts ...grpc.CallOption,
) (*asset.FreezeAssetResp, error) {
	c.mu.Lock()
	shouldFail := in.BizNo == c.targetBizNo && c.failures == 0
	if shouldFail {
		c.failures++
	}
	c.mu.Unlock()
	if !shouldFail {
		return c.AssetClient.FreezeAsset(ctx, in, opts...)
	}
	if _, err := c.AssetClient.FreezeAsset(ctx, in, opts...); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unavailable, "COMBO-003 injected freeze response loss after commit")
}

func (c *failOnceComboFreezeAssetClient) failureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

func testP2ComboOrderAcceptance(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const feeUserID int64 = 5969
	now := time.Now().Unix()
	calendarCode := "P2_COMBO_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	contractOne := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-COMBO-100-C", calendarCode, feeUserID, now,
	)
	contractTwo := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-COMBO-110-C", calendarCode, feeUserID, now,
	)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractOne.Id, "100", "10", now)
	insertP0ExerciseMarket(t, ctx, serviceCtx, contractTwo.Id, "100", "8", now)

	t.Run("COMBO-002 50-way canonical idempotency", func(t *testing.T) {
		testP2ComboConcurrentIdempotency(t, ctx, db, serviceCtx, contractOne.Id, contractTwo.Id)
	})
	t.Run("COMBO-003 committed freeze response loss", func(t *testing.T) {
		testP2ComboFreezeResponseLoss(
			t, ctx, db, assetClient, serviceCtx, contractOne.Id, contractTwo.Id,
		)
	})
	t.Run("COMBO-007 post-funding kill and margin gates", func(t *testing.T) {
		testP2ComboPostFundingControls(
			t, ctx, db, assetClient, serviceCtx, contractOne.Id, contractTwo.Id,
		)
	})
	t.Run("COMBO-004-008 atomic matching rollback STP FOK and debit barrier", func(t *testing.T) {
		testP2ComboAtomicMatching(
			t, ctx, db, assetClient, serviceCtx, contractOne.Id, contractTwo.Id,
		)
	})
}

func testP2ComboConcurrentIdempotency(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	contractOneID, contractTwoID int64,
) {
	t.Helper()
	const (
		userID    int64 = 5961
		accountID int64 = 6961
		parallel        = 50
	)
	req := p2ComboMakerRequest(accountID, "P2-COMBO-IDEMPOTENT-50", contractOneID, contractTwoID, "1")
	start := make(chan struct{})
	type result struct {
		id   int64
		code int32
		err  error
	}
	results := make(chan result, parallel)
	var wg sync.WaitGroup
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			resp, err := applogic.NewPlaceComboOrderLogic(
				p0OrderUserContext(ctx, userID), serviceCtx,
			).PlaceComboOrder(req)
			item := result{err: err}
			if resp != nil && resp.Base != nil {
				item.code = resp.Base.Code
			}
			if resp != nil && resp.Data != nil && resp.Data.ComboOrder != nil {
				item.id = resp.Data.ComboOrder.Id
			}
			results <- item
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	var parentID int64
	for item := range results {
		if item.err != nil || item.code != 200 || item.id <= 0 {
			t.Fatalf("50-way combo replay result=%+v", item)
		}
		if parentID == 0 {
			parentID = item.id
		}
		if item.id != parentID {
			t.Fatalf("50-way combo returned parent=%d want=%d", item.id, parentID)
		}
	}

	var parents, legs, children, freezes int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_combo_order WHERE tenant_id=? AND user_id=? AND client_combo_id=?),
		(SELECT COUNT(*) FROM t_option_combo_order_leg WHERE tenant_id=? AND combo_order_id=?),
		(SELECT COUNT(*) FROM t_option_order WHERE tenant_id=? AND combo_order_id=?),
		(SELECT COUNT(*) FROM t_option_asset_instruction instruction
		 JOIN t_option_order child ON child.tenant_id=instruction.tenant_id AND child.id=instruction.order_id
		 WHERE instruction.tenant_id=? AND child.combo_order_id=? AND instruction.action=?)`,
		p0AssetE2ETenantID, userID, req.ClientComboId,
		p0AssetE2ETenantID, parentID,
		p0AssetE2ETenantID, parentID,
		p0AssetE2ETenantID, parentID,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
	).Scan(&parents, &legs, &children, &freezes); err != nil {
		t.Fatal(err)
	}
	if parents != 1 || legs != 2 || children != 2 || freezes != 2 {
		t.Fatalf("50-way combo parent/legs/children/freezes=%d/%d/%d/%d", parents, legs, children, freezes)
	}

	conflict := p2ComboMakerRequest(accountID, req.ClientComboId, contractOneID, contractTwoID, "2")
	resp, err := applogic.NewPlaceComboOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).PlaceComboOrder(conflict)
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code == 200 {
		t.Fatalf("different-payload combo replay resp=%+v err=%v", resp, err)
	}
	if err := applogic.CancelComboOrderByControl(
		ctx, serviceCtx, parentID, "COMBO_002_ACCEPTANCE_CLEANUP",
	); err != nil {
		t.Fatal(err)
	}
	p2AssertComboParentState(
		t, ctx, serviceCtx, parentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED, "COMBO_002_ACCEPTANCE_CLEANUP",
	)
}

func testP2ComboFreezeResponseLoss(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contractOneID, contractTwoID int64,
) {
	t.Helper()
	const (
		userID    int64 = 5962
		accountID int64 = 6962
	)
	creditAsset(t, ctx, assetClient, userID, "1000", "P2-COMBO-FREEZE-LOSS-SEED")
	parentID := p2PlaceCombo(
		t, ctx, serviceCtx, userID,
		p2ComboMakerRequest(accountID, "P2-COMBO-FREEZE-LOSS", contractOneID, contractTwoID, "1"),
	)
	var targetBizNo string
	if err := db.QueryRowContext(ctx, `SELECT order_no FROM t_option_order
		WHERE tenant_id=? AND combo_order_id=? ORDER BY combo_leg_no LIMIT 1`,
		p0AssetE2ETenantID, parentID,
	).Scan(&targetBizNo); err != nil {
		t.Fatal(err)
	}
	fault := &failOnceComboFreezeAssetClient{AssetClient: assetClient, targetBizNo: targetBizNo}
	original := serviceCtx.AssetClient
	serviceCtx.AssetClient = fault
	defer func() { serviceCtx.AssetClient = original }()
	processAssetInstructions(t, ctx, serviceCtx)
	if fault.failureCount() != 1 {
		t.Fatalf("combo committed freeze response losses=%d want=1", fault.failureCount())
	}
	var failed, success int64
	if err := db.QueryRowContext(ctx, `SELECT
		SUM(instruction.status=?),SUM(instruction.status=?)
		FROM t_option_asset_instruction instruction
		JOIN t_option_order child ON child.tenant_id=instruction.tenant_id AND child.id=instruction.order_id
		WHERE instruction.tenant_id=? AND child.combo_order_id=? AND instruction.action=?`,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID, parentID,
		int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
	).Scan(&failed, &success); err != nil {
		t.Fatal(err)
	}
	if failed != 1 || success != 1 {
		t.Fatalf("combo freeze loss failed/success=%d/%d", failed, success)
	}
	p2AssertComboParentState(
		t, ctx, serviceCtx, parentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING, "",
	)
	p2MakeComboInstructionsRetryable(t, ctx, db, parentID)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, parentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE, "",
	)
	if err := applogic.CancelComboOrderByControl(
		ctx, serviceCtx, parentID, "COMBO_003_ACCEPTANCE_CLEANUP",
	); err != nil {
		t.Fatal(err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, parentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED, "COMBO_003_ACCEPTANCE_CLEANUP",
	)
	assertWalletAmounts(t, ctx, db, userID,
		"1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
}

func testP2ComboPostFundingControls(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contractOneID, contractTwoID int64,
) {
	t.Helper()
	const (
		killUserID   int64 = 5963
		killAccount  int64 = 6963
		marginUserID int64 = 5964
		marginAcct   int64 = 6964
		adminUserID  int64 = 5968
		adminAcct    int64 = 6968
		operatorID   int64 = 9968
	)
	creditAsset(t, ctx, assetClient, killUserID, "1000", "P2-COMBO-KILL-SEED")
	killParentID := p2PlaceCombo(
		t, ctx, serviceCtx, killUserID,
		p2ComboMakerRequest(killAccount, "P2-COMBO-KILL-AFTER-PLACE", contractOneID, contractTwoID, "1"),
	)
	if _, err := db.ExecContext(ctx, `UPDATE t_option_user_trading_control
		SET kill_switch=?,activated_at=?,reason='COMBO-007 acceptance',update_times=?
		WHERE tenant_id=? AND user_id=?`,
		int64(common.YesNo_YES_NO_YES), time.Now().Unix(), time.Now().Unix(),
		p0AssetE2ETenantID, killUserID,
	); err != nil {
		t.Fatal(err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, killParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED,
		"COMBO_USER_KILL_SWITCH_AFTER_FUNDING",
	)
	assertWalletAmounts(t, ctx, db, killUserID,
		"1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
	if _, err := db.ExecContext(ctx, `UPDATE t_option_user_trading_control
		SET kill_switch=?,activated_at=0,reason='',update_times=? WHERE tenant_id=? AND user_id=?`,
		int64(common.YesNo_YES_NO_NO), time.Now().Unix(), p0AssetE2ETenantID, killUserID,
	); err != nil {
		t.Fatal(err)
	}

	creditAsset(t, ctx, assetClient, marginUserID, "1000", "P2-COMBO-MARGIN-SEED")
	marginParentID := p2PlaceCombo(
		t, ctx, serviceCtx, marginUserID,
		p2ComboMakerRequest(marginAcct, "P2-COMBO-MARGIN-DRIFT", contractOneID, contractTwoID, "1"),
	)
	now := time.Now().Unix()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_market
		SET underlying_price=1000,underlying_snapshot_time=?,snapshot_time=?,update_times=?
		WHERE tenant_id=? AND contract_id=?`,
		now, now, now, p0AssetE2ETenantID, contractTwoID,
	); err != nil {
		t.Fatal(err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, marginParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED,
		"COMBO_SELL_MARGIN_INSUFFICIENT_AFTER_FUNDING",
	)
	assertWalletAmounts(t, ctx, db, marginUserID,
		"1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
	now = time.Now().Unix()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_market
		SET underlying_price=100,underlying_snapshot_time=?,mark_snapshot_time=?,snapshot_time=?,update_times=?
		WHERE tenant_id=? AND contract_id IN (?,?)`,
		now, now, now, now, p0AssetE2ETenantID, contractOneID, contractTwoID,
	); err != nil {
		t.Fatal(err)
	}

	creditAsset(t, ctx, assetClient, adminUserID, "1000", "P2-COMBO-ADMIN-SEED")
	adminParentID := p2PlaceCombo(
		t, ctx, serviceCtx, adminUserID,
		p2ComboMakerRequest(adminAcct, "P2-COMBO-ADMIN-CANCEL", contractOneID, contractTwoID, "1"),
	)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, adminParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE, "",
	)
	unauthorized, err := adminlogic.NewGetAdminComboOrderLogic(
		p0AdminContext(ctx, operatorID, p0AssetE2ETenantID+1), serviceCtx,
	).GetAdminComboOrder(&option.GetAdminComboOrderReq{
		TenantId: p0AssetE2ETenantID, Id: adminParentID,
	})
	if err != nil || unauthorized == nil || unauthorized.Base == nil || unauthorized.Base.Code == 200 {
		t.Fatalf("cross-tenant combo detail was not rejected resp=%+v err=%v", unauthorized, err)
	}
	adminCtx := p0AdminContext(ctx, operatorID, p0AssetE2ETenantID)
	detail, err := adminlogic.NewGetAdminComboOrderLogic(adminCtx, serviceCtx).
		GetAdminComboOrder(&option.GetAdminComboOrderReq{
			TenantId: p0AssetE2ETenantID, Id: adminParentID,
		})
	if err != nil || detail == nil || detail.Base == nil || detail.Base.Code != 200 ||
		detail.Data == nil || len(detail.Data.Legs) != 2 || len(detail.Data.ChildOrders) != 2 ||
		detail.Data.AssetInstructionTotal != 2 {
		t.Fatalf("admin combo detail resp=%+v err=%v", detail, err)
	}
	missingReason, err := adminlogic.NewForceCancelComboOrderLogic(adminCtx, serviceCtx).
		ForceCancelComboOrder(&option.ForceCancelComboOrderReq{
			TenantId: p0AssetE2ETenantID, Id: adminParentID,
		})
	if err != nil || missingReason == nil || missingReason.Base == nil || missingReason.Base.Code == 200 {
		t.Fatalf("reasonless combo force cancel was not rejected resp=%+v err=%v", missingReason, err)
	}
	forbiddenCancel, err := adminlogic.NewForceCancelComboOrderLogic(
		p0AdminContext(ctx, operatorID, p0AssetE2ETenantID+1), serviceCtx,
	).ForceCancelComboOrder(&option.ForceCancelComboOrderReq{
		TenantId: p0AssetE2ETenantID, Id: adminParentID, Reason: "CROSS_TENANT",
	})
	if err != nil || forbiddenCancel == nil || forbiddenCancel.Base == nil || forbiddenCancel.Base.Code == 200 {
		t.Fatalf("cross-tenant combo force cancel was not rejected resp=%+v err=%v", forbiddenCancel, err)
	}
	canceled, err := adminlogic.NewForceCancelComboOrderLogic(adminCtx, serviceCtx).
		ForceCancelComboOrder(&option.ForceCancelComboOrderReq{
			TenantId: p0AssetE2ETenantID, Id: adminParentID, Reason: "COMBO_009_ACCEPTANCE",
		})
	if err != nil || canceled == nil || canceled.Base == nil || canceled.Base.Code != 200 {
		t.Fatalf("admin combo force cancel resp=%+v err=%v", canceled, err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, adminParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED,
		"ADMIN_FORCE_CANCEL:COMBO_009_ACCEPTANCE",
	)
	assertWalletAmounts(t, ctx, db, adminUserID,
		"1000.000000000000000000", "1000.000000000000000000", "0.000000000000000000")
}

func testP2ComboAtomicMatching(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contractOneID, contractTwoID int64,
) {
	t.Helper()
	const (
		makerUserID    int64 = 5965
		makerAccountID int64 = 6965
		fokUserID      int64 = 5966
		stpAccountID   int64 = 6966
		takerUserID    int64 = 5967
		takerAccountID int64 = 6967
	)
	for _, seed := range []struct {
		userID int64
		bizNo  string
	}{
		{makerUserID, "P2-COMBO-MAKER-SEED"},
		{fokUserID, "P2-COMBO-FOK-SEED"},
		{takerUserID, "P2-COMBO-TAKER-SEED"},
	} {
		creditAsset(t, ctx, assetClient, seed.userID, "2000", seed.bizNo)
	}
	makerParentID := p2PlaceCombo(
		t, ctx, serviceCtx, makerUserID,
		p2ComboMakerRequest(makerAccountID, "P2-COMBO-MAKER", contractOneID, contractTwoID, "1"),
	)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, makerParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_ACTIVE, "",
	)
	levels, err := serviceCtx.OptionOrderModel.FindOrderBookLevels(
		ctx, p0AssetE2ETenantID, contractOneID, int64(common.Side_SIDE_BUY), 100,
	)
	if err != nil || len(levels) != 0 {
		t.Fatalf("combo shadow leaked into ordinary book levels=%+v err=%v", levels, err)
	}

	fokParentID := p2PlaceCombo(
		t, ctx, serviceCtx, fokUserID,
		p2ComboTakerRequest(fokUserID+1000, "P2-COMBO-FOK-INSUFFICIENT", contractOneID, contractTwoID, "2", option.ComboOrderType_COMBO_ORDER_TYPE_FOK),
	)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, fokParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED, "FOK_NOT_FILLED",
	)
	p2AssertNoComboTrades(t, ctx, db, makerParentID, fokParentID)

	stpParentID := p2PlaceCombo(
		t, ctx, serviceCtx, makerUserID,
		p2ComboTakerRequest(stpAccountID, "P2-COMBO-STP", contractOneID, contractTwoID, "1", option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT),
	)
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertComboParentState(
		t, ctx, serviceCtx, stpParentID,
		option.ComboOrderStatus_COMBO_ORDER_STATUS_CANCELED, "SELF_TRADE_PREVENTED",
	)
	p2AssertNoComboTrades(t, ctx, db, makerParentID, stpParentID)

	takerParentID := p2PlaceCombo(
		t, ctx, serviceCtx, takerUserID,
		p2ComboTakerRequest(takerAccountID, "P2-COMBO-TAKER", contractOneID, contractTwoID, "1", option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT),
	)
	const triggerName = "trg_accept_combo_leg2_failure"
	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS "+triggerName); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), "DROP TRIGGER IF EXISTS "+triggerName)
	}()
	if _, err := db.ExecContext(ctx, `CREATE TRIGGER `+triggerName+`
		BEFORE INSERT ON t_option_trade FOR EACH ROW
		SET NEW.trade_no=IF(NEW.combo_match_no<>'' AND NEW.combo_leg_no=2,NULL,NEW.trade_no)`); err != nil {
		t.Fatalf("install combo second-leg failure trigger: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	p2AssertNoComboTrades(t, ctx, db, makerParentID, takerParentID)
	var sequences int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_match_sequence
		WHERE tenant_id=? AND contract_id IN (?,?)`,
		p0AssetE2ETenantID, contractOneID, contractTwoID,
	).Scan(&sequences); err != nil {
		t.Fatal(err)
	}
	if sequences != 0 {
		t.Fatalf("second-leg rollback left match sequences=%d", sequences)
	}
	if _, err := db.ExecContext(ctx, "DROP TRIGGER IF EXISTS "+triggerName); err != nil {
		t.Fatal(err)
	}
	p2MakeComboInstructionsRetryable(t, ctx, db, takerParentID)
	processAssetInstructions(t, ctx, serviceCtx)

	var trades, distinctMatches, distinctLegs int64
	var totalQty decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(DISTINCT combo_match_no),
		COUNT(DISTINCT combo_leg_no),COALESCE(SUM(qty),0)
		FROM t_option_trade WHERE tenant_id=? AND
		(buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)) OR
		 sell_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))`,
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
	).Scan(&trades, &distinctMatches, &distinctLegs, &totalQty); err != nil {
		t.Fatal(err)
	}
	if trades != 2 || distinctMatches != 1 || distinctLegs != 2 || !totalQty.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("combo trades/matches/legs/qty=%d/%d/%d/%s", trades, distinctMatches, distinctLegs, totalQty)
	}
	p2AssertComboFilledShape(t, ctx, db, makerParentID)
	p2AssertComboFilledShape(t, ctx, db, takerParentID)

	var targetBuyBizNo string
	if err := db.QueryRowContext(ctx, `SELECT buy_order_no FROM t_option_trade
		WHERE tenant_id=? AND combo_match_no<>'' AND
		buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?))
		ORDER BY combo_leg_no LIMIT 1`,
		p0AssetE2ETenantID, p0AssetE2ETenantID, makerParentID, takerParentID,
	).Scan(&targetBuyBizNo); err != nil {
		t.Fatal(err)
	}
	fault := &failOnceDeductAssetClient{
		AssetClient: assetClient, targetBizNo: targetBuyBizNo, failAfterCommit: true,
	}
	original := serviceCtx.AssetClient
	serviceCtx.AssetClient = fault
	defer func() { serviceCtx.AssetClient = original }()
	processAssetInstructions(t, ctx, serviceCtx)
	if fault.failureCount() != 1 {
		t.Fatalf("combo committed debit response losses=%d want=1", fault.failureCount())
	}
	processP0TradeEvents(t, ctx, serviceCtx)
	var pendingOutbox, positions int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_outbox outbox JOIN t_option_trade trade
		 ON trade.tenant_id=outbox.tenant_id AND trade.id=outbox.trade_id
		 WHERE outbox.tenant_id=? AND trade.combo_match_no<>'' AND outbox.status<>?),
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND user_id IN (?,?)
		 AND contract_id IN (?,?))`,
		p0AssetE2ETenantID, int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
		p0AssetE2ETenantID, makerUserID, takerUserID, contractOneID, contractTwoID,
	).Scan(&pendingOutbox, &positions); err != nil {
		t.Fatal(err)
	}
	if pendingOutbox != 2 || positions != 0 {
		t.Fatalf("combo debit barrier pending-outbox/positions=%d/%d", pendingOutbox, positions)
	}
	p2MakeComboInstructionsRetryable(t, ctx, db, takerParentID)
	p2MakeComboInstructionsRetryable(t, ctx, db, makerParentID)
	for i := 0; i < 4; i++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	processP0TradeEvents(t, ctx, serviceCtx)
	p2AssertComboSettlementEvidence(
		t, ctx, db, makerParentID, takerParentID,
		makerUserID, takerUserID, contractOneID, contractTwoID,
	)
}

func p2ComboMakerRequest(
	accountID int64,
	clientID string,
	contractOneID, contractTwoID int64,
	qty string,
) *option.PlaceComboOrderReq {
	return &option.PlaceComboOrderReq{
		AccountId: accountID, ClientComboId: clientID,
		OrderType: option.ComboOrderType_COMBO_ORDER_TYPE_LIMIT,
		Qty:       qty, NetPrice: "-6",
		Legs: []*option.ComboOrderLegInput{
			{ContractId: contractOneID, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "10"},
			{ContractId: contractTwoID, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 2, Price: "8"},
		},
	}
}

func p2ComboTakerRequest(
	accountID int64,
	clientID string,
	contractOneID, contractTwoID int64,
	qty string,
	orderType option.ComboOrderType,
) *option.PlaceComboOrderReq {
	return &option.PlaceComboOrderReq{
		AccountId: accountID, ClientComboId: clientID, OrderType: orderType,
		Qty: qty, NetPrice: "6",
		Legs: []*option.ComboOrderLegInput{
			{ContractId: contractTwoID, Side: common.Side_SIDE_BUY, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 2, Price: "8"},
			{ContractId: contractOneID, Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN, Ratio: 1, Price: "10"},
		},
	}
}

func p2PlaceCombo(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	userID int64,
	req *option.PlaceComboOrderReq,
) int64 {
	t.Helper()
	resp, err := applogic.NewPlaceComboOrderLogic(
		p0OrderUserContext(ctx, userID), serviceCtx,
	).PlaceComboOrder(req)
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 ||
		resp.Data == nil || resp.Data.ComboOrder == nil || resp.Data.ComboOrder.Id <= 0 {
		t.Fatalf("place combo user=%d resp=%+v err=%v", userID, resp, err)
	}
	return resp.Data.ComboOrder.Id
}

func p2AssertComboParentState(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	parentID int64,
	want option.ComboOrderStatus,
	wantReason string,
) {
	t.Helper()
	parent, err := serviceCtx.OptionComboOrderModel.FindOne(ctx, parentID)
	if err != nil {
		t.Fatal(err)
	}
	if parent.Status != int64(want) || (wantReason != "" && parent.CancelReason != wantReason) {
		t.Fatalf("combo parent=%d status/reason=%d/%q want=%d/%q", parentID, parent.Status, parent.CancelReason, want, wantReason)
	}
}

func p2MakeComboInstructionsRetryable(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	parentID int64,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction instruction
		JOIN t_option_order child ON child.tenant_id=instruction.tenant_id AND child.id=instruction.order_id
		SET instruction.next_retry_at=0,instruction.update_times=?
		WHERE instruction.tenant_id=? AND child.combo_order_id=? AND instruction.status=?`,
		time.Now().Unix(), p0AssetE2ETenantID, parentID,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED),
	); err != nil {
		t.Fatal(err)
	}
}

func p2AssertNoComboTrades(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	firstParentID, secondParentID int64,
) {
	t.Helper()
	var trades, outbox, tradeInstructions int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_trade trade WHERE trade.tenant_id=? AND
		 (trade.buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)) OR
		  trade.sell_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))),
		(SELECT COUNT(*) FROM t_option_outbox outbox JOIN t_option_trade trade
		 ON trade.tenant_id=outbox.tenant_id AND trade.id=outbox.trade_id
		 WHERE outbox.tenant_id=? AND
		 (trade.buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)) OR
		  trade.sell_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))),
		(SELECT COUNT(*) FROM t_option_asset_instruction instruction
		 WHERE instruction.tenant_id=? AND instruction.trade_id>0 AND
		 instruction.order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))`,
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, firstParentID, secondParentID,
		p0AssetE2ETenantID, firstParentID, secondParentID,
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, firstParentID, secondParentID,
		p0AssetE2ETenantID, firstParentID, secondParentID,
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, firstParentID, secondParentID,
	).Scan(&trades, &outbox, &tradeInstructions); err != nil {
		t.Fatal(err)
	}
	if trades != 0 || outbox != 0 || tradeInstructions != 0 {
		t.Fatalf("combo atomic rollback trades/outbox/instructions=%d/%d/%d", trades, outbox, tradeInstructions)
	}
}

func p2AssertComboFilledShape(t *testing.T, ctx context.Context, db *sql.DB, parentID int64) {
	t.Helper()
	var parentFilled, legs, filledLegs int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_combo_order WHERE tenant_id=? AND id=? AND status=? AND filled_qty=qty AND unfilled_qty=0),
		COUNT(*),SUM(filled_qty=qty AND unfilled_qty=0)
		FROM t_option_combo_order_leg WHERE tenant_id=? AND combo_order_id=?`,
		p0AssetE2ETenantID, parentID, int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_FILLED),
		p0AssetE2ETenantID, parentID,
	).Scan(&parentFilled, &legs, &filledLegs); err != nil {
		t.Fatal(err)
	}
	if parentFilled != 1 || legs != 2 || filledLegs != 2 {
		t.Fatalf("combo filled shape parent/legs/filled=%d/%d/%d", parentFilled, legs, filledLegs)
	}
}

func p2AssertComboSettlementEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	makerParentID, takerParentID int64,
	makerUserID, takerUserID int64,
	contractOneID, contractTwoID int64,
) {
	t.Helper()
	var outbox, positions, instructions, success, distinctInstructions int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_outbox outbox JOIN t_option_trade trade
		 ON trade.tenant_id=outbox.tenant_id AND trade.id=outbox.trade_id
		 WHERE outbox.tenant_id=? AND trade.combo_match_no<>'' AND outbox.status=?),
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND user_id IN (?,?) AND contract_id IN (?,?)),
		COUNT(*),SUM(instruction.status=?),COUNT(DISTINCT instruction.instruction_no)
		FROM t_option_asset_instruction instruction JOIN t_option_trade trade
		 ON trade.tenant_id=instruction.tenant_id AND trade.id=instruction.trade_id
		WHERE instruction.tenant_id=? AND
		(trade.buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)) OR
		 trade.sell_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))`,
		p0AssetE2ETenantID, int64(option.OptionEventStatus_OPTION_EVENT_STATUS_SUCCESS),
		p0AssetE2ETenantID, makerUserID, takerUserID, contractOneID, contractTwoID,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS),
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
	).Scan(&outbox, &positions, &instructions, &success, &distinctInstructions); err != nil {
		t.Fatal(err)
	}
	if outbox != 2 || positions != 4 || instructions == 0 || success != instructions ||
		distinctInstructions != instructions {
		t.Fatalf("combo settlement outbox/positions/instructions/success/distinct=%d/%d/%d/%d/%d",
			outbox, positions, instructions, success, distinctInstructions)
	}
	var ratioOne, ratioTwo decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT
		MAX(CASE WHEN combo_leg_no=1 THEN qty ELSE 0 END),
		MAX(CASE WHEN combo_leg_no=2 THEN qty ELSE 0 END)
		FROM t_option_trade WHERE tenant_id=? AND combo_match_no<>'' AND
		(buy_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)) OR
		 sell_order_id IN (SELECT id FROM t_option_order WHERE tenant_id=? AND combo_order_id IN (?,?)))`,
		p0AssetE2ETenantID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
		p0AssetE2ETenantID, makerParentID, takerParentID,
	).Scan(&ratioOne, &ratioTwo); err != nil {
		t.Fatal(err)
	}
	if !ratioOne.Equal(decimal.NewFromInt(1)) || !ratioTwo.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("combo execution ratio leg1/leg2=%s/%s", ratioOne, ratioTwo)
	}
	fmt.Printf("combo_acceptance= parents=2 legs=4 trades=2 ratio=1:2 instructions=%d positions=%d\n",
		instructions, positions)
}

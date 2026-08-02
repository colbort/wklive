package tasklogic

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
)

const (
	p2CorporateActionPositionCount int64 = 5001
	p2CorporateActionUserID        int64 = 96701
)

func testP2CorporateActionCapacityRestart(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P2_CORPORATE_ACTION_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	source := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-CORPORATE-ACTION-SOURCE", calendarCode, 96799, now,
	)
	successor := insertP2CorporateActionSuccessor(t, ctx, serviceCtx, source, now)
	firstShortID, blockedShortID := seedP2CorporateActionPositions(
		t, ctx, db, source, now,
	)

	creditAsset(t, ctx, assetClient, p2CorporateActionUserID, "200", "P2-CORPORATE-ACTION-ASSET-SEED")
	firstFreezeBiz := "P2-CORPORATE-ACTION-FREEZE-0001"
	blockedFreezeBiz := "P2-CORPORATE-ACTION-FREEZE-0101"
	freezeP2CorporateActionMargin(t, ctx, assetClient, firstShortID, firstFreezeBiz)
	freezeP2CorporateActionMargin(t, ctx, assetClient, blockedShortID, blockedFreezeBiz)
	firstLotID := insertP2CorporateActionMarginLot(
		t, ctx, serviceCtx, source.Id, firstShortID, 100001, firstFreezeBiz, false, now,
	)
	blockedLotID := insertP2CorporateActionMarginLot(
		t, ctx, serviceCtx, source.Id, blockedShortID, 100101, blockedFreezeBiz, true, now,
	)
	assertP2CorporateActionAssetEvidence(
		t, ctx, db, firstFreezeBiz, blockedFreezeBiz,
	)

	created, err := adminlogic.NewCreateCorporateActionLogic(
		p0AdminContext(ctx, 96798, p0AssetE2ETenantID), serviceCtx,
	).CreateCorporateAction(&option.CreateCorporateActionReq{
		TenantId: p0AssetE2ETenantID, EventNo: "P2-CORPORATE-ACTION-5001",
		ExternalEventRef: "P2-EXTERNAL-SPLIT-5001", UnderlyingSymbol: source.UnderlyingSymbol,
		ActionType:       option.CorporateActionType_CORPORATE_ACTION_TYPE_SPLIT,
		AnnouncementTime: now - 3600, ExTime: now - 300, RecordTime: now - 200,
		EffectiveTime: now - 1, PayTime: now + 3600,
		EvidenceRef:  "P2-CORPORATE-ACTION-AUTHORITATIVE-NOTICE",
		EvidenceHash: strings.Repeat("a", 64), Description: "2-for-1 repository capacity acceptance",
		Contracts: []*option.CorporateActionContractInput{{
			SourceContractId: source.Id, SuccessorContractId: successor.Id,
			ExecutionMode:     option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_AUTO_CASH_SUCCESSOR,
			QuantityNumerator: "2", QuantityDenominator: "1",
		}},
	})
	assertP2CorporateActionResponseOK(t, created, err, "create")
	if created.Data == nil || created.Data.Id <= 0 || len(created.Data.Contracts) != 1 {
		t.Fatalf("corporate action create data=%+v", created.Data)
	}
	actionID := created.Data.Id
	mappingID := created.Data.Contracts[0].Id
	haltID := created.Data.Contracts[0].HaltId

	if sameReviewer, sameErr := adminlogic.NewReviewCorporateActionLogic(
		p0AdminContext(ctx, 96798, p0AssetE2ETenantID), serviceCtx,
	).ReviewCorporateAction(&option.ReviewCorporateActionReq{
		TenantId: p0AssetE2ETenantID, ActionId: actionID, Approve: true, Reason: "self review forbidden",
	}); sameErr == nil && sameReviewer != nil && sameReviewer.Base != nil && sameReviewer.Base.Code == 200 {
		t.Fatal("corporate action creator reviewed its own event")
	}

	expiringOrder := insertP2CorporateActionExpiringOrder(t, ctx, serviceCtx, source, now)
	if blockedReview, blockedErr := adminlogic.NewReviewCorporateActionLogic(
		p0AdminContext(ctx, 96797, p0AssetE2ETenantID), serviceCtx,
	).ReviewCorporateAction(&option.ReviewCorporateActionReq{
		TenantId: p0AssetE2ETenantID, ActionId: actionID, Approve: true,
		Reason: "must remain blocked while expiry release is in flight",
	}); blockedErr == nil && blockedReview != nil && blockedReview.Base != nil && blockedReview.Base.Code == 200 {
		t.Fatal("corporate action review crossed an EXPIRING order")
	}
	if _, err = db.ExecContext(ctx, `UPDATE t_option_order
		SET status=?,cancel_reason='P2_CORPORATE_ACTION_EXPIRY_RELEASED',cancel_time=?,update_times=?
		WHERE tenant_id=? AND id=? AND status=?`,
		int64(option.OrderStatus_ORDER_STATUS_EXPIRED), now, now,
		p0AssetE2ETenantID, expiringOrder.Id,
		int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
	); err != nil {
		t.Fatalf("finish corporate action expiry blocker: %v", err)
	}

	reviewed, err := adminlogic.NewReviewCorporateActionLogic(
		p0AdminContext(ctx, 96797, p0AssetE2ETenantID), serviceCtx,
	).ReviewCorporateAction(&option.ReviewCorporateActionReq{
		TenantId: p0AssetE2ETenantID, ActionId: actionID, Approve: true,
		Reason: "independent review confirms exact 2-for-1 conversion",
	})
	assertP2CorporateActionResponseOK(t, reviewed, err, "review")

	var taskErr, riskErr error
	start := make(chan struct{})
	var done sync.WaitGroup
	done.Add(2)
	go func() {
		defer done.Done()
		<-start
		_, taskErr = NewProcessCorporateActionsLogic(ctx, serviceCtx).
			ProcessCorporateActions(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID})
	}()
	go func() {
		defer done.Done()
		<-start
		_, riskErr = NewProcessRiskAccountsLogic(ctx, serviceCtx).
			ProcessRiskAccounts(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID})
	}()
	close(start)
	done.Wait()
	if taskErr != nil {
		t.Fatalf("first corporate action batch error=%v", taskErr)
	}
	if riskErr == nil || !strings.Contains(riskErr.Error(), "corporate action migration active") {
		t.Fatalf("risk scan did not fail closed for corporate action migration: %v", riskErr)
	}
	assertP2CorporateActionMappingState(t, ctx, db, mappingID,
		int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_EXECUTING), 100, 0)
	assertP2CorporateActionRiskRestricted(t, ctx, db)

	// Position 101 carries a still-pending margin reservation. The second task
	// invocation must commit no partial successor/audit row for that position,
	// persist the cursor at 100, and retain the original Asset freeze.
	if _, err = NewProcessCorporateActionsLogic(ctx, serviceCtx).
		ProcessCorporateActions(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID}); err != nil {
		t.Fatalf("corporate action failure isolation task: %v", err)
	}
	assertP2CorporateActionMappingState(t, ctx, db, mappingID,
		int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_FAILED), 100, 1)
	var blockedAudit, blockedSuccessor int64
	if err = db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_corporate_action_position
		 WHERE tenant_id=? AND action_contract_id=? AND source_position_id=?),
		(SELECT COUNT(*) FROM t_option_position
		 WHERE tenant_id=? AND contract_id=? AND account_id=100101 AND status=?)`,
		p0AssetE2ETenantID, mappingID, blockedShortID,
		p0AssetE2ETenantID, successor.Id,
		int64(option.PositionStatus_POSITION_STATUS_HOLDING),
	).Scan(&blockedAudit, &blockedSuccessor); err != nil {
		t.Fatal(err)
	}
	if blockedAudit != 0 || blockedSuccessor != 0 {
		t.Fatalf("failed position left partial audit/successor=%d/%d", blockedAudit, blockedSuccessor)
	}
	assertP2CorporateActionAssetEvidence(
		t, ctx, db, firstFreezeBiz, blockedFreezeBiz,
	)

	if _, err = db.ExecContext(ctx, `UPDATE t_option_margin_lot
		SET pending_margin=0,update_times=? WHERE tenant_id=? AND id=? AND pending_margin=50`,
		now, p0AssetE2ETenantID, blockedLotID,
	); err != nil {
		t.Fatalf("resolve corporate action pending margin evidence: %v", err)
	}

	completed := false
	processCalls := 0
	for processCalls < 60 {
		processCalls++
		// A new logic object on every invocation proves that the cursor and all
		// recovery state are durable rather than process-local.
		if _, err = NewProcessCorporateActionsLogic(ctx, serviceCtx).
			ProcessCorporateActions(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID}); err != nil {
			t.Fatalf("corporate action restart call %d: %v", processCalls, err)
		}
		var status int64
		if err = db.QueryRowContext(ctx, `SELECT status FROM t_option_corporate_action
			WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, actionID).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status == int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_COMPLETED) {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatalf("corporate action did not complete after %d restart calls", processCalls)
	}

	assertP2CorporateActionFinalEvidence(
		t, ctx, db, actionID, mappingID, haltID, source, successor,
		firstShortID, blockedShortID, firstLotID, blockedLotID,
		firstFreezeBiz, blockedFreezeBiz, processCalls,
	)
	testP2CorporateActionManualOnly(t, ctx, db, serviceCtx, calendarCode, now)
}

func insertP2CorporateActionSuccessor(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	source *models.TOptionContract,
	now int64,
) *models.TOptionContract {
	t.Helper()
	successor := *source
	successor.Id = 0
	successor.ContractCode = "P2-CORPORATE-ACTION-SUCCESSOR"
	successor.StrikePrice = decimal.NewFromInt(50)
	successor.MaxUserLongQty = decimal.NewFromInt(20000)
	successor.MaxUserShortQty = decimal.NewFromInt(20000)
	successor.MaxOpenInterest = decimal.NewFromInt(20000)
	successor.Status = int64(option.ContractStatus_CONTRACT_STATUS_PENDING)
	successor.CreateTimes = now
	successor.UpdateTimes = now
	result, err := serviceCtx.OptionContractModel.Insert(ctx, &successor)
	if err != nil {
		t.Fatalf("insert corporate action successor: %v", err)
	}
	successor.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return &successor
}

func seedP2CorporateActionPositions(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contract *models.TOptionContract,
	now int64,
) (int64, int64) {
	t.Helper()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO t_option_position (
		tenant_id,user_id,account_id,contract_id,underlying_symbol,side,
		position_qty,available_qty,frozen_qty,open_avg_price,mark_price,position_value,
		margin_amount,maintenance_margin,unrealized_pnl,realized_pnl,trade_realized_pnl,
		settlement_realized_pnl,fee_paid,total_return,exerciseable_qty,status,last_calc_time,
		create_times,update_times
	) VALUES (?,?,?,?,?,?,1,1,0,10,10,10,?,?,?,?,?,?,?,?,1,?,0,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	var firstShortID, blockedShortID int64
	for index := int64(1); index <= p2CorporateActionPositionCount; index++ {
		side := int64(common.PositionSide_POSITION_SIDE_LONG)
		margin := "0"
		maintenance := "0"
		if index == 1 || index == 101 {
			side = int64(common.PositionSide_POSITION_SIDE_SHORT)
			margin = "50"
			maintenance = "20"
		}
		result, execErr := stmt.ExecContext(ctx,
			p0AssetE2ETenantID, p2CorporateActionUserID, 100000+index,
			contract.Id, contract.UnderlyingSymbol, side,
			margin, maintenance, "0", "0", "0", "0", "0", "0",
			int64(option.PositionStatus_POSITION_STATUS_HOLDING), now-index, now-index,
		)
		if execErr != nil {
			t.Fatalf("insert corporate action position %d: %v", index, execErr)
		}
		positionID, idErr := result.LastInsertId()
		if idErr != nil {
			t.Fatal(idErr)
		}
		if index == 1 {
			firstShortID = positionID
		}
		if index == 101 {
			blockedShortID = positionID
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return firstShortID, blockedShortID
}

func freezeP2CorporateActionMargin(
	t *testing.T,
	ctx context.Context,
	assetClient asset.AssetClient,
	positionID int64,
	bizNo string,
) {
	t.Helper()
	response, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: p2CorporateActionUserID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "50",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: positionID, BizNo: bizNo, Remark: "corporate action preserved margin lot",
	})
	assertAssetOK(t, response, err)
}

func insertP2CorporateActionMarginLot(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractID, positionID, accountID int64,
	freezeBizNo string,
	pending bool,
	now int64,
) int64 {
	t.Helper()
	pendingMargin := decimal.Zero
	if pending {
		pendingMargin = decimal.NewFromInt(50)
	}
	lot := &models.TOptionMarginLot{
		TenantId: p0AssetE2ETenantID, UserId: p2CorporateActionUserID, AccountId: accountID,
		ContractId: contractID, PositionId: positionID,
		OriginContractId: contractID, OriginPositionId: positionID,
		TradeId: -positionID, FreezeBizNo: freezeBizNo, CollateralCoin: "USDT",
		Quantity: decimal.NewFromInt(1), RemainingQuantity: decimal.NewFromInt(1),
		InitialMargin: decimal.NewFromInt(50), RemainingMargin: decimal.NewFromInt(50),
		PendingMargin: pendingMargin, Status: int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionMarginLotModel.Insert(ctx, lot)
	if err != nil {
		t.Fatalf("insert corporate action margin lot: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func insertP2CorporateActionExpiringOrder(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	now int64,
) *models.TOptionOrder {
	t.Helper()
	return insertP0MarginOrder(t, ctx, serviceCtx, &models.TOptionOrder{
		TenantId: p0AssetE2ETenantID, OrderNo: "P2-CORPORATE-ACTION-EXPIRING",
		UserId: p2CorporateActionUserID, AccountId: 199999, ContractId: contract.Id,
		UnderlyingSymbol: contract.UnderlyingSymbol, Side: int64(common.Side_SIDE_BUY),
		PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		OrderType:      int64(option.OrderType_ORDER_TYPE_LIMIT), Price: decimal.NewFromInt(1),
		Qty: decimal.NewFromInt(1), UnfilledQty: decimal.NewFromInt(1), FeeCoin: "USDT",
		MarginCoin: "USDT", Source: int64(option.OrderSource_ORDER_SOURCE_APP),
		ClientOrderId: "P2-CORPORATE-ACTION-EXPIRING", ReduceOnly: int64(common.YesNo_YES_NO_NO),
		Mmp: int64(common.YesNo_YES_NO_NO), Status: int64(option.OrderStatus_ORDER_STATUS_EXPIRING),
		CreateTimes: now, UpdateTimes: now,
	})
}

func assertP2CorporateActionResponseOK(
	t *testing.T,
	response *option.GetCorporateActionResp,
	err error,
	operation string,
) {
	t.Helper()
	if err != nil || response == nil || response.Base == nil || response.Base.Code != 200 {
		t.Fatalf("corporate action %s response=%+v err=%v", operation, response, err)
	}
}

func assertP2CorporateActionMappingState(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	mappingID, expectedStatus, expectedCompleted, expectedRetry int64,
) {
	t.Helper()
	var status, total, completed, failed, retry int64
	if err := db.QueryRowContext(ctx, `SELECT status,position_total,position_completed,position_failed,retry_count
		FROM t_option_corporate_action_contract WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, mappingID,
	).Scan(&status, &total, &completed, &failed, &retry); err != nil {
		t.Fatal(err)
	}
	if status != expectedStatus || total != p2CorporateActionPositionCount ||
		completed != expectedCompleted || failed != 0 || retry != expectedRetry {
		t.Fatalf("corporate mapping status/total/completed/failed/retry=%d/%d/%d/%d/%d",
			status, total, completed, failed, retry)
	}
}

func assertP2CorporateActionRiskRestricted(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	var status, lastCalc int64
	if err := db.QueryRowContext(ctx, `SELECT status,last_calc_time FROM t_option_risk_account
		WHERE tenant_id=? AND user_id=? AND account_id=0 AND settle_coin='USDT'`,
		p0AssetE2ETenantID, p2CorporateActionUserID,
	).Scan(&status, &lastCalc); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED) || lastCalc != 0 {
		t.Fatalf("corporate action risk status/last_calc=%d/%d", status, lastCalc)
	}
}

func assertP2CorporateActionAssetEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	firstFreezeBiz, blockedFreezeBiz string,
) {
	t.Helper()
	assertWalletAmounts(t, ctx, db, p2CorporateActionUserID,
		"200.000000000000000000", "100.000000000000000000", "100.000000000000000000")
	var freezes, flows, instructions int64
	var amount, remain decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(amount),0),COALESCE(SUM(remain_amount),0)
		FROM t_asset_freeze WHERE tenant_id=? AND user_id=? AND biz_no IN (?,?)`,
		p0AssetE2ETenantID, p2CorporateActionUserID, firstFreezeBiz, blockedFreezeBiz,
	).Scan(&freezes, &amount, &remain); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND user_id=? AND biz_no IN (?,?)`,
		p0AssetE2ETenantID, p2CorporateActionUserID, firstFreezeBiz, blockedFreezeBiz,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_asset_instruction
		WHERE tenant_id=? AND user_id=?`,
		p0AssetE2ETenantID, p2CorporateActionUserID,
	).Scan(&instructions); err != nil {
		t.Fatal(err)
	}
	if freezes != 2 || flows != 2 || instructions != 0 ||
		!amount.Equal(decimal.NewFromInt(100)) || !remain.Equal(decimal.NewFromInt(100)) {
		t.Fatalf("corporate action freeze/amount/remain/flows/instructions=%d/%s/%s/%d/%d",
			freezes, amount, remain, flows, instructions)
	}
}

func assertP2CorporateActionFinalEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	actionID, mappingID, haltID int64,
	source, successor *models.TOptionContract,
	firstShortID, blockedShortID, firstLotID, blockedLotID int64,
	firstFreezeBiz, blockedFreezeBiz string,
	processCalls int,
) {
	t.Helper()
	var mappingStatus, positionTotal, positionCompleted, positionFailed, retryCount, lastPositionID int64
	if err := db.QueryRowContext(ctx, `SELECT status,position_total,position_completed,position_failed,retry_count,last_position_id
		FROM t_option_corporate_action_contract WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, mappingID,
	).Scan(&mappingStatus, &positionTotal, &positionCompleted, &positionFailed, &retryCount, &lastPositionID); err != nil {
		t.Fatal(err)
	}
	var maxSourceID int64
	if err := db.QueryRowContext(ctx, `SELECT MAX(id) FROM t_option_position
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, source.Id).Scan(&maxSourceID); err != nil {
		t.Fatal(err)
	}
	if mappingStatus != int64(option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_COMPLETED) ||
		positionTotal != p2CorporateActionPositionCount || positionCompleted != positionTotal ||
		positionFailed != 0 || retryCount != 1 || lastPositionID != maxSourceID {
		t.Fatalf("final mapping status/total/completed/failed/retry/cursor/max=%d/%d/%d/%d/%d/%d/%d",
			mappingStatus, positionTotal, positionCompleted, positionFailed, retryCount, lastPositionID, maxSourceID)
	}

	var actionStatus, sourceStatus, successorStatus, haltStatus int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT status FROM t_option_corporate_action WHERE tenant_id=? AND id=?),
		(SELECT status FROM t_option_contract WHERE tenant_id=? AND id=?),
		(SELECT status FROM t_option_contract WHERE tenant_id=? AND id=?),
		(SELECT status FROM t_option_trading_halt WHERE tenant_id=? AND id=?)`,
		p0AssetE2ETenantID, actionID,
		p0AssetE2ETenantID, source.Id,
		p0AssetE2ETenantID, successor.Id,
		p0AssetE2ETenantID, haltID,
	).Scan(&actionStatus, &sourceStatus, &successorStatus, &haltStatus); err != nil {
		t.Fatal(err)
	}
	if actionStatus != int64(option.CorporateActionStatus_CORPORATE_ACTION_STATUS_COMPLETED) ||
		sourceStatus != int64(option.ContractStatus_CONTRACT_STATUS_OFFLINE) ||
		successorStatus != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
		haltStatus != int64(option.TradingHaltStatus_TRADING_HALT_STATUS_LIFTED) {
		t.Fatalf("final action/source/successor/halt statuses=%d/%d/%d/%d",
			actionStatus, sourceStatus, successorStatus, haltStatus)
	}

	var sourceMigrated, successorHolding, audits, uniqueSources, marginAudits int64
	var sourceQty, successorQty, costBefore, costAfter decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(position_qty),0)
		FROM t_option_position WHERE tenant_id=? AND contract_id=? AND status=?`,
		p0AssetE2ETenantID, source.Id, int64(option.PositionStatus_POSITION_STATUS_MIGRATED),
	).Scan(&sourceMigrated, &sourceQty); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(position_qty),0)
		FROM t_option_position WHERE tenant_id=? AND contract_id=? AND status=?`,
		p0AssetE2ETenantID, successor.Id, int64(option.PositionStatus_POSITION_STATUS_HOLDING),
	).Scan(&successorHolding, &successorQty); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(DISTINCT source_position_id),
		COALESCE(SUM(cost_basis_before),0),COALESCE(SUM(cost_basis_after),0)
		FROM t_option_corporate_action_position WHERE tenant_id=? AND action_contract_id=?`,
		p0AssetE2ETenantID, mappingID,
	).Scan(&audits, &uniqueSources, &costBefore, &costAfter); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_corporate_action_margin_lot
		WHERE tenant_id=? AND action_position_id IN (
			SELECT id FROM t_option_corporate_action_position
			WHERE tenant_id=? AND action_contract_id=?)`,
		p0AssetE2ETenantID, p0AssetE2ETenantID, mappingID,
	).Scan(&marginAudits); err != nil {
		t.Fatal(err)
	}
	wantCost := decimal.NewFromInt(p2CorporateActionPositionCount * 10)
	if sourceMigrated != p2CorporateActionPositionCount || successorHolding != p2CorporateActionPositionCount ||
		audits != p2CorporateActionPositionCount || uniqueSources != audits || marginAudits != 2 ||
		!sourceQty.Equal(decimal.NewFromInt(p2CorporateActionPositionCount)) ||
		!successorQty.Equal(decimal.NewFromInt(p2CorporateActionPositionCount*2)) ||
		!costBefore.Equal(wantCost) || !costAfter.Equal(wantCost) {
		t.Fatalf("corporate action migration source/successor/audits/unique/margins qty=%d/%d/%d/%d/%d %s/%s cost=%s/%s",
			sourceMigrated, successorHolding, audits, uniqueSources, marginAudits,
			sourceQty, successorQty, costBefore, costAfter)
	}

	var migratedLots, preservedOrigins, preservedFreezeBiz int64
	var lotQty, remainingQty, initialMargin, remainingMargin, pendingMargin decimal.Decimal
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),
		COALESCE(SUM(origin_contract_id=? AND origin_position_id IN (?,?)),0),
		COALESCE(SUM(freeze_biz_no IN (?,?)),0),
		COALESCE(SUM(quantity),0),COALESCE(SUM(remaining_quantity),0),
		COALESCE(SUM(initial_margin),0),COALESCE(SUM(remaining_margin),0),COALESCE(SUM(pending_margin),0)
		FROM t_option_margin_lot WHERE tenant_id=? AND id IN (?,?) AND contract_id=?`,
		source.Id, firstShortID, blockedShortID, firstFreezeBiz, blockedFreezeBiz,
		p0AssetE2ETenantID, firstLotID, blockedLotID, successor.Id,
	).Scan(&migratedLots, &preservedOrigins, &preservedFreezeBiz, &lotQty, &remainingQty,
		&initialMargin, &remainingMargin, &pendingMargin); err != nil {
		t.Fatal(err)
	}
	if migratedLots != 2 || preservedOrigins != 2 || preservedFreezeBiz != 2 ||
		!lotQty.Equal(decimal.NewFromInt(4)) || !remainingQty.Equal(decimal.NewFromInt(4)) ||
		!initialMargin.Equal(decimal.NewFromInt(100)) || !remainingMargin.Equal(decimal.NewFromInt(100)) ||
		!pendingMargin.IsZero() {
		t.Fatalf("corporate action lots count/origin/freeze qty/remain/margin/pending=%d/%d/%d %s/%s/%s/%s/%s",
			migratedLots, preservedOrigins, preservedFreezeBiz, lotQty, remainingQty,
			initialMargin, remainingMargin, pendingMargin)
	}
	assertP2CorporateActionAssetEvidence(
		t, ctx, db, firstFreezeBiz, blockedFreezeBiz,
	)

	if _, err := db.ExecContext(ctx, `UPDATE t_option_corporate_action_position
		SET cost_basis_after=cost_basis_after+1 WHERE tenant_id=? AND action_contract_id=? LIMIT 1`,
		p0AssetE2ETenantID, mappingID,
	); err == nil {
		t.Fatal("corporate action position evidence was mutable")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_corporate_action_position
		WHERE tenant_id=? AND action_contract_id=? LIMIT 1`, p0AssetE2ETenantID, mappingID); err == nil {
		t.Fatal("corporate action position evidence was deletable")
	}
	t.Logf("corporate_action_capacity=positions:%d restart_calls:%d retry:%d source_qty:%s successor_qty:%s frozen:100 flows:2",
		positionCompleted, processCalls, retryCount, sourceQty, successorQty)
}

func testP2CorporateActionManualOnly(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	source := insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2-CORPORATE-ACTION-MANUAL-MERGER", calendarCode, 96799, now,
	)
	created, err := adminlogic.NewCreateCorporateActionLogic(
		p0AdminContext(ctx, 96796, p0AssetE2ETenantID), serviceCtx,
	).CreateCorporateAction(&option.CreateCorporateActionReq{
		TenantId: p0AssetE2ETenantID, EventNo: "P2-CORPORATE-ACTION-MANUAL-ONLY",
		ExternalEventRef: "P2-EXTERNAL-MERGER-MANUAL", UnderlyingSymbol: source.UnderlyingSymbol,
		ActionType:       option.CorporateActionType_CORPORATE_ACTION_TYPE_MERGER,
		AnnouncementTime: now - 3600, EffectiveTime: now - 1,
		EvidenceRef: "P2-MANUAL-MERGER-NOTICE", EvidenceHash: strings.Repeat("b", 64),
		Description: "unsupported basket economics must remain manual",
		Contracts: []*option.CorporateActionContractInput{{
			SourceContractId:  source.Id,
			ExecutionMode:     option.CorporateActionExecutionMode_CORPORATE_ACTION_EXECUTION_MODE_MANUAL_ONLY,
			QuantityNumerator: "1", QuantityDenominator: "1",
		}},
	})
	assertP2CorporateActionResponseOK(t, created, err, "create manual-only")
	reviewed, err := adminlogic.NewReviewCorporateActionLogic(
		p0AdminContext(ctx, 96795, p0AssetE2ETenantID), serviceCtx,
	).ReviewCorporateAction(&option.ReviewCorporateActionReq{
		TenantId: p0AssetE2ETenantID, ActionId: created.Data.Id, Approve: true,
		Reason: "manual-only economics acknowledged",
	})
	assertP2CorporateActionResponseOK(t, reviewed, err, "review manual-only")
	if reviewed.Data.Status != option.CorporateActionStatus_CORPORATE_ACTION_STATUS_MANUAL_REVIEW ||
		len(reviewed.Data.Contracts) != 1 ||
		reviewed.Data.Contracts[0].Status != option.CorporateActionContractStatus_CORPORATE_ACTION_CONTRACT_STATUS_MANUAL_REVIEW {
		t.Fatalf("manual-only corporate action escaped manual review: %+v", reviewed.Data)
	}
	if _, err = NewProcessCorporateActionsLogic(ctx, serviceCtx).
		ProcessCorporateActions(&option.OptionTaskReq{TenantId: p0AssetE2ETenantID}); err != nil {
		t.Fatal(err)
	}
	var sourceStatus, successorPositions int64
	if err = db.QueryRowContext(ctx, `SELECT
		(SELECT status FROM t_option_contract WHERE tenant_id=? AND id=?),
		(SELECT COUNT(*) FROM t_option_corporate_action_position WHERE tenant_id=? AND action_id=?)`,
		p0AssetE2ETenantID, source.Id, p0AssetE2ETenantID, created.Data.Id,
	).Scan(&sourceStatus, &successorPositions); err != nil {
		t.Fatal(err)
	}
	if sourceStatus != int64(option.ContractStatus_CONTRACT_STATUS_PAUSED) || successorPositions != 0 {
		t.Fatalf("manual-only source/status migrated unexpectedly=%d/%d", sourceStatus, successorPositions)
	}
}

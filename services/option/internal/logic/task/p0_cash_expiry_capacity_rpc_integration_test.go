package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	gosqlx "github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestP0CashExpiryCapacityAssetRPC proves the remaining repository-owned
// OPT-P0-005 boundaries in one governed cash-expiry batch:
//   - 502 positions cross the 500-row position page;
//   - 501 distinct long accounts include one 0.5 partial quantity;
//   - the 10% exercise fee is split from every long credit;
//   - the gross short debit equals net long credits plus fee credits;
//   - replay creates no additional instruction or Asset flow.
func TestP0CashExpiryCapacityAssetRPC(t *testing.T) {
	gosqlx.DisableLog()
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	rpcAddr := os.Getenv("OPTION_P0_ASSET_E2E_RPC_ADDR")
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || rpcAddr == "" || redisAddr == "" {
		t.Skip("OPTION_P0_ASSET_E2E_DSN, OPTION_P0_ASSET_E2E_RPC_ADDR and OPTION_P0_ASSET_E2E_REDIS_ADDR are required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping acceptance database: %v", err)
	}

	grpcConn, err := grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("connect Asset RPC: %v", err)
	}
	defer grpcConn.Close()
	assetClient := asset.NewAssetClient(grpcConn)
	waitForAssetRPC(t, ctx, assetClient)
	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, assetClient)

	started := time.Now()
	now := time.Now().Unix()
	contract := insertP0ExerciseContract(
		t, ctx, serviceCtx, "P0-CASH-EXPIRY-CAPACITY-501",
		option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
		option.ContractStatus_CONTRACT_STATUS_EXPIRED,
		now-3600, now-10, now-10, now-1,
		common.YesNo_YES_NO_NO, 29999, 92999,
	)

	const (
		longCount  = 501
		longBaseID = int64(21000)
		shortUser  = int64(22000)
	)
	partialQty := decimal.RequireFromString("0.5")
	oneQty := decimal.NewFromInt(1)
	firstLongPositionID := int64(0)
	for i := 0; i < longCount; i++ {
		qty := oneQty
		if i == 0 {
			qty = partialQty
		}
		position := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
			TenantId: p0AssetE2ETenantID, UserId: longBaseID + int64(i), AccountId: 31000 + int64(i),
			ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
			Side: int64(common.PositionSide_POSITION_SIDE_LONG), PositionQty: qty,
			AvailableQty: qty, OpenAvgPrice: decimal.NewFromInt(10), ExerciseableQty: qty,
			Status:      int64(option.PositionStatus_POSITION_STATUS_EXERCISED),
			CreateTimes: now - 1000 + int64(i), UpdateTimes: now - 1000 + int64(i),
		})
		if i == 0 {
			firstLongPositionID = position.Id
		}
	}
	seedDuration := time.Since(started)

	totalQty := decimal.RequireFromString("500.5")
	shortPosition := insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: shortUser, AccountId: 42000,
		ContractId: contract.Id, UnderlyingSymbol: "BTCUSDT",
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: totalQty,
		AvailableQty: totalQty, OpenAvgPrice: decimal.NewFromInt(10),
		MarginAmount: decimal.NewFromInt(11000), MaintenanceMargin: decimal.NewFromInt(4004),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: now - 1500, UpdateTimes: now - 1500,
	})
	lot := insertP0ExerciseMarginLot(
		t, ctx, serviceCtx, shortPosition, "P0-CASH-EXPIRY-CAPACITY-SHORT-MARGIN",
		"500.5", "11000", now-1400,
	)
	creditAsset(t, ctx, assetClient, shortUser, "12000", "P0-CASH-EXPIRY-CAPACITY-SHORT-SEED")
	freezeP0ExerciseMargin(t, ctx, assetClient, shortPosition, lot, "11000")

	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contract.Id, contract.ExpireTime, now,
		fmt.Sprintf("P0-CASH-EXPIRY-CAPACITY-%d", contract.Id),
		[]string{"119", "120", "121"}, "120",
	)
	clearingStarted := time.Now()
	logic := NewProcessContractLifecycleLogic(ctx, serviceCtx)
	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("create capacity cash expiry: %v", err)
	}
	clearingDuration := time.Since(clearingStarted)

	assetStarted := time.Now()
	processAssetInstructions(t, ctx, serviceCtx)
	processAssetInstructions(t, ctx, serviceCtx)
	assetDuration := time.Since(assetStarted)
	assertP0CashExpiryCapacity(
		t, ctx, db, contract.Id, firstLongPositionID, shortPosition.Id, lot.Id,
	)

	if err := logic.processExpiredContracts(now); err != nil {
		t.Fatalf("replay capacity cash expiry: %v", err)
	}
	processAssetInstructions(t, ctx, serviceCtx)
	assertP0CashExpiryCapacity(
		t, ctx, db, contract.Id, firstLongPositionID, shortPosition.Id, lot.Id,
	)

	t.Logf(
		"cash_expiry_capacity_longs=%d positions=%d seed=%s clearing=%s asset_rpc=%s instructions=%d",
		longCount, longCount+1, seedDuration.Round(time.Millisecond), clearingDuration.Round(time.Millisecond),
		assetDuration.Round(time.Millisecond), 1004,
	)
}

func assertP0CashExpiryCapacity(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, partialLongPositionID, shortPositionID, marginLotID int64,
) {
	t.Helper()
	var (
		settlementStatus, batchStatus, contractStatus                 int64
		batchInstructionCount, instructionCount, success, reconciled  int64
		flows                                                         int64
		positionCount, settledPositions, accountCount, detailCount    int64
		longCredits, feeCredits, shortDebits, releases                int64
		totalCredit, totalDebit, longCreditAmount, feeAmount          string
		shortDebitAmount, releaseAmount, walletTotal, walletAvailable string
		walletFrozen, remainingMargin, pendingMargin                  string
	)
	if err := db.QueryRowContext(ctx, `SELECT s.status,b.status,c.status,b.instruction_count,
		CAST(b.total_credit AS CHAR),CAST(b.total_debit AS CHAR),
		COUNT(DISTINCT i.id),SUM(i.status=3),SUM(i.reconciliation_status=2),COUNT(DISTINCT f.id)
		FROM t_option_settlement s
		JOIN t_option_settlement_batch b ON b.tenant_id=s.tenant_id AND b.batch_no=s.settlement_no
		JOIN t_option_contract c ON c.id=s.contract_id
		JOIN t_option_asset_instruction i ON i.tenant_id=s.tenant_id AND i.biz_no=s.settlement_no
		LEFT JOIN t_asset_flow f ON f.tenant_id=i.tenant_id AND f.biz_no=i.instruction_no
		WHERE s.tenant_id=? AND s.contract_id=?
		GROUP BY s.status,b.status,c.status,b.instruction_count,b.total_credit,b.total_debit`,
		p0AssetE2ETenantID, contractID,
	).Scan(
		&settlementStatus, &batchStatus, &contractStatus, &batchInstructionCount,
		&totalCredit, &totalDebit, &instructionCount, &success, &reconciled, &flows,
	); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE) ||
		contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) ||
		batchInstructionCount != 1004 || instructionCount != 1004 || success != 1004 ||
		reconciled != 1004 || flows != 1004 ||
		totalCredit != "10010.0000000000000000" || totalDebit != "10010.0000000000000000" {
		t.Fatalf("cash capacity settlement=%d/%d/%d instructions=%d/%d/%d/%d flows=%d credit/debit=%s/%s",
			settlementStatus, batchStatus, contractStatus, batchInstructionCount, instructionCount,
			success, reconciled, flows,
			totalCredit, totalDebit)
	}

	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=5),COUNT(DISTINCT account_id)
		FROM t_option_position WHERE tenant_id=? AND contract_id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&positionCount, &settledPositions, &accountCount); err != nil {
		t.Fatal(err)
	}
	if positionCount != 502 || settledPositions != 502 || accountCount != 502 {
		t.Fatalf("cash capacity positions/settled/accounts=%d/%d/%d want=502/502/502",
			positionCount, settledPositions, accountCount)
	}

	var partialQty, partialPayoff string
	if err := db.QueryRowContext(ctx, `SELECT CAST(quantity AS CHAR),CAST(payoff AS CHAR)
		FROM t_option_settlement_detail WHERE tenant_id=? AND contract_id=? AND position_id=?`,
		p0AssetE2ETenantID, contractID, partialLongPositionID,
	).Scan(&partialQty, &partialPayoff); err != nil {
		t.Fatal(err)
	}
	if partialQty != "0.5000000000000000" || partialPayoff != "10.0000000000000000" {
		t.Fatalf("partial cash expiry quantity/payoff=%s/%s want=0.5/10", partialQty, partialPayoff)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_settlement_detail
		WHERE tenant_id=? AND contract_id=?`, p0AssetE2ETenantID, contractID).Scan(&detailCount); err != nil {
		t.Fatal(err)
	}
	if detailCount != 502 {
		t.Fatalf("cash capacity settlement details=%d want=502", detailCount)
	}

	if err := db.QueryRowContext(ctx, `SELECT
		SUM(action=4 AND user_id BETWEEN 21000 AND 21500),
		SUM(action=4 AND user_id=29999),
		SUM(action=2 AND user_id=22000),
		SUM(action=3 AND user_id=22000),
		CAST(SUM(CASE WHEN action=4 AND user_id BETWEEN 21000 AND 21500 THEN amount ELSE 0 END) AS CHAR),
		CAST(SUM(CASE WHEN action=4 AND user_id=29999 THEN amount ELSE 0 END) AS CHAR),
		CAST(SUM(CASE WHEN action=2 AND user_id=22000 THEN amount ELSE 0 END) AS CHAR),
		CAST(SUM(CASE WHEN action=3 AND user_id=22000 THEN amount ELSE 0 END) AS CHAR)
		FROM t_option_asset_instruction WHERE tenant_id=? AND biz_no=(
			SELECT settlement_no FROM t_option_settlement WHERE tenant_id=? AND contract_id=?)`,
		p0AssetE2ETenantID, p0AssetE2ETenantID, contractID,
	).Scan(
		&longCredits, &feeCredits, &shortDebits, &releases,
		&longCreditAmount, &feeAmount, &shortDebitAmount, &releaseAmount,
	); err != nil {
		t.Fatal(err)
	}
	if longCredits != 501 || feeCredits != 501 || shortDebits != 1 || releases != 1 ||
		longCreditAmount != "9009.0000000000000000" || feeAmount != "1001.0000000000000000" ||
		shortDebitAmount != "10010.0000000000000000" || releaseAmount != "990.0000000000000000" {
		t.Fatalf("cash capacity legs=%d/%d/%d/%d amounts=%s/%s/%s/%s",
			longCredits, feeCredits, shortDebits, releases,
			longCreditAmount, feeAmount, shortDebitAmount, releaseAmount)
	}

	if err := db.QueryRowContext(ctx, `SELECT CAST(SUM(total_amount) AS CHAR),
		CAST(SUM(available_amount) AS CHAR),CAST(SUM(frozen_amount) AS CHAR)
		FROM t_user_asset WHERE tenant_id=? AND wallet_type=? AND coin='USDT'
		  AND ((user_id BETWEEN 21000 AND 21500) OR user_id IN (22000,29999))`,
		p0AssetE2ETenantID, int64(common.WalletType_WALLET_TYPE_OPTION),
	).Scan(&walletTotal, &walletAvailable, &walletFrozen); err != nil {
		t.Fatal(err)
	}
	if walletTotal != "12000.000000000000000000" || walletAvailable != "12000.000000000000000000" ||
		walletFrozen != "0.000000000000000000" {
		t.Fatalf("cash capacity wallet total/available/frozen=%s/%s/%s",
			walletTotal, walletAvailable, walletFrozen)
	}
	if err := db.QueryRowContext(ctx, `SELECT CAST(remaining_margin AS CHAR),CAST(pending_margin AS CHAR)
		FROM t_option_margin_lot WHERE id=? AND position_id=?`, marginLotID, shortPositionID,
	).Scan(&remainingMargin, &pendingMargin); err != nil {
		t.Fatal(err)
	}
	if remainingMargin != "0.0000000000000000" || pendingMargin != "0.0000000000000000" {
		t.Fatalf("cash capacity margin remaining/pending=%s/%s", remainingMargin, pendingMargin)
	}
}

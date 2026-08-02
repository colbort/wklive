package assetlogic

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"wklive/proto/asset"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const backstopRPCTestCoin = "USDT"

func TestPlatformBackstopRPCLimitsMySQL(t *testing.T) {
	dsn := os.Getenv("ASSET_BACKSTOP_E2E_DSN")
	rpcAddr := os.Getenv("ASSET_BACKSTOP_E2E_RPC_ADDR")
	if dsn == "" || rpcAddr == "" {
		t.Skip("ASSET_BACKSTOP_E2E_DSN and ASSET_BACKSTOP_E2E_RPC_ADDR are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err = db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	conn, err := grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := asset.NewAssetClient(conn)

	t.Run("missing draft rejected expired and disabled fail closed", func(t *testing.T) {
		testBackstopPolicyStates(t, ctx, db, client)
	})
	t.Run("prefunded exact boundaries and replay", func(t *testing.T) {
		testBackstopPrefundedBoundaries(t, ctx, db, client)
	})
	t.Run("credit floor twenty request concurrency", func(t *testing.T) {
		testBackstopCreditConcurrency(t, ctx, db, client)
	})
	t.Run("version switch preserves utc daily usage", func(t *testing.T) {
		testBackstopVersionSwitch(t, ctx, db, client)
	})
}

func testBackstopPolicyStates(t *testing.T, ctx context.Context, db *sql.DB, client asset.AssetClient) {
	t.Helper()
	const (
		missingTenant  int64 = 997100
		draftTenant    int64 = 997101
		rejectedTenant int64 = 997102
		disabledTenant int64 = 997103
		expiredTenant  int64 = 997104
	)
	for _, tenantID := range []int64{missingTenant, draftTenant, rejectedTenant, disabledTenant, expiredTenant} {
		seedBackstopRPCAccount(t, ctx, db, tenantID, backstopRPCTestCoin, "0")
	}
	seedBackstopRPCPolicy(t, ctx, db, draftTenant, backstopRPCTestCoin, 1, 3, "1", "10", "-10", 1, 5*time.Second)
	seedBackstopRPCPolicy(t, ctx, db, rejectedTenant, backstopRPCTestCoin, 1, 3, "1", "10", "-10", 3, 5*time.Second)
	seedBackstopRPCPolicy(t, ctx, db, disabledTenant, backstopRPCTestCoin, 1, 1, "0", "0", "0", 2, 5*time.Second)
	seedBackstopRPCPolicy(t, ctx, db, expiredTenant, backstopRPCTestCoin, 1, 3, "1", "10", "-10", 2, 250*time.Millisecond)
	time.Sleep(350 * time.Millisecond)

	for _, tenantID := range []int64{missingTenant, draftTenant, rejectedTenant, disabledTenant, expiredTenant} {
		requireBackstopRPCPrecondition(t, func() error {
			_, err := coverBackstopRPC(ctx, client, tenantID, tenantID, fmt.Sprintf("BST-STATE-%d", tenantID), "1")
			return err
		}, fmt.Sprintf("tenant %d ineffective policy", tenantID))
		assertBackstopRPCFacts(t, ctx, db, tenantID, "0", "0", 0, 0)
	}
	requireBackstopRPCPrecondition(t, func() error {
		_, err := coverBackstopRPC(ctx, client, disabledTenant, 999001, "BST-CROSS-COIN", "1", "BTC")
		return err
	}, "policy/account scope across coin")
}

func testBackstopPrefundedBoundaries(t *testing.T, ctx context.Context, db *sql.DB, client asset.AssetClient) {
	t.Helper()
	const tenantID int64 = 997110
	seedBackstopRPCAccount(t, ctx, db, tenantID, backstopRPCTestCoin, "20")
	seedBackstopRPCPolicy(t, ctx, db, tenantID, backstopRPCTestCoin, 1, 2, "10", "20", "0", 2, 10*time.Second)
	time.Sleep(150 * time.Millisecond)
	requireBackstopRPCPrecondition(t, func() error {
		_, err := coverBackstopRPC(ctx, client, tenantID, 1, "BST-PREFUND-OVER", "10.000000000000000001")
		return err
	}, "per-request limit plus minimum unit")
	assertBackstopRPCFacts(t, ctx, db, tenantID, "20", "0", 0, 0)

	first, err := coverBackstopRPC(ctx, client, tenantID, 2, "BST-PREFUND-EXACT-1", "10")
	if err != nil {
		t.Fatal(err)
	}
	if first.GetPolicyMode() != asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_PREFUNDED ||
		first.GetDailyUsedAmount() != "10" || first.GetPlatformAccountBalance() != "10" {
		t.Fatalf("unexpected first boundary response: %+v", first)
	}
	replay, err := coverBackstopRPC(ctx, client, tenantID, 2, "BST-PREFUND-EXACT-1", "10")
	if err != nil || !replay.GetIdempotentReplay() || replay.GetPolicyId() != first.GetPolicyId() ||
		replay.GetDailyUsedAmount() != first.GetDailyUsedAmount() ||
		replay.GetPlatformAccountBalance() != first.GetPlatformAccountBalance() {
		t.Fatalf("response replay changed snapshot response=%+v err=%v", replay, err)
	}
	if _, err = coverBackstopRPC(ctx, client, tenantID, 2, "BST-PREFUND-EXACT-1", "9"); err == nil {
		t.Fatal("idempotency key accepted changed amount")
	}
	if _, err = coverBackstopRPC(ctx, client, tenantID, 3, "BST-PREFUND-EXACT-2", "10"); err != nil {
		t.Fatal(err)
	}
	if _, err = coverBackstopRPC(ctx, client, tenantID, 4, "BST-PREFUND-DAILY-OVER", "0.000000000000000001"); err == nil {
		t.Fatal("daily limit plus minimum unit was accepted")
	}
	assertBackstopRPCFacts(t, ctx, db, tenantID, "0", "20", 2, 2)
	if _, err = db.ExecContext(ctx, `UPDATE t_asset_platform_account
		SET available_amount=available_amount+10,version=version+1
		WHERE tenant_id=? AND account_type='OPTION_BACKSTOP' AND coin=?`, tenantID, backstopRPCTestCoin); err != nil {
		t.Fatal(err)
	}
	if _, err = coverBackstopRPC(ctx, client, tenantID, 5, "BST-PREFUND-TOPUP-NO-RESET", "0.000000000000000001"); err == nil {
		t.Fatal("top-up reset or enlarged the utc daily limit")
	}
	assertBackstopRPCFacts(t, ctx, db, tenantID, "10", "20", 2, 2)
}

func testBackstopCreditConcurrency(t *testing.T, ctx context.Context, db *sql.DB, client asset.AssetClient) {
	t.Helper()
	const tenantID int64 = 997120
	seedBackstopRPCAccount(t, ctx, db, tenantID, backstopRPCTestCoin, "0")
	seedBackstopRPCPolicy(t, ctx, db, tenantID, backstopRPCTestCoin, 1, 3, "1", "10", "-10", 2, 10*time.Second)
	time.Sleep(150 * time.Millisecond)
	type result struct {
		requestNo string
		resp      *asset.CoverPlatformBackstopDeficitResp
		err       error
	}
	start := make(chan struct{})
	results := make(chan result, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			requestNo := fmt.Sprintf("BST-CONCURRENT-%02d", index)
			resp, err := coverBackstopRPC(ctx, client, tenantID, int64(100+index), requestNo, "1")
			results <- result{requestNo: requestNo, resp: resp, err: err}
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	succeeded, rejected := 0, 0
	var replayRequest string
	var replayLiquidation int64
	for item := range results {
		if item.err != nil {
			rejected++
			continue
		}
		succeeded++
		if item.resp.GetPolicyMode() != asset.PlatformBackstopMode_PLATFORM_BACKSTOP_MODE_CREDIT_FLOOR {
			t.Fatalf("unexpected credit policy response: %+v", item.resp)
		}
		if replayRequest == "" {
			replayRequest = item.requestNo
			_, _ = fmt.Sscanf(item.requestNo, "BST-CONCURRENT-%d", &replayLiquidation)
		}
	}
	if succeeded != 10 || rejected != 10 {
		t.Fatalf("concurrent success/rejected=%d/%d want=10/10", succeeded, rejected)
	}
	assertBackstopRPCFacts(t, ctx, db, tenantID, "-10", "10", 10, 10)

	// Resolve the successful request's liquidation ID from durable evidence so replay
	// verifies the exact original key and snapshot after all concurrent commits.
	if err := db.QueryRowContext(ctx, `SELECT liquidation_id FROM t_asset_backstop_cover
		WHERE tenant_id=? AND liquidation_no=?`, tenantID, replayRequest).Scan(&replayLiquidation); err != nil {
		t.Fatal(err)
	}
	replay, err := coverBackstopRPC(ctx, client, tenantID, replayLiquidation, replayRequest, "1")
	if err != nil || !replay.GetIdempotentReplay() {
		t.Fatalf("concurrent result replay failed response=%+v err=%v", replay, err)
	}
	assertBackstopRPCFacts(t, ctx, db, tenantID, "-10", "10", 10, 10)
}

func testBackstopVersionSwitch(t *testing.T, ctx context.Context, db *sql.DB, client asset.AssetClient) {
	t.Helper()
	const tenantID int64 = 997130
	seedBackstopRPCAccount(t, ctx, db, tenantID, backstopRPCTestCoin, "0")
	firstPolicy := seedBackstopRPCPolicy(t, ctx, db, tenantID, backstopRPCTestCoin, 1, 3, "4", "10", "-20", 2, 10*time.Second)
	time.Sleep(150 * time.Millisecond)
	first, err := coverBackstopRPC(ctx, client, tenantID, 1, "BST-VERSION-1", "4")
	if err != nil || first.GetPolicyId() != firstPolicy {
		t.Fatalf("version one cover response=%+v err=%v", first, err)
	}
	secondPolicy := seedBackstopRPCPolicy(t, ctx, db, tenantID, backstopRPCTestCoin, 2, 3, "6", "10", "-20", 2, 10*time.Second)
	time.Sleep(150 * time.Millisecond)
	second, err := coverBackstopRPC(ctx, client, tenantID, 2, "BST-VERSION-2", "6")
	if err != nil || second.GetPolicyId() != secondPolicy || second.GetDailyUsedAmount() != "10" {
		t.Fatalf("version two did not preserve daily usage response=%+v err=%v", second, err)
	}
	if _, err = coverBackstopRPC(ctx, client, tenantID, 3, "BST-VERSION-OVER", "0.000000000000000001"); err == nil {
		t.Fatal("new version reset or enlarged utc daily usage")
	}
	assertBackstopRPCFacts(t, ctx, db, tenantID, "-10", "10", 2, 2)
	var distinctPolicies int64
	if err = db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT policy_id) FROM t_asset_backstop_cover WHERE tenant_id=?`, tenantID).Scan(&distinctPolicies); err != nil {
		t.Fatal(err)
	}
	if distinctPolicies != 2 {
		t.Fatalf("version evidence count=%d want=2", distinctPolicies)
	}
}

func seedBackstopRPCAccount(t *testing.T, ctx context.Context, db *sql.DB, tenantID int64, coin, amount string) {
	t.Helper()
	now := time.Now().UnixMilli()
	if _, err := db.ExecContext(ctx, `INSERT INTO t_asset_platform_account
		(tenant_id,account_type,coin,available_amount,frozen_amount,status,version,create_times,update_times)
		VALUES (?,'OPTION_BACKSTOP',?,?,0,1,0,?,?)`, tenantID, coin, amount, now, now); err != nil {
		t.Fatal(err)
	}
}

func seedBackstopRPCPolicy(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	tenantID int64,
	coin string,
	version, mode int64,
	perRequest, daily, floor string,
	status int64,
	duration time.Duration,
) int64 {
	t.Helper()
	now := time.Now().UnixMilli()
	effectiveFrom := now + 100
	effectiveUntil := effectiveFrom + duration.Milliseconds()
	requestNo := fmt.Sprintf("BST-POLICY-%d-%s-%d", tenantID, coin, version)
	result, err := db.ExecContext(ctx, `INSERT INTO t_asset_backstop_policy
		(tenant_id,coin,request_no,version,mode,per_request_limit,daily_limit,balance_floor,
		 effective_from,effective_until,status,reason,evidence_ref,created_by,reviewed_by,
		 review_reason,create_times,update_times)
		VALUES (?,?,?,?,?,?,?,?,?,?,1,'rpc limit acceptance','test://platform-backstop-rpc',
		 9101,0,'',?,?)`, tenantID, coin, requestNo, version, mode, perRequest, daily, floor,
		effectiveFrom, effectiveUntil, now, now)
	if err != nil {
		t.Fatal(err)
	}
	policyID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	if status == 2 || status == 3 {
		if _, err = db.ExecContext(ctx, `UPDATE t_asset_backstop_policy
			SET status=?,reviewed_by=9102,review_reason='independent rpc acceptance review',update_times=?
			WHERE id=?`, status, time.Now().UnixMilli(), policyID); err != nil {
			t.Fatal(err)
		}
	}
	return policyID
}

func coverBackstopRPC(
	ctx context.Context,
	client asset.AssetClient,
	tenantID, liquidationID int64,
	requestNo, amount string,
	coinOverride ...string,
) (*asset.CoverPlatformBackstopDeficitResp, error) {
	coin := backstopRPCTestCoin
	if len(coinOverride) > 0 {
		coin = coinOverride[0]
	}
	return client.CoverPlatformBackstopDeficit(ctx, &asset.CoverPlatformBackstopDeficitReq{
		TenantId: tenantID, Coin: coin, LiquidationId: liquidationID,
		LiquidationNo: requestNo, RequestedAmount: amount, Remark: "platform backstop RPC acceptance",
	})
}

func requireBackstopRPCPrecondition(t *testing.T, call func() error, scenario string) {
	t.Helper()
	err := call()
	if err == nil {
		t.Fatalf("%s was accepted", scenario)
	}
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("%s rejection code=%s want=%s err=%v", scenario, status.Code(err), codes.FailedPrecondition, err)
	}
}

func assertBackstopRPCFacts(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	tenantID int64,
	wantBalance, wantDaily string,
	wantCovers, wantFlows int64,
) {
	t.Helper()
	var balance, daily decimal.Decimal
	var covers, flows int64
	if err := db.QueryRowContext(ctx, `SELECT available_amount FROM t_asset_platform_account
		WHERE tenant_id=? AND account_type='OPTION_BACKSTOP' AND coin=?`, tenantID, backstopRPCTestCoin).Scan(&balance); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(SUM(covered_amount),0) FROM t_asset_backstop_usage_daily
		WHERE tenant_id=? AND coin=?`, tenantID, backstopRPCTestCoin).Scan(&daily); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_backstop_cover WHERE tenant_id=?`, tenantID).Scan(&covers); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_platform_flow
		WHERE tenant_id=? AND account_type='OPTION_BACKSTOP' AND scene_type='platform_backstop_cover'`, tenantID).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if !balance.Equal(decimal.RequireFromString(wantBalance)) ||
		!daily.Equal(decimal.RequireFromString(wantDaily)) || covers != wantCovers || flows != wantFlows {
		t.Fatalf("facts balance/daily/covers/flows=%s/%s/%d/%d want=%s/%s/%d/%d",
			balance, daily, covers, flows, wantBalance, wantDaily, wantCovers, wantFlows)
	}
}

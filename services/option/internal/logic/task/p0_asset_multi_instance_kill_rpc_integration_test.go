package tasklogic

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/option"
	"wklive/services/option/models"

	_ "github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	p0MultiInstanceWorkerEnv = "OPTION_P0_MULTI_INSTANCE_WORKER"
	p0MultiInstanceRPCEnv    = "OPTION_P0_ASSET_E2E_RPC_ADDR"
	p0MultiInstanceLockKey   = "option:task:process_asset_instructions"
)

// TestP0AssetMultiInstanceKillTakeover proves the process boundary that the
// in-process stale-recovery test cannot cover: Asset commits a freeze, the
// Option worker is SIGKILLed before receiving the response, its Redis lease
// expires naturally, and two fresh worker processes compete to take over.
func TestP0AssetMultiInstanceKillTakeover(t *testing.T) {
	if os.Getenv(p0MultiInstanceWorkerEnv) == "1" {
		t.Skip("parent-only multi-instance acceptance")
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	directRPCAddr := os.Getenv(p0MultiInstanceRPCEnv)
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || directRPCAddr == "" || redisAddr == "" {
		t.Skip("Option P0 Asset E2E environment is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 75*time.Second)
	defer cancel()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping acceptance database: %v", err)
	}
	directConn, err := grpc.NewClient(
		directRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("connect direct Asset RPC: %v", err)
	}
	defer directConn.Close()
	directAsset := asset.NewAssetClient(directConn)
	waitForAssetRPC(t, ctx, directAsset)
	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, directAsset)

	locked, err := serviceCtx.Redis.ExistsCtx(ctx, p0MultiInstanceLockKey)
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Fatalf("task lock %s was left by an earlier scenario", p0MultiInstanceLockKey)
	}

	const (
		userID        int64 = 194
		accountID     int64 = 7069
		instructionNo       = "P0-MULTI-INSTANCE-KILL-FREEZE"
		freezeBizNo         = "P0-MULTI-INSTANCE-KILL-FREEZE-BIZ"
	)
	creditAsset(t, ctx, directAsset, userID, "100", "P0-MULTI-INSTANCE-KILL-SEED")
	now := time.Now().Unix()
	instruction := insertAssetInstruction(t, ctx, serviceCtx, &models.TOptionAssetInstruction{
		TenantId: p0AssetE2ETenantID, InstructionNo: instructionNo,
		BizNo: freezeBizNo, UserId: userID, AccountId: accountID,
		Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
		TargetBizNo: freezeBizNo, Coin: "USDT", Amount: decimal.NewFromInt(10),
		StepNo: 1, Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
		ReconciliationStatus: int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING),
		CreateTimes:          now, UpdateTimes: now,
	})

	victimProxy := newP0FreezeCommitProxy(t, directAsset)
	defer victimProxy.stop()
	victim := startP0AssetWorker(t, victimProxy.address())
	select {
	case <-victimProxy.committed:
	case <-ctx.Done():
		t.Fatalf("victim did not reach committed Asset freeze: %v", ctx.Err())
	}
	assertP0CommittedFreezeBeforeKill(
		t, ctx, db, instruction.Id, userID, freezeBizNo,
	)
	if err := victim.cmd.Process.Kill(); err != nil {
		t.Fatalf("SIGKILL victim worker: %v", err)
	}
	if err := victim.cmd.Wait(); err == nil {
		t.Fatal("SIGKILLed victim worker exited successfully")
	}
	victimProxy.stop()
	if victimProxy.calls.Load() != 1 {
		t.Fatalf("victim Asset calls=%d want=1", victimProxy.calls.Load())
	}

	blocked := runP0AssetWorker(t, directRPCAddr)
	if blocked.code == 200 {
		t.Fatalf("fresh worker bypassed the killed instance lease: %s", blocked.output)
	}
	startedWaiting := time.Now()
	waitP0TaskLeaseExpiry(t, ctx, serviceCtx.Redis)
	t.Logf("killed worker lease expired naturally after %s", time.Since(startedWaiting).Round(time.Millisecond))

	// The 30-second Redis lease has elapsed in wall-clock time. Accelerate only
	// the independent 60-second database stale-age threshold to keep this gate
	// bounded; the threshold itself has dedicated model boundary tests.
	result, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET update_times=? WHERE id=? AND status=?`,
		time.Now().Unix()-61, instruction.Id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING),
	)
	if err != nil {
		t.Fatalf("age killed instruction for takeover: %v", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		t.Fatalf("age killed instruction rows=%d err=%v", affected, err)
	}

	takeoverProxy := newP0FreezeCommitProxy(t, directAsset)
	defer takeoverProxy.stop()
	first := startP0AssetWorker(t, takeoverProxy.address())
	second := startP0AssetWorker(t, takeoverProxy.address())
	results := make(chan p0WorkerResult, 2)
	go func() { results <- waitP0AssetWorker(first) }()
	go func() { results <- waitP0AssetWorker(second) }()
	select {
	case <-takeoverProxy.committed:
	case <-ctx.Done():
		t.Fatalf("takeover worker did not replay committed Asset freeze: %v", ctx.Err())
	}

	var loser p0WorkerResult
	select {
	case loser = <-results:
	case <-ctx.Done():
		t.Fatalf("competing worker did not observe the active lease: %v", ctx.Err())
	}
	if loser.err != nil || loser.code == 200 {
		t.Fatalf("expected one clean lease rejection, code=%d err=%v output=%s",
			loser.code, loser.err, loser.output)
	}
	close(takeoverProxy.release)
	winner := <-results
	if winner.err != nil || winner.code != 200 {
		t.Fatalf("takeover worker code=%d err=%v output=%s", winner.code, winner.err, winner.output)
	}
	takeoverProxy.stop()
	if takeoverProxy.calls.Load() != 1 {
		t.Fatalf("competing takeover Asset calls=%d want=1", takeoverProxy.calls.Load())
	}

	assertP0MultiInstanceTakeoverEvidence(
		t, ctx, db, serviceCtx.Redis, instruction.Id, userID, freezeBizNo,
	)
}

// TestP0AssetMultiInstanceWorker is executed as a standalone copy of the Go
// test binary. It intentionally accepts either task-lock result; the parent
// coordinates two independent processes and asserts exactly one winner.
func TestP0AssetMultiInstanceWorker(t *testing.T) {
	if os.Getenv(p0MultiInstanceWorkerEnv) != "1" {
		t.Skip("worker helper is launched by the multi-instance acceptance")
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	rpcAddr := os.Getenv(p0MultiInstanceRPCEnv)
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || rpcAddr == "" || redisAddr == "" {
		t.Fatal("worker environment is incomplete")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(rpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	serviceCtx := newP0AssetE2EServiceContext(dsn, redisAddr, asset.NewAssetClient(conn))
	resp, err := NewProcessAssetInstructionsLogic(ctx, serviceCtx).ProcessAssetInstructions(
		&option.OptionTaskReq{TenantId: p0AssetE2ETenantID},
	)
	if err != nil {
		t.Fatal(err)
	}
	code := int32(0)
	if resp != nil && resp.Base != nil {
		code = resp.Base.Code
	}
	fmt.Printf("P0_MULTI_INSTANCE_WORKER_CODE=%d\n", code)
}

type p0FreezeCommitProxy struct {
	asset.UnimplementedAssetServer

	upstream  asset.AssetClient
	committed chan struct{}
	release   chan struct{}
	once      sync.Once
	calls     atomic.Int64
	listener  net.Listener
	server    *grpc.Server
}

func newP0FreezeCommitProxy(t *testing.T, upstream asset.AssetClient) *p0FreezeCommitProxy {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := &p0FreezeCommitProxy{
		upstream: upstream, committed: make(chan struct{}), release: make(chan struct{}),
		listener: listener, server: grpc.NewServer(),
	}
	asset.RegisterAssetServer(proxy.server, proxy)
	go func() {
		_ = proxy.server.Serve(listener)
	}()
	return proxy
}

func (p *p0FreezeCommitProxy) FreezeAsset(
	ctx context.Context,
	in *asset.FreezeAssetReq,
) (*asset.FreezeAssetResp, error) {
	p.calls.Add(1)
	resp, err := p.upstream.FreezeAsset(ctx, in)
	if err != nil {
		return nil, err
	}
	p.once.Do(func() { close(p.committed) })
	select {
	case <-p.release:
		return resp, nil
	case <-ctx.Done():
		return nil, status.FromContextError(ctx.Err()).Err()
	}
}

func (p *p0FreezeCommitProxy) GetAssetFlowByBizNo(
	ctx context.Context,
	in *asset.GetAssetFlowByBizNoReq,
) (*asset.GetAssetFlowByBizNoResp, error) {
	return p.upstream.GetAssetFlowByBizNo(ctx, in)
}

func (p *p0FreezeCommitProxy) address() string {
	return p.listener.Addr().String()
}

func (p *p0FreezeCommitProxy) stop() {
	if p.server != nil {
		p.server.Stop()
		p.server = nil
	}
}

type p0WorkerProcess struct {
	cmd    *exec.Cmd
	output *bytes.Buffer
}

type p0WorkerResult struct {
	code   int64
	err    error
	output string
}

func startP0AssetWorker(t *testing.T, rpcAddr string) *p0WorkerProcess {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	output := &bytes.Buffer{}
	cmd := exec.Command(executable,
		"-test.run=^TestP0AssetMultiInstanceWorker$", "-test.v", "-test.count=1",
	)
	cmd.Env = p0WorkerEnvironment(os.Environ(), rpcAddr)
	cmd.Stdout = output
	cmd.Stderr = output
	if err := cmd.Start(); err != nil {
		t.Fatalf("start Option worker: %v", err)
	}
	return &p0WorkerProcess{cmd: cmd, output: output}
}

func runP0AssetWorker(t *testing.T, rpcAddr string) p0WorkerResult {
	t.Helper()
	return waitP0AssetWorker(startP0AssetWorker(t, rpcAddr))
}

func waitP0AssetWorker(worker *p0WorkerProcess) p0WorkerResult {
	err := worker.cmd.Wait()
	output := worker.output.String()
	code := int64(-1)
	match := regexp.MustCompile(`P0_MULTI_INSTANCE_WORKER_CODE=(-?[0-9]+)`).FindStringSubmatch(output)
	if len(match) == 2 {
		if parsed, parseErr := strconv.ParseInt(match[1], 10, 64); parseErr == nil {
			code = parsed
		}
	}
	return p0WorkerResult{code: code, err: err, output: output}
}

func p0WorkerEnvironment(environment []string, rpcAddr string) []string {
	filtered := make([]string, 0, len(environment)+2)
	for _, item := range environment {
		if strings.HasPrefix(item, p0MultiInstanceWorkerEnv+"=") ||
			strings.HasPrefix(item, p0MultiInstanceRPCEnv+"=") {
			continue
		}
		filtered = append(filtered, item)
	}
	return append(filtered,
		p0MultiInstanceWorkerEnv+"=1",
		p0MultiInstanceRPCEnv+"="+rpcAddr,
	)
}

func waitP0TaskLeaseExpiry(
	t *testing.T,
	ctx context.Context,
	redis interface {
		ExistsCtx(context.Context, string) (bool, error)
	},
) {
	t.Helper()
	for {
		exists, err := redis.ExistsCtx(ctx, p0MultiInstanceLockKey)
		if err != nil {
			t.Fatalf("read task lock while waiting for expiry: %v", err)
		}
		if !exists {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("wait for killed worker lease expiry: %v", ctx.Err())
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func assertP0CommittedFreezeBeforeKill(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	instructionID, userID int64,
	freezeBizNo string,
) {
	t.Helper()
	var statusValue, retryCount, flows, freezes int64
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count
		FROM t_option_asset_instruction WHERE id=?`, instructionID,
	).Scan(&statusValue, &retryCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND user_id=? AND biz_no=?`,
		p0AssetE2ETenantID, userID, freezeBizNo,
	).Scan(&flows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_freeze
		WHERE tenant_id=? AND user_id=? AND biz_no=?`,
		p0AssetE2ETenantID, userID, freezeBizNo,
	).Scan(&freezes); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING) ||
		retryCount != 0 || flows != 1 || freezes != 1 {
		t.Fatalf("before kill status/retry/flows/freezes=%d/%d/%d/%d",
			statusValue, retryCount, flows, freezes)
	}
}

func assertP0MultiInstanceTakeoverEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	redis interface {
		ExistsCtx(context.Context, string) (bool, error)
	},
	instructionID, userID int64,
	freezeBizNo string,
) {
	t.Helper()
	var statusValue, retryCount, reconciled, flows, freezes, duplicateFlows int64
	var lastError string
	if err := db.QueryRowContext(ctx, `SELECT status,retry_count,reconciliation_status,last_error_msg
		FROM t_option_asset_instruction WHERE id=?`, instructionID,
	).Scan(&statusValue, &retryCount, &reconciled, &lastError); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(*)-COUNT(DISTINCT flow_no)
		FROM t_asset_flow WHERE tenant_id=? AND user_id=? AND biz_no=?`,
		p0AssetE2ETenantID, userID, freezeBizNo,
	).Scan(&flows, &duplicateFlows); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_freeze
		WHERE tenant_id=? AND user_id=? AND biz_no=?`,
		p0AssetE2ETenantID, userID, freezeBizNo,
	).Scan(&freezes); err != nil {
		t.Fatal(err)
	}
	if statusValue != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_SUCCESS) ||
		retryCount != 0 || reconciled != int64(option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_MATCHED) ||
		lastError != "" || flows != 1 || freezes != 1 || duplicateFlows != 0 {
		t.Fatalf("takeover status/retry/reconciled/error/flows/freezes/duplicate_flows=%d/%d/%d/%q/%d/%d/%d",
			statusValue, retryCount, reconciled, lastError, flows, freezes, duplicateFlows)
	}
	assertWalletAmounts(t, ctx, db, userID,
		"100.000000000000000000", "90.000000000000000000", "10.000000000000000000")
	locked, err := redis.ExistsCtx(ctx, p0MultiInstanceLockKey)
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Fatalf("task lock %s remained after takeover", p0MultiInstanceLockKey)
	}
}

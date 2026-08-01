package tasklogic

import (
	"context"
	"database/sql"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// TestP1PhysicalDeliveryProcessKillTakeover proves the physical-delivery
// process boundary: both debit legs finish, the first credit commits in Asset,
// and the Option worker is SIGKILLed before receiving the response. A fresh
// worker must reconcile the same instruction and finish the remaining credit
// without duplicating any debit or credit.
func TestP1PhysicalDeliveryProcessKillTakeover(t *testing.T) {
	if os.Getenv(p0MultiInstanceWorkerEnv) == "1" {
		t.Skip("parent-only physical delivery process-kill acceptance")
	}
	dsn := os.Getenv("OPTION_P0_ASSET_E2E_DSN")
	directRPCAddr := os.Getenv(p0MultiInstanceRPCEnv)
	redisAddr := os.Getenv("OPTION_P0_ASSET_E2E_REDIS_ADDR")
	if dsn == "" || directRPCAddr == "" || redisAddr == "" {
		t.Skip("Option P0 Asset E2E environment is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 85*time.Second)
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
	assertP1PhysicalTaskLockAbsent(t, ctx, serviceCtx.Redis)

	const (
		contractID  int64 = 999901
		longUserID  int64 = 3141
		shortUserID int64 = 3142
	)
	now := time.Now().Unix()
	prefix := "P1-PHYSICAL-CALL-CREDIT-PROCESS-KILL"
	seedP1PhysicalContract(
		t, ctx, db, contractID, prefix, option.OptionType_OPTION_TYPE_CALL, now-10, now-1,
	)
	creditAssetCoin(t, ctx, directAsset, longUserID, "USDT", "150", prefix+"-LONG-SEED")
	creditAssetCoin(t, ctx, directAsset, shortUserID, "BTC", "1", prefix+"-SHORT-SEED")
	insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, longUserID, longUserID,
		common.PositionSide_POSITION_SIDE_LONG, now-200, decimal.Zero)
	shortPosition := insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, shortUserID, shortUserID,
		common.PositionSide_POSITION_SIDE_SHORT, now-190, decimal.NewFromInt(1))
	lot := insertP1PhysicalMarginLot(
		t, ctx, serviceCtx, shortPosition, prefix+"-SHORT-COLLATERAL", "BTC", "1", now-180,
	)
	freezeP1PhysicalCollateral(t, ctx, directAsset, lot, "BTC", "1")
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, now-10, now, prefix, physicalEvidencePrices("120"), "120",
	)
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).processExpiredContracts(now); err != nil {
		t.Fatalf("create process-kill physical delivery: %v", err)
	}
	unit := findP1PhysicalUnitByLongUser(t, ctx, db, contractID, longUserID)
	instructions, err := serviceCtx.OptionAssetInstructionModel.FindByDeliveryUnit(
		ctx, p0AssetE2ETenantID, unit.Id,
	)
	if err != nil {
		t.Fatal(err)
	}
	var targetCredit, remainingCredit *models.TOptionAssetInstruction
	for _, instruction := range instructions {
		if instruction.Action != int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_CREDIT_AVAILABLE) {
			continue
		}
		if instruction.UserId == longUserID {
			targetCredit = instruction
		} else {
			remainingCredit = instruction
		}
	}
	if targetCredit == nil || remainingCredit == nil {
		t.Fatalf("physical process-kill credits target=%+v remaining=%+v", targetCredit, remainingCredit)
	}
	for wantDebitSuccess := int64(1); wantDebitSuccess <= 2; wantDebitSuccess++ {
		worker := runP0AssetWorker(t, directRPCAddr)
		if worker.err != nil || worker.code != 200 {
			t.Fatalf("physical debit worker %d code=%d err=%v output=%s",
				wantDebitSuccess, worker.code, worker.err, worker.output)
		}
		assertP1PhysicalDebitWorkerBoundary(t, ctx, db, unit.Id, wantDebitSuccess)
	}

	victimProxy := newP1PhysicalCreditCommitProxy(t, directAsset, targetCredit.InstructionNo)
	defer victimProxy.stop()
	victim := startP0AssetWorker(t, victimProxy.address())
	victimResults := make(chan p0WorkerResult, 1)
	go func() { victimResults <- waitP0AssetWorker(victim) }()
	select {
	case <-victimProxy.committed:
	case exited := <-victimResults:
		t.Fatalf("physical victim exited before committed credit, code=%d err=%v output=%s",
			exited.code, exited.err, exited.output)
	case <-ctx.Done():
		t.Fatalf("physical victim did not reach committed credit: %v output=%s", ctx.Err(), victim.output.String())
	}
	assertP1PhysicalCommittedCreditBeforeKill(
		t, ctx, db, unit.Id, targetCredit.Id, remainingCredit.Id,
	)
	if err := victim.cmd.Process.Kill(); err != nil {
		t.Fatalf("SIGKILL physical victim worker: %v", err)
	}
	if killed := <-victimResults; killed.err == nil {
		t.Fatal("SIGKILLed physical victim worker exited successfully")
	}
	victimProxy.stop()
	if victimProxy.calls.Load() != 1 {
		t.Fatalf("physical victim credit calls=%d want=1", victimProxy.calls.Load())
	}

	blocked := runP0AssetWorker(t, directRPCAddr)
	if blocked.code == 200 {
		t.Fatalf("fresh physical worker bypassed the killed instance lease: %s", blocked.output)
	}
	startedWaiting := time.Now()
	waitP0TaskLeaseExpiry(t, ctx, serviceCtx.Redis)
	t.Logf("physical killed worker lease expired naturally after %s", time.Since(startedWaiting).Round(time.Millisecond))

	result, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
		SET update_times=? WHERE id=? AND status=?`,
		time.Now().Unix()-61, targetCredit.Id,
		int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING),
	)
	if err != nil {
		t.Fatalf("age physical killed instruction for takeover: %v", err)
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		t.Fatalf("age physical killed instruction rows=%d err=%v", affected, rowsErr)
	}

	takeoverProxy := newP1PhysicalCreditCommitProxy(t, directAsset, targetCredit.InstructionNo)
	defer takeoverProxy.stop()
	first := startP0AssetWorker(t, takeoverProxy.address())
	second := startP0AssetWorker(t, takeoverProxy.address())
	results := make(chan p0WorkerResult, 2)
	go func() { results <- waitP0AssetWorker(first) }()
	go func() { results <- waitP0AssetWorker(second) }()
	select {
	case <-takeoverProxy.committed:
	case <-ctx.Done():
		t.Fatalf("physical takeover worker did not replay committed credit: %v", ctx.Err())
	}

	select {
	case loser := <-results:
		if loser.err != nil || loser.code == 200 {
			t.Fatalf("expected one clean physical lease rejection, code=%d err=%v output=%s",
				loser.code, loser.err, loser.output)
		}
	case <-ctx.Done():
		t.Fatalf("competing physical worker did not observe the active lease: %v", ctx.Err())
	}
	close(takeoverProxy.release)
	winner := <-results
	if winner.err != nil || winner.code != 200 {
		t.Fatalf("physical takeover worker code=%d err=%v output=%s", winner.code, winner.err, winner.output)
	}
	takeoverProxy.stop()
	if takeoverProxy.calls.Load() != 1 {
		t.Fatalf("competing physical takeover credit calls=%d want=1", takeoverProxy.calls.Load())
	}

	assertP1PhysicalDeliveryCompleted(t, ctx, db, contractID, unit.Id, lot.Id)
	assertP1PhysicalSuccessBalances(t, ctx, db, true, longUserID, shortUserID)
	assertP1PhysicalInstructionCounts(t, ctx, db, contractID, 4, 4, 4)
	assertP1PhysicalFlowIdentity(t, ctx, db, targetCredit.InstructionNo, 1)
	assertP1PhysicalTaskLockAbsent(t, ctx, serviceCtx.Redis)
}

func assertP1PhysicalDebitWorkerBoundary(
	t *testing.T, ctx context.Context, db *sql.DB, unitID, wantDebitSuccess int64,
) {
	t.Helper()
	var debitSuccess, debitFlows, creditSuccess, creditFlows int64
	if err := db.QueryRowContext(ctx, `SELECT
		COUNT(DISTINCT IF(instruction.step_no<3 AND instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.step_no<3,flow.id,NULL)),
		COUNT(DISTINCT IF(instruction.step_no=3 AND instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.step_no=3,flow.id,NULL))
		FROM t_option_asset_instruction instruction
		LEFT JOIN t_asset_flow flow
		  ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		WHERE instruction.delivery_unit_id=?`, unitID,
	).Scan(&debitSuccess, &debitFlows, &creditSuccess, &creditFlows); err != nil {
		t.Fatal(err)
	}
	if debitSuccess != wantDebitSuccess || debitFlows != wantDebitSuccess || creditSuccess != 0 || creditFlows != 0 {
		t.Fatalf("physical worker boundary debit-success/flows/credit-success/flows=%d/%d/%d/%d want=%d/%d/0/0",
			debitSuccess, debitFlows, creditSuccess, creditFlows, wantDebitSuccess, wantDebitSuccess)
	}
}

type p1PhysicalCreditCommitProxy struct {
	asset.UnimplementedAssetServer

	upstream    asset.AssetClient
	targetBizNo string
	committed   chan struct{}
	release     chan struct{}
	once        sync.Once
	calls       atomic.Int64
	listener    net.Listener
	server      *grpc.Server
}

func newP1PhysicalCreditCommitProxy(
	t *testing.T, upstream asset.AssetClient, targetBizNo string,
) *p1PhysicalCreditCommitProxy {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxy := &p1PhysicalCreditCommitProxy{
		upstream: upstream, targetBizNo: targetBizNo,
		committed: make(chan struct{}), release: make(chan struct{}),
		listener: listener, server: grpc.NewServer(),
	}
	asset.RegisterAssetServer(proxy.server, proxy)
	go func() { _ = proxy.server.Serve(listener) }()
	return proxy
}

func (p *p1PhysicalCreditCommitProxy) SubAvailable(
	ctx context.Context, in *asset.SubAvailableReq,
) (*asset.ChangeAssetResp, error) {
	return p.upstream.SubAvailable(ctx, in)
}

func (p *p1PhysicalCreditCommitProxy) DeductFrozenAssetByBizNo(
	ctx context.Context, in *asset.DeductFrozenAssetByBizNoReq,
) (*asset.ChangeAssetResp, error) {
	return p.upstream.DeductFrozenAssetByBizNo(ctx, in)
}

func (p *p1PhysicalCreditCommitProxy) AddAvailable(
	ctx context.Context, in *asset.AddAvailableReq,
) (*asset.ChangeAssetResp, error) {
	if in.BizNo != p.targetBizNo {
		return p.upstream.AddAvailable(ctx, in)
	}
	p.calls.Add(1)
	resp, err := p.upstream.AddAvailable(ctx, in)
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

func (p *p1PhysicalCreditCommitProxy) GetAssetFlowByBizNo(
	ctx context.Context, in *asset.GetAssetFlowByBizNoReq,
) (*asset.GetAssetFlowByBizNoResp, error) {
	return p.upstream.GetAssetFlowByBizNo(ctx, in)
}

func (p *p1PhysicalCreditCommitProxy) address() string {
	return p.listener.Addr().String()
}

func (p *p1PhysicalCreditCommitProxy) stop() {
	if p.server != nil {
		p.server.Stop()
		p.server = nil
	}
}

func assertP1PhysicalCommittedCreditBeforeKill(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	unitID, targetCreditID, remainingCreditID int64,
) {
	t.Helper()
	var unitStatus, debitSuccess, debitFlows int64
	var targetStatus, targetRetry, targetFlows, remainingStatus, remainingFlows int64
	if err := db.QueryRowContext(ctx, `SELECT unit.status,
		(SELECT COUNT(*) FROM t_option_asset_instruction debit
		 WHERE debit.delivery_unit_id=unit.id AND debit.step_no<3 AND debit.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction debit
		   ON debit.tenant_id=flow.tenant_id AND debit.instruction_no=flow.biz_no
		 WHERE debit.delivery_unit_id=unit.id AND debit.step_no<3),
		target.status,target.retry_count,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=target.tenant_id AND flow.biz_no=target.instruction_no),
		remaining.status,
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=remaining.tenant_id AND flow.biz_no=remaining.instruction_no)
		FROM t_option_physical_delivery_unit unit
		JOIN t_option_asset_instruction target ON target.id=? AND target.delivery_unit_id=unit.id
		JOIN t_option_asset_instruction remaining ON remaining.id=? AND remaining.delivery_unit_id=unit.id
		WHERE unit.id=?`, targetCreditID, remainingCreditID, unitID,
	).Scan(
		&unitStatus, &debitSuccess, &debitFlows,
		&targetStatus, &targetRetry, &targetFlows, &remainingStatus, &remainingFlows,
	); err != nil {
		t.Fatal(err)
	}
	if unitStatus != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_ASSET_PROCESSING) ||
		debitSuccess != 2 || debitFlows != 2 ||
		targetStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PROCESSING) ||
		targetRetry != 0 || targetFlows != 1 ||
		remainingStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING) || remainingFlows != 0 {
		t.Fatalf("physical before kill unit/debits/flows/target/retry/flow/remaining/flow=%d/%d/%d/%d/%d/%d/%d/%d",
			unitStatus, debitSuccess, debitFlows, targetStatus, targetRetry, targetFlows, remainingStatus, remainingFlows)
	}
}

func assertP1PhysicalTaskLockAbsent(
	t *testing.T,
	ctx context.Context,
	redis interface {
		ExistsCtx(context.Context, string) (bool, error)
	},
) {
	t.Helper()
	locked, err := redis.ExistsCtx(ctx, p0MultiInstanceLockKey)
	if err != nil {
		t.Fatal(err)
	}
	if locked {
		t.Fatalf("task lock %s remained after physical delivery process takeover", p0MultiInstanceLockKey)
	}
}

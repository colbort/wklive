package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type failOnceSubAvailableClient struct {
	asset.AssetClient

	mu          sync.Mutex
	targetBizNo string
	failures    int
}

type failOnceDeductFrozenClient struct {
	asset.AssetClient

	mu          sync.Mutex
	targetBizNo string
	failures    int
}

func (c *failOnceDeductFrozenClient) setTarget(bizNo string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.targetBizNo = bizNo
}

func (c *failOnceDeductFrozenClient) DeductFrozenAssetByBizNo(
	ctx context.Context,
	in *asset.DeductFrozenAssetByBizNoReq,
	opts ...grpc.CallOption,
) (*asset.ChangeAssetResp, error) {
	c.mu.Lock()
	shouldFail := in.BizNo == c.targetBizNo && c.failures == 0
	if shouldFail {
		c.failures++
	}
	c.mu.Unlock()
	if !shouldFail {
		return c.AssetClient.DeductFrozenAssetByBizNo(ctx, in, opts...)
	}
	if _, err := c.AssetClient.DeductFrozenAssetByBizNo(ctx, in, opts...); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unavailable, "P1 physical delivery injected response loss after committed collateral debit")
}

func (c *failOnceDeductFrozenClient) failureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

type p1PhysicalFailureMode int

const (
	p1PhysicalNoFailure p1PhysicalFailureMode = iota
	p1PhysicalBuyerDebitResponseLoss
	p1PhysicalSellerCollateralResponseLoss
)

func (c *failOnceSubAvailableClient) setTarget(bizNo string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.targetBizNo = bizNo
}

func (c *failOnceSubAvailableClient) SubAvailable(
	ctx context.Context,
	in *asset.SubAvailableReq,
	opts ...grpc.CallOption,
) (*asset.ChangeAssetResp, error) {
	c.mu.Lock()
	shouldFail := in.BizNo == c.targetBizNo && c.failures == 0
	if shouldFail {
		c.failures++
	}
	c.mu.Unlock()
	if !shouldFail {
		return c.AssetClient.SubAvailable(ctx, in, opts...)
	}
	if _, err := c.AssetClient.SubAvailable(ctx, in, opts...); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unavailable, "P1 physical delivery injected response loss after committed debit")
}

func (c *failOnceSubAvailableClient) failureCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

func testP1PhysicalDeliveryAssetRPC(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	t.Run("covered Call response loss recovery", func(t *testing.T) {
		testP1PhysicalDeliverySuccess(t, ctx, db, assetClient, serviceCtx, true, p1PhysicalBuyerDebitResponseLoss)
	})
	t.Run("covered Put success", func(t *testing.T) {
		testP1PhysicalDeliverySuccess(t, ctx, db, assetClient, serviceCtx, false, p1PhysicalNoFailure)
	})
	t.Run("covered Call seller collateral response loss recovery", func(t *testing.T) {
		testP1PhysicalDeliverySuccess(t, ctx, db, assetClient, serviceCtx, true, p1PhysicalSellerCollateralResponseLoss)
	})
	t.Run("insufficient buyer isolated default and original identity recovery", func(t *testing.T) {
		testP1PhysicalDeliveryInsufficientRecovery(t, ctx, db, assetClient, serviceCtx)
	})
}

func testP1PhysicalDeliverySuccess(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	isCall bool,
	failureMode p1PhysicalFailureMode,
) {
	t.Helper()
	now := time.Now().Unix()
	contractID := int64(997001)
	optionType := option.OptionType_OPTION_TYPE_CALL
	deliveryPrice := "120"
	longUserID, shortUserID := int64(3101), int64(3102)
	longDebitCoin, longDebitSeed := "USDT", "150"
	shortCollateralCoin, shortCollateral := "BTC", "1"
	if isCall && failureMode == p1PhysicalSellerCollateralResponseLoss {
		contractID = 997004
		longUserID, shortUserID = 3131, 3132
	}
	if !isCall {
		contractID = 997002
		optionType = option.OptionType_OPTION_TYPE_PUT
		deliveryPrice = "80"
		longUserID, shortUserID = 3111, 3112
		longDebitCoin, longDebitSeed = "BTC", "2"
		shortCollateralCoin, shortCollateral = "USDT", "100"
	}
	prefix := fmt.Sprintf("P1-PHYSICAL-%s", optionType.String())
	if failureMode == p1PhysicalSellerCollateralResponseLoss {
		prefix = "P1-PHYSICAL-CALL-COLLATERAL-RESPONSE-LOSS"
	}
	seedP1PhysicalContract(t, ctx, db, contractID, prefix, optionType, now-10, now-1)
	creditAssetCoin(t, ctx, assetClient, longUserID, longDebitCoin, longDebitSeed, prefix+"-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, shortUserID, shortCollateralCoin, shortCollateral, prefix+"-SHORT-SEED")

	longPosition := insertP1PhysicalPosition(
		t, ctx, serviceCtx, contractID, longUserID, longUserID,
		common.PositionSide_POSITION_SIDE_LONG, now-200, decimal.Zero,
	)
	shortPosition := insertP1PhysicalPosition(
		t, ctx, serviceCtx, contractID, shortUserID, shortUserID,
		common.PositionSide_POSITION_SIDE_SHORT, now-190, decimal.RequireFromString(shortCollateral),
	)
	lot := insertP1PhysicalMarginLot(
		t, ctx, serviceCtx, shortPosition, prefix+"-SHORT-COLLATERAL",
		shortCollateralCoin, shortCollateral, now-180,
	)
	freezeP1PhysicalCollateral(t, ctx, assetClient, lot, shortCollateralCoin, shortCollateral)
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, now-10, now, prefix,
		physicalEvidencePrices(deliveryPrice), deliveryPrice,
	)
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).processExpiredContracts(now); err != nil {
		t.Fatalf("create %s physical delivery: %v", prefix, err)
	}
	unit := findP1PhysicalUnitByLongUser(t, ctx, db, contractID, longUserID)
	longDebit := findP1PhysicalInstruction(t, ctx, serviceCtx, unit.Id,
		option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE)
	shortDebit := findP1PhysicalInstruction(t, ctx, serviceCtx, unit.Id,
		option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEDUCT_FROZEN)

	originalClient := serviceCtx.AssetClient
	if failureMode == p1PhysicalBuyerDebitResponseLoss {
		faultClient := &failOnceSubAvailableClient{AssetClient: assetClient}
		faultClient.setTarget(longDebit.InstructionNo)
		serviceCtx.AssetClient = faultClient
		defer func() { serviceCtx.AssetClient = originalClient }()

		processAssetInstructions(t, ctx, serviceCtx)
		if faultClient.failureCount() != 1 {
			t.Fatalf("physical committed-debit response losses=%d want=1", faultClient.failureCount())
		}
		assertP1PhysicalDebitRecoveryBarrier(t, ctx, db, unit.Id, longDebit.InstructionNo)
		if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
			SET next_retry_at=0 WHERE id=? AND status=?`, longDebit.Id,
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED)); err != nil {
			t.Fatal(err)
		}
	} else if failureMode == p1PhysicalSellerCollateralResponseLoss {
		faultClient := &failOnceDeductFrozenClient{AssetClient: assetClient}
		faultClient.setTarget(shortDebit.InstructionNo)
		serviceCtx.AssetClient = faultClient
		defer func() { serviceCtx.AssetClient = originalClient }()

		for attempt := 0; attempt < 2 && faultClient.failureCount() == 0; attempt++ {
			processAssetInstructions(t, ctx, serviceCtx)
		}
		if faultClient.failureCount() != 1 {
			t.Fatalf("physical committed-collateral response losses=%d want=1", faultClient.failureCount())
		}
		assertP1PhysicalCollateralRecoveryBarrier(
			t, ctx, db, unit.Id, longDebit.InstructionNo, shortDebit.InstructionNo,
		)
		if _, err := db.ExecContext(ctx, `UPDATE t_option_asset_instruction
			SET next_retry_at=0 WHERE id=? AND status=?`, shortDebit.Id,
			int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED)); err != nil {
			t.Fatal(err)
		}
	}
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1PhysicalDeliveryCompleted(t, ctx, db, contractID, unit.Id, lot.Id)
	assertP1PhysicalSuccessBalances(
		t, ctx, db, isCall, longUserID, shortUserID,
	)
	if failureMode == p1PhysicalBuyerDebitResponseLoss {
		assertP1PhysicalFlowIdentity(t, ctx, db, longDebit.InstructionNo, 1)
	} else if failureMode == p1PhysicalSellerCollateralResponseLoss {
		assertP1PhysicalFlowIdentity(t, ctx, db, shortDebit.InstructionNo, 1)
	}

	for attempt := 0; attempt < 2; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1PhysicalDeliveryCompleted(t, ctx, db, contractID, unit.Id, lot.Id)
	assertP1PhysicalInstructionCounts(t, ctx, db, contractID, 4, 4, 4)
	_ = longPosition
}

func testP1PhysicalDeliveryInsufficientRecovery(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		contractID         int64 = 997003
		failedLongUserID   int64 = 3121
		healthyLongUserID  int64 = 3122
		failedShortUserID  int64 = 3123
		healthyShortUserID int64 = 3124
	)
	now := time.Now().Unix()
	prefix := "P1-PHYSICAL-CALL-INSUFFICIENT"
	seedP1PhysicalContract(
		t, ctx, db, contractID, prefix, option.OptionType_OPTION_TYPE_CALL, now-10, now-1,
	)
	creditAssetCoin(t, ctx, assetClient, healthyLongUserID, "USDT", "100", prefix+"-HEALTHY-LONG-SEED")
	creditAssetCoin(t, ctx, assetClient, failedShortUserID, "BTC", "1", prefix+"-FAILED-SHORT-SEED")
	creditAssetCoin(t, ctx, assetClient, healthyShortUserID, "BTC", "1", prefix+"-HEALTHY-SHORT-SEED")

	insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, failedLongUserID, failedLongUserID,
		common.PositionSide_POSITION_SIDE_LONG, now-240, decimal.Zero)
	insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, healthyLongUserID, healthyLongUserID,
		common.PositionSide_POSITION_SIDE_LONG, now-230, decimal.Zero)
	failedShort := insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, failedShortUserID, failedShortUserID,
		common.PositionSide_POSITION_SIDE_SHORT, now-220, decimal.NewFromInt(1))
	healthyShort := insertP1PhysicalPosition(t, ctx, serviceCtx, contractID, healthyShortUserID, healthyShortUserID,
		common.PositionSide_POSITION_SIDE_SHORT, now-210, decimal.NewFromInt(1))
	failedLot := insertP1PhysicalMarginLot(t, ctx, serviceCtx, failedShort,
		prefix+"-FAILED-SHORT-COLLATERAL", "BTC", "1", now-200)
	healthyLot := insertP1PhysicalMarginLot(t, ctx, serviceCtx, healthyShort,
		prefix+"-HEALTHY-SHORT-COLLATERAL", "BTC", "1", now-190)
	freezeP1PhysicalCollateral(t, ctx, assetClient, failedLot, "BTC", "1")
	freezeP1PhysicalCollateral(t, ctx, assetClient, healthyLot, "BTC", "1")
	seedP0SettlementPriceEvidenceWithSamples(
		t, ctx, db, contractID, now-10, now, prefix,
		physicalEvidencePrices("120"), "120",
	)
	if err := NewProcessContractLifecycleLogic(ctx, serviceCtx).processExpiredContracts(now); err != nil {
		t.Fatalf("create insufficient physical delivery: %v", err)
	}
	failedUnit := findP1PhysicalUnitByLongUser(t, ctx, db, contractID, failedLongUserID)
	healthyUnit := findP1PhysicalUnitByLongUser(t, ctx, db, contractID, healthyLongUserID)
	failedDebit := findP1PhysicalInstruction(t, ctx, serviceCtx, failedUnit.Id,
		option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_DEBIT_AVAILABLE)

	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1PhysicalUnitStatus(t, ctx, db, failedUnit.Id,
		option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_CURE_REQUIRED)
	assertP1PhysicalUnitStatus(t, ctx, db, healthyUnit.Id,
		option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_COMPLETED)
	assertP1PhysicalFailureBarrier(t, ctx, db, failedUnit.Id, failedDebit.InstructionNo)
	assertP1PhysicalHealthyUnitBalances(t, ctx, db, healthyLongUserID, healthyShortUserID)

	if err := NewProcessAssetInstructionsLogic(ctx, serviceCtx).
		expirePhysicalDeliveryUnit(failedUnit.Id, failedUnit.CureDeadline); err != nil {
		t.Fatalf("expire physical cure: %v", err)
	}
	assertP1PhysicalUnitStatus(t, ctx, db, failedUnit.Id,
		option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_DEFAULTED)
	assertP1PhysicalBatchFailed(t, ctx, db, contractID)
	creditAssetCoin(t, ctx, assetClient, failedLongUserID, "USDT", "100", prefix+"-FAILED-LONG-TOPUP")

	retryP1PhysicalDeliveryConcurrently(t, ctx, serviceCtx, failedUnit.Id)
	for attempt := 0; attempt < 4; attempt++ {
		processAssetInstructions(t, ctx, serviceCtx)
	}
	assertP1PhysicalDeliveryCompleted(t, ctx, db, contractID, failedUnit.Id, failedLot.Id)
	assertP1PhysicalDeliveryCompleted(t, ctx, db, contractID, healthyUnit.Id, healthyLot.Id)
	assertP1PhysicalRecoveredBalances(
		t, ctx, db, failedLongUserID, healthyLongUserID, failedShortUserID, healthyShortUserID,
	)
	assertP1PhysicalManualRetryEvidence(t, ctx, db, contractID, failedUnit.Id, failedDebit.InstructionNo)
	assertP1PhysicalInstructionCounts(t, ctx, db, contractID, 8, 8, 8)
}

func retryP1PhysicalDeliveryConcurrently(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	unitID int64,
) {
	t.Helper()
	const workers = 20
	type retryResult struct {
		code int32
		err  error
	}
	start := make(chan struct{})
	results := make(chan retryResult, workers)
	var group sync.WaitGroup
	group.Add(workers)
	for worker := 0; worker < workers; worker++ {
		go func() {
			defer group.Done()
			<-start
			adminCtx := metadata.NewIncomingContext(ctx, metadata.Pairs(
				utils.CtxKeyUserType, fmt.Sprint(utils.SysUserTypeSystemAdmin),
				utils.CtxKeyUid, "9005",
			))
			response, err := adminlogic.NewRetryPhysicalDeliveryUnitLogic(
				adminCtx, serviceCtx,
			).RetryPhysicalDeliveryUnit(&option.RetryPhysicalDeliveryUnitReq{
				TenantId: p0AssetE2ETenantID, DeliveryUnitId: unitID,
				Reason: "BUYER_ASSET_TOPUP_VERIFIED",
			})
			code := int32(0)
			if response != nil && response.GetBase() != nil {
				code = response.GetBase().GetCode()
			}
			results <- retryResult{code: code, err: err}
		}()
	}
	close(start)
	group.Wait()
	close(results)

	success, rejected := 0, 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("concurrent physical manual retry: %v", result.err)
		}
		if result.code == 200 {
			success++
		} else {
			rejected++
		}
	}
	if success != 1 || rejected != workers-1 {
		t.Fatalf("concurrent physical manual retry success/rejected=%d/%d want=1/%d",
			success, rejected, workers-1)
	}
}

func seedP1PhysicalContract(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID int64,
	code string,
	optionType option.OptionType,
	expireTime, deliverTime int64,
) {
	t.Helper()
	_, err := db.ExecContext(ctx, `INSERT INTO t_option_contract (
		id,tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
		option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
		max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
		deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
		circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
		maintenance_margin_rate,min_margin_rate,physical_delivery_policy,
		physical_delivery_cure_seconds,status,is_deleted,create_times,update_times
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		contractID, p0AssetE2ETenantID, code, "BTCUSDT", "BTC", "USDT", "USDT",
		int64(optionType), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		int64(option.SettlementType_SETTLEMENT_TYPE_PHYSICAL), "100", "1", "1", "1000", "0.1", "1", "1",
		expireTime-3600, expireTime, expireTime, deliverTime, "10000", "10000", "10000", "0.2", "0.5", 60,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_COVERED_DELIVERY), "1", "0.1", "1",
		int64(option.PhysicalDeliveryPolicy_PHYSICAL_DELIVERY_POLICY_STRICT), 3600,
		int64(option.ContractStatus_CONTRACT_STATUS_EXPIRED), int64(common.YesNo_YES_NO_NO), expireTime-3600, expireTime,
	)
	if err != nil {
		t.Fatalf("seed physical delivery contract %s: %v", code, err)
	}
}

func physicalEvidencePrices(deliveryPrice string) []string {
	price := decimal.RequireFromString(deliveryPrice)
	return []string{price.Sub(decimal.NewFromInt(1)).String(), price.String(), price.Add(decimal.NewFromInt(1)).String()}
}

func insertP1PhysicalPosition(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractID, userID, accountID int64,
	side common.PositionSide,
	createTimes int64,
	margin decimal.Decimal,
) *models.TOptionPosition {
	t.Helper()
	return insertP0SettlementPosition(t, ctx, serviceCtx, &models.TOptionPosition{
		TenantId: p0AssetE2ETenantID, UserId: userID, AccountId: accountID,
		ContractId: contractID, UnderlyingSymbol: "BTCUSDT", Side: int64(side),
		PositionQty: decimal.NewFromInt(1), AvailableQty: decimal.NewFromInt(1),
		OpenAvgPrice: decimal.NewFromInt(10), MarkPrice: decimal.NewFromInt(10),
		PositionValue: decimal.NewFromInt(10), MarginAmount: margin,
		ExerciseableQty: decimal.NewFromInt(1), Status: int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: createTimes, UpdateTimes: createTimes,
	})
}

func insertP1PhysicalMarginLot(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	position *models.TOptionPosition,
	freezeBizNo, coin, amount string,
	createTimes int64,
) *models.TOptionMarginLot {
	t.Helper()
	margin := decimal.RequireFromString(amount)
	lot := &models.TOptionMarginLot{
		TenantId: position.TenantId, UserId: position.UserId, AccountId: position.AccountId,
		ContractId: position.ContractId, PositionId: position.Id,
		OriginContractId: position.ContractId, OriginPositionId: position.Id,
		TradeId: -position.Id, FreezeBizNo: freezeBizNo, CollateralCoin: coin,
		Quantity: decimal.NewFromInt(1), RemainingQuantity: decimal.NewFromInt(1),
		InitialMargin: margin, RemainingMargin: margin,
		Status:      int64(option.MarginLotStatus_MARGIN_LOT_STATUS_ACTIVE),
		CreateTimes: createTimes, UpdateTimes: createTimes,
	}
	result, err := serviceCtx.OptionMarginLotModel.Insert(ctx, lot)
	if err != nil {
		t.Fatalf("insert physical margin lot: %v", err)
	}
	lot.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return lot
}

func freezeP1PhysicalCollateral(
	t *testing.T,
	ctx context.Context,
	assetClient asset.AssetClient,
	lot *models.TOptionMarginLot,
	coin, amount string,
) {
	t.Helper()
	response, err := assetClient.FreezeAsset(ctx, &asset.FreezeAssetReq{
		TenantId: p0AssetE2ETenantID, UserId: lot.UserId,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: coin, Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_PLACE_ORDER,
		BizId: lot.Id, BizNo: lot.FreezeBizNo, Remark: "P1 physical delivery covered collateral",
	})
	assertAssetOK(t, response, err)
}

func findP1PhysicalUnitByLongUser(
	t *testing.T, ctx context.Context, db *sql.DB, contractID, longUserID int64,
) *models.TOptionPhysicalDeliveryUnit {
	t.Helper()
	var unit models.TOptionPhysicalDeliveryUnit
	err := db.QueryRowContext(ctx, `SELECT id,tenant_id,delivery_unit_no,batch_id,batch_no,contract_id,
		long_position_id,long_user_id,long_account_id,short_position_id,short_user_id,short_account_id,
		quantity,delivery_coin,delivery_quantity,payment_coin,payment_amount,status,cure_deadline,
		failed_instruction_id,last_error_msg,completed_at,manual_retry_count,create_times,update_times
		FROM t_option_physical_delivery_unit WHERE tenant_id=? AND contract_id=? AND long_user_id=?`,
		p0AssetE2ETenantID, contractID, longUserID,
	).Scan(
		&unit.Id, &unit.TenantId, &unit.DeliveryUnitNo, &unit.BatchId, &unit.BatchNo, &unit.ContractId,
		&unit.LongPositionId, &unit.LongUserId, &unit.LongAccountId, &unit.ShortPositionId,
		&unit.ShortUserId, &unit.ShortAccountId, &unit.Quantity, &unit.DeliveryCoin,
		&unit.DeliveryQuantity, &unit.PaymentCoin, &unit.PaymentAmount, &unit.Status,
		&unit.CureDeadline, &unit.FailedInstructionId, &unit.LastErrorMsg, &unit.CompletedAt,
		&unit.ManualRetryCount, &unit.CreateTimes, &unit.UpdateTimes,
	)
	if err != nil {
		t.Fatalf("find physical delivery unit: %v", err)
	}
	return &unit
}

func findP1PhysicalInstruction(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	unitID int64,
	action option.AssetInstructionAction,
) *models.TOptionAssetInstruction {
	t.Helper()
	instructions, err := serviceCtx.OptionAssetInstructionModel.FindByDeliveryUnit(
		ctx, p0AssetE2ETenantID, unitID,
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, instruction := range instructions {
		if instruction.Action == int64(action) {
			return instruction
		}
	}
	t.Fatalf("physical unit %d has no action %s", unitID, action)
	return nil
}

func assertP1PhysicalDebitRecoveryBarrier(
	t *testing.T, ctx context.Context, db *sql.DB, unitID int64, instructionNo string,
) {
	t.Helper()
	var unitStatus, instructionStatus, retryCount, laterSuccess, flowCount int64
	if err := db.QueryRowContext(ctx, `SELECT unit.status,instruction.status,instruction.retry_count,
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.delivery_unit_id=unit.id AND later.step_no>1 AND later.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow WHERE flow.tenant_id=unit.tenant_id AND flow.biz_no=instruction.instruction_no)
		FROM t_option_physical_delivery_unit unit
		JOIN t_option_asset_instruction instruction ON instruction.delivery_unit_id=unit.id
		WHERE unit.id=? AND instruction.instruction_no=?`, unitID, instructionNo,
	).Scan(&unitStatus, &instructionStatus, &retryCount, &laterSuccess, &flowCount); err != nil {
		t.Fatal(err)
	}
	if unitStatus != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_CURE_REQUIRED) ||
		instructionStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) ||
		retryCount != 1 || laterSuccess != 0 || flowCount != 1 {
		t.Fatalf("physical response-loss barrier unit/status/retry/later/flow=%d/%d/%d/%d/%d",
			unitStatus, instructionStatus, retryCount, laterSuccess, flowCount)
	}
}

func assertP1PhysicalCollateralRecoveryBarrier(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	unitID int64,
	longDebitInstructionNo, shortDebitInstructionNo string,
) {
	t.Helper()
	var unitStatus, shortStatus, shortRetry, laterSuccess, shortFlows, longSuccess, longFlows int64
	if err := db.QueryRowContext(ctx, `SELECT unit.status,short_debit.status,short_debit.retry_count,
		(SELECT COUNT(*) FROM t_option_asset_instruction later
		 WHERE later.delivery_unit_id=unit.id AND later.step_no>2 AND later.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=unit.tenant_id AND flow.biz_no=short_debit.instruction_no),
		(SELECT COUNT(*) FROM t_option_asset_instruction long_debit
		 WHERE long_debit.delivery_unit_id=unit.id AND long_debit.instruction_no=? AND long_debit.status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 WHERE flow.tenant_id=unit.tenant_id AND flow.biz_no=?)
		FROM t_option_physical_delivery_unit unit
		JOIN t_option_asset_instruction short_debit ON short_debit.delivery_unit_id=unit.id
		WHERE unit.id=? AND short_debit.instruction_no=?`,
		longDebitInstructionNo, longDebitInstructionNo, unitID, shortDebitInstructionNo,
	).Scan(&unitStatus, &shortStatus, &shortRetry, &laterSuccess, &shortFlows, &longSuccess, &longFlows); err != nil {
		t.Fatal(err)
	}
	if unitStatus != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_CURE_REQUIRED) ||
		shortStatus != int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_FAILED) ||
		shortRetry != 1 || laterSuccess != 0 || shortFlows != 1 || longSuccess != 1 || longFlows != 1 {
		t.Fatalf("physical collateral response-loss barrier unit/status/retry/later/short-flow/long-success/long-flow=%d/%d/%d/%d/%d/%d/%d",
			unitStatus, shortStatus, shortRetry, laterSuccess, shortFlows, longSuccess, longFlows)
	}
}

func assertP1PhysicalFailureBarrier(
	t *testing.T, ctx context.Context, db *sql.DB, unitID int64, instructionNo string,
) {
	t.Helper()
	var failed, laterSuccess, flows int64
	if err := db.QueryRowContext(ctx, `SELECT
		SUM(instruction_no=? AND status=4),SUM(step_no>1 AND status=3),
		(SELECT COUNT(*) FROM t_asset_flow flow
		 JOIN t_option_asset_instruction linked ON linked.tenant_id=flow.tenant_id AND linked.instruction_no=flow.biz_no
		 WHERE linked.delivery_unit_id=?)
		FROM t_option_asset_instruction WHERE delivery_unit_id=?`,
		instructionNo, unitID, unitID,
	).Scan(&failed, &laterSuccess, &flows); err != nil {
		t.Fatal(err)
	}
	if failed != 1 || laterSuccess != 0 || flows != 0 {
		t.Fatalf("physical insufficient barrier failed/later/flows=%d/%d/%d", failed, laterSuccess, flows)
	}
}

func assertP1PhysicalUnitStatus(
	t *testing.T, ctx context.Context, db *sql.DB, unitID int64, want option.PhysicalDeliveryUnitStatus,
) {
	t.Helper()
	var got int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_physical_delivery_unit WHERE id=?`, unitID).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != int64(want) {
		t.Fatalf("physical unit %d status=%d want=%s", unitID, got, want)
	}
}

func assertP1PhysicalDeliveryCompleted(
	t *testing.T, ctx context.Context, db *sql.DB, contractID, unitID, lotID int64,
) {
	t.Helper()
	var settlementStatus, batchStatus, contractStatus, unitStatus int64
	var instructions, success, reconciled, flows, lotStatus int64
	var remaining, pending string
	if err := db.QueryRowContext(ctx, `SELECT settlement.status,batch.status,contract.status,unit.status,
		COUNT(DISTINCT instruction.id),COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),COUNT(DISTINCT flow.id),
		lot.status,CAST(lot.remaining_margin AS CHAR),CAST(lot.pending_margin AS CHAR)
		FROM t_option_physical_delivery_unit unit
		JOIN t_option_settlement settlement ON settlement.tenant_id=unit.tenant_id AND settlement.settlement_no=unit.batch_no
		JOIN t_option_settlement_batch batch ON batch.tenant_id=unit.tenant_id AND batch.id=unit.batch_id
		JOIN t_option_contract contract ON contract.tenant_id=unit.tenant_id AND contract.id=unit.contract_id
		JOIN t_option_asset_instruction instruction ON instruction.tenant_id=unit.tenant_id AND instruction.delivery_unit_id=unit.id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		JOIN t_option_margin_lot lot ON lot.tenant_id=unit.tenant_id AND lot.id=?
		WHERE unit.id=? AND unit.contract_id=?
		GROUP BY settlement.id,batch.id,contract.id,unit.id,lot.id`, lotID, unitID, contractID,
	).Scan(&settlementStatus, &batchStatus, &contractStatus, &unitStatus,
		&instructions, &success, &reconciled, &flows, &lotStatus, &remaining, &pending); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_DONE) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_DONE) ||
		contractStatus != int64(option.ContractStatus_CONTRACT_STATUS_SETTLED) ||
		unitStatus != int64(option.PhysicalDeliveryUnitStatus_PHYSICAL_DELIVERY_UNIT_STATUS_COMPLETED) ||
		instructions != 4 || success != 4 || reconciled != 4 || flows != 4 ||
		lotStatus != int64(option.MarginLotStatus_MARGIN_LOT_STATUS_RESOLVED) ||
		remaining != "0.0000000000000000" || pending != "0.0000000000000000" {
		t.Fatalf("physical completion settlement/batch/contract/unit=%d/%d/%d/%d instructions=%d/%d/%d/%d lot=%d/%s/%s",
			settlementStatus, batchStatus, contractStatus, unitStatus, instructions, success, reconciled, flows,
			lotStatus, remaining, pending)
	}
}

func assertP1PhysicalInstructionCounts(
	t *testing.T, ctx context.Context, db *sql.DB, contractID, wantTotal, wantSuccess, wantFlows int64,
) {
	t.Helper()
	var total, success, reconciled, flows int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(DISTINCT instruction.id),
		COUNT(DISTINCT IF(instruction.status=3,instruction.id,NULL)),
		COUNT(DISTINCT IF(instruction.reconciliation_status=2,instruction.id,NULL)),
		COUNT(DISTINCT flow.id)
		FROM t_option_asset_instruction instruction
		JOIN t_option_physical_delivery_unit unit ON unit.tenant_id=instruction.tenant_id AND unit.id=instruction.delivery_unit_id
		LEFT JOIN t_asset_flow flow ON flow.tenant_id=instruction.tenant_id AND flow.biz_no=instruction.instruction_no
		WHERE unit.tenant_id=? AND unit.contract_id=?`, p0AssetE2ETenantID, contractID,
	).Scan(&total, &success, &reconciled, &flows); err != nil {
		t.Fatal(err)
	}
	if total != wantTotal || success != wantSuccess || reconciled != wantSuccess || flows != wantFlows {
		t.Fatalf("physical instructions total/success/reconciled/flows=%d/%d/%d/%d want=%d/%d/%d",
			total, success, reconciled, flows, wantTotal, wantSuccess, wantFlows)
	}
}

func assertP1PhysicalSuccessBalances(
	t *testing.T, ctx context.Context, db *sql.DB, isCall bool, longUserID, shortUserID int64,
) {
	t.Helper()
	if isCall {
		assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "50.000000000000000000", "50.000000000000000000", "0.000000000000000000")
		assertWalletCoinAmounts(t, ctx, db, longUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
		assertWalletCoinAmounts(t, ctx, db, shortUserID, "BTC", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
		assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
		return
	}
	assertWalletCoinAmounts(t, ctx, db, longUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
}

func assertP1PhysicalHealthyUnitBalances(
	t *testing.T, ctx context.Context, db *sql.DB, longUserID, shortUserID int64,
) {
	t.Helper()
	assertWalletCoinAmounts(t, ctx, db, longUserID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, longUserID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "BTC", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
	assertWalletCoinAmounts(t, ctx, db, shortUserID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
}

func assertP1PhysicalRecoveredBalances(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	failedLongUserID, healthyLongUserID, failedShortUserID, healthyShortUserID int64,
) {
	t.Helper()
	for _, userID := range []int64{failedLongUserID, healthyLongUserID} {
		assertWalletCoinAmounts(t, ctx, db, userID, "USDT", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
		assertWalletCoinAmounts(t, ctx, db, userID, "BTC", "1.000000000000000000", "1.000000000000000000", "0.000000000000000000")
	}
	for _, userID := range []int64{failedShortUserID, healthyShortUserID} {
		assertWalletCoinAmounts(t, ctx, db, userID, "BTC", "0.000000000000000000", "0.000000000000000000", "0.000000000000000000")
		assertWalletCoinAmounts(t, ctx, db, userID, "USDT", "100.000000000000000000", "100.000000000000000000", "0.000000000000000000")
	}
}

func assertP1PhysicalBatchFailed(t *testing.T, ctx context.Context, db *sql.DB, contractID int64) {
	t.Helper()
	var settlementStatus, batchStatus int64
	if err := db.QueryRowContext(ctx, `SELECT settlement.status,batch.status
		FROM t_option_settlement settlement
		JOIN t_option_settlement_batch batch ON batch.tenant_id=settlement.tenant_id AND batch.batch_no=settlement.settlement_no
		WHERE settlement.tenant_id=? AND settlement.contract_id=?`, p0AssetE2ETenantID, contractID,
	).Scan(&settlementStatus, &batchStatus); err != nil {
		t.Fatal(err)
	}
	if settlementStatus != int64(option.SettlementStatus_SETTLEMENT_STATUS_FAILED) ||
		batchStatus != int64(option.SettlementBatchStatus_SETTLEMENT_BATCH_STATUS_FAILED) {
		t.Fatalf("physical default settlement/batch=%d/%d want failed", settlementStatus, batchStatus)
	}
}

func assertP1PhysicalManualRetryEvidence(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	contractID, unitID int64,
	instructionNo string,
) {
	t.Helper()
	var manualRetryCount, instructionRetry, instructionFlows, eventCount, operatorID int64
	var reason string
	if err := db.QueryRowContext(ctx, `SELECT unit.manual_retry_count,instruction.retry_count,
		(SELECT COUNT(*) FROM t_asset_flow flow WHERE flow.tenant_id=unit.tenant_id AND flow.biz_no=instruction.instruction_no),
		(SELECT COUNT(*) FROM t_option_trading_control_event event
		 WHERE event.tenant_id=unit.tenant_id AND event.contract_id=unit.contract_id
		   AND event.event_type='PHYSICAL_DELIVERY_MANUAL_RETRY'),
		COALESCE((SELECT MAX(event.operator_id) FROM t_option_trading_control_event event
		 WHERE event.tenant_id=unit.tenant_id AND event.contract_id=unit.contract_id
		   AND event.event_type='PHYSICAL_DELIVERY_MANUAL_RETRY'),0),
		COALESCE((SELECT MAX(event.reason) FROM t_option_trading_control_event event
		 WHERE event.tenant_id=unit.tenant_id AND event.contract_id=unit.contract_id
		   AND event.event_type='PHYSICAL_DELIVERY_MANUAL_RETRY'),'')
		FROM t_option_physical_delivery_unit unit
		JOIN t_option_asset_instruction instruction ON instruction.delivery_unit_id=unit.id
		WHERE unit.id=? AND unit.contract_id=? AND instruction.instruction_no=?`,
		unitID, contractID, instructionNo,
	).Scan(&manualRetryCount, &instructionRetry, &instructionFlows, &eventCount, &operatorID, &reason); err != nil {
		t.Fatal(err)
	}
	if manualRetryCount != 1 || instructionRetry != 0 || instructionFlows != 1 ||
		eventCount != 1 || operatorID != 9005 || reason != "BUYER_ASSET_TOPUP_VERIFIED" {
		t.Fatalf("physical retry evidence manual/instruction/flows/events/operator/reason=%d/%d/%d/%d/%d/%q",
			manualRetryCount, instructionRetry, instructionFlows, eventCount, operatorID, reason)
	}
}

func assertP1PhysicalFlowIdentity(
	t *testing.T, ctx context.Context, db *sql.DB, instructionNo string, want int64,
) {
	t.Helper()
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_asset_flow
		WHERE tenant_id=? AND biz_no=?`, p0AssetE2ETenantID, instructionNo).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("physical instruction %s flows=%d want=%d", instructionNo, count, want)
	}
}

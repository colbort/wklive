package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/delayqueue"
	adminlogic "wklive/services/option/internal/logic/admin"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/proto"
)

const (
	p2SeriesCreatorID   int64 = 96801
	p2SeriesReviewerA   int64 = 96802
	p2SeriesReviewerB   int64 = 96803
	p2SeriesLauncherA   int64 = 96804
	p2SeriesLauncherB   int64 = 96805
	p2SeriesContractNum int64 = 500
)

func testP2ContractSeriesCapacityAndLaunch(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	calendarCode := "P2_SERIES_24_7"
	seedP0OpenTradingCalendar(t, ctx, db, calendarCode, now)
	request := p2ContractSeriesRequest(now, calendarCode, "P2-SERIES-500-REQUEST", "P2SER500", 125)

	const createConcurrency = 20
	createResponses := make([]*option.GetContractSeriesResp, createConcurrency)
	createErrors := make([]error, createConcurrency)
	start := make(chan struct{})
	var createWG sync.WaitGroup
	createWG.Add(createConcurrency)
	for index := 0; index < createConcurrency; index++ {
		go func(index int) {
			defer createWG.Done()
			<-start
			createResponses[index], createErrors[index] = adminlogic.NewCreateContractSeriesLogic(
				p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
			).CreateContractSeries(request)
		}(index)
	}
	close(start)
	createWG.Wait()
	var seriesID int64
	var payloadHash string
	for index := range createResponses {
		response := createResponses[index]
		if createErrors[index] != nil || response == nil || response.Base == nil ||
			response.Base.Code != 200 || response.Data == nil {
			t.Fatalf("contract series concurrent create[%d] response=%+v err=%v",
				index, response, createErrors[index])
		}
		if index == 0 {
			seriesID, payloadHash = response.Data.Id, response.Data.PayloadHash
		}
		if response.Data.Id != seriesID || response.Data.PayloadHash != payloadHash {
			t.Fatalf("contract series create identity[%d]=%d/%s want=%d/%s",
				index, response.Data.Id, response.Data.PayloadHash, seriesID, payloadHash)
		}
	}
	var seriesRows, expiryRows, bandRows int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_contract_series WHERE tenant_id=? AND request_key=?),
		(SELECT COUNT(*) FROM t_option_contract_series_expiry WHERE tenant_id=? AND series_id=?),
		(SELECT COUNT(*) FROM t_option_contract_series_strike_band WHERE tenant_id=? AND series_id=?)`,
		p0AssetE2ETenantID, request.RequestKey,
		p0AssetE2ETenantID, seriesID,
		p0AssetE2ETenantID, seriesID,
	).Scan(&seriesRows, &expiryRows, &bandRows); err != nil {
		t.Fatal(err)
	}
	if seriesRows != 1 || expiryRows != 2 || bandRows != 1 {
		t.Fatalf("contract series create rows=%d/%d/%d want=1/2/1", seriesRows, expiryRows, bandRows)
	}
	conflict := proto.Clone(request).(*option.CreateContractSeriesReq)
	conflict.ReferencePrice = "101"
	conflictResp, conflictErr := adminlogic.NewCreateContractSeriesLogic(
		p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateContractSeries(conflict)
	if conflictErr == nil && conflictResp != nil && conflictResp.Base != nil && conflictResp.Base.Code == 200 {
		t.Fatal("same contract-series request key accepted a different economic payload")
	}

	selfReview, selfReviewErr := adminlogic.NewReviewContractSeriesLogic(
		p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
	).ReviewContractSeries(&option.ReviewContractSeriesReq{
		TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true, Reason: "self review forbidden",
	})
	if selfReviewErr == nil && selfReview != nil && selfReview.Base != nil && selfReview.Base.Code == 200 {
		t.Fatal("contract series creator approved its own generation")
	}

	reviewers := []int64{p2SeriesReviewerA, p2SeriesReviewerB}
	reviewResponses := make([]*option.GetContractSeriesResp, len(reviewers))
	reviewErrors := make([]error, len(reviewers))
	start = make(chan struct{})
	var reviewWG sync.WaitGroup
	reviewWG.Add(len(reviewers))
	reviewStartedAt := time.Now()
	for index, reviewerID := range reviewers {
		go func(index int, reviewerID int64) {
			defer reviewWG.Done()
			<-start
			reviewResponses[index], reviewErrors[index] = adminlogic.NewReviewContractSeriesLogic(
				p0AdminContext(ctx, reviewerID, p0AssetE2ETenantID), serviceCtx,
			).ReviewContractSeries(&option.ReviewContractSeriesReq{
				TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true,
				Reason: fmt.Sprintf("independent capacity review %d", reviewerID),
			})
		}(index, reviewerID)
	}
	close(start)
	reviewWG.Wait()
	reviewElapsed := time.Since(reviewStartedAt)
	for index := range reviewResponses {
		response := reviewResponses[index]
		if reviewErrors[index] != nil || response == nil || response.Base == nil ||
			response.Base.Code != 200 || response.Data == nil || response.Data.Id != seriesID {
			t.Fatalf("contract series concurrent review[%d] response=%+v err=%v",
				index, response, reviewErrors[index])
		}
	}
	assertP2ContractSeriesGenerated(t, ctx, db, seriesID)

	// A lost successful response can be retried by either concurrent reviewer.
	// Both calls must return the same generated batch without inserting again.
	for _, reviewerID := range reviewers {
		response, err := adminlogic.NewReviewContractSeriesLogic(
			p0AdminContext(ctx, reviewerID, p0AssetE2ETenantID), serviceCtx,
		).ReviewContractSeries(&option.ReviewContractSeriesReq{
			TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true,
			Reason: "response-loss retry must return the original generated series",
		})
		if err != nil || response == nil || response.Base == nil || response.Base.Code != 200 ||
			response.Data == nil || response.Data.GeneratedContractCount != p2SeriesContractNum {
			t.Fatalf("contract series review replay response=%+v err=%v", response, err)
		}
	}
	assertP2ContractSeriesGenerated(t, ctx, db, seriesID)

	var contractID, listTime int64
	if err := db.QueryRowContext(ctx, `SELECT contract.id,contract.list_time
		FROM t_option_contract_series_detail detail
		JOIN t_option_contract contract ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
		WHERE detail.tenant_id=? AND detail.series_id=? ORDER BY detail.id LIMIT 1`,
		p0AssetE2ETenantID, seriesID,
	).Scan(&contractID, &listTime); err != nil {
		t.Fatal(err)
	}
	listMessage := delayqueue.Message{
		Action: delayqueue.ActionListContract, TenantID: p0AssetE2ETenantID,
		ContractID: contractID, DueAt: listTime,
	}
	if err := handleDelayMessage(ctx, serviceCtx, listMessage); err != nil {
		t.Fatalf("pre-launch delayed listing: %v", err)
	}
	assertP2ContractStatus(t, ctx, db, contractID, option.ContractStatus_CONTRACT_STATUS_PENDING)

	selfLaunch, selfLaunchErr := adminlogic.NewReviewContractSeriesLaunchLogic(
		p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
	).ReviewContractSeriesLaunch(&option.ReviewContractSeriesLaunchReq{
		TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true, Reason: "self launch forbidden",
	})
	if selfLaunchErr == nil && selfLaunch != nil && selfLaunch.Base != nil && selfLaunch.Base.Code == 200 {
		t.Fatal("contract series creator approved its own launch")
	}
	launchers := []int64{p2SeriesLauncherA, p2SeriesLauncherB}
	start = make(chan struct{})
	launchResponses := make([]*option.GetContractSeriesResp, len(launchers))
	launchErrors := make([]error, len(launchers))
	var launchWG sync.WaitGroup
	launchWG.Add(len(launchers))
	for index, launcherID := range launchers {
		go func(index int, launcherID int64) {
			defer launchWG.Done()
			<-start
			launchResponses[index], launchErrors[index] = adminlogic.NewReviewContractSeriesLaunchLogic(
				p0AdminContext(ctx, launcherID, p0AssetE2ETenantID), serviceCtx,
			).ReviewContractSeriesLaunch(&option.ReviewContractSeriesLaunchReq{
				TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true,
				Reason: fmt.Sprintf("independent launch review %d", launcherID),
			})
		}(index, launcherID)
	}
	close(start)
	launchWG.Wait()
	for index := range launchResponses {
		if launchErrors[index] != nil || launchResponses[index] == nil ||
			launchResponses[index].Base == nil || launchResponses[index].Base.Code != 200 {
			t.Fatalf("contract series concurrent launch[%d] response=%+v err=%v",
				index, launchResponses[index], launchErrors[index])
		}
	}
	var launchStatus, launchReviewedBy int64
	if err := db.QueryRowContext(ctx, `SELECT launch_status,launch_reviewed_by
		FROM t_option_contract_series WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, seriesID,
	).Scan(&launchStatus, &launchReviewedBy); err != nil {
		t.Fatal(err)
	}
	if launchStatus != int64(option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_APPROVED) ||
		(launchReviewedBy != p2SeriesLauncherA && launchReviewedBy != p2SeriesLauncherB) {
		t.Fatalf("contract series launch status/reviewer=%d/%d", launchStatus, launchReviewedBy)
	}

	// Approval alone is insufficient. A delayed message must still preserve
	// PENDING while the market is missing or stale.
	if err := handleDelayMessage(ctx, serviceCtx, listMessage); err != nil {
		t.Fatalf("missing-market delayed listing: %v", err)
	}
	assertP2ContractStatus(t, ctx, db, contractID, option.ContractStatus_CONTRACT_STATUS_PENDING)
	marketID := insertP2ContractSeriesMarket(t, ctx, serviceCtx, contractID, now-31)
	if err := handleDelayMessage(ctx, serviceCtx, listMessage); err != nil {
		t.Fatalf("stale-market delayed listing: %v", err)
	}
	assertP2ContractStatus(t, ctx, db, contractID, option.ContractStatus_CONTRACT_STATUS_PENDING)
	market, err := serviceCtx.OptionMarketModel.FindOne(ctx, marketID)
	if err != nil {
		t.Fatal(err)
	}
	freshAt := time.Now().Unix()
	market.SnapshotTime = freshAt
	market.UnderlyingSnapshotTime = freshAt
	market.MarkSnapshotTime = freshAt
	market.GreeksSnapshotTime = freshAt
	market.UpdateTimes = freshAt
	if err = serviceCtx.OptionMarketModel.Update(ctx, market); err != nil {
		t.Fatal(err)
	}
	wrongDue := listMessage
	wrongDue.DueAt++
	if err = handleDelayMessage(ctx, serviceCtx, wrongDue); err != nil {
		t.Fatalf("stale delayed identity: %v", err)
	}
	assertP2ContractStatus(t, ctx, db, contractID, option.ContractStatus_CONTRACT_STATUS_PENDING)
	if err = handleDelayMessage(ctx, serviceCtx, listMessage); err != nil {
		t.Fatalf("eligible delayed listing: %v", err)
	}
	assertP2ContractStatus(t, ctx, db, contractID, option.ContractStatus_CONTRACT_STATUS_TRADING)

	testP2ContractSeriesCollisionRollback(t, ctx, db, serviceCtx, calendarCode, now)
	testP2ContractSeriesOversizeRejected(t, ctx, db, serviceCtx, calendarCode, now)
	assertP2ContractSeriesImmutable(t, ctx, db, seriesID, contractID)
	t.Logf("contract_series_capacity=contracts:%d expiries:2 strikes:125 create_concurrency:%d review:%s launch_reviewer:%d",
		p2SeriesContractNum, createConcurrency, reviewElapsed.Round(time.Millisecond), launchReviewedBy)
}

func p2ContractSeriesRequest(
	now int64,
	calendarCode, requestKey, seriesCode string,
	strikeCount int64,
) *option.CreateContractSeriesReq {
	return &option.CreateContractSeriesReq{
		TenantId: p0AssetE2ETenantID, RequestKey: requestKey, SeriesCode: seriesCode,
		ReferencePrice: "100", ReferenceSource: "P2-AUTHORITATIVE-INDEX-SNAPSHOT",
		ReferenceTime: now, EvidenceRef: "P2-CONTRACT-SERIES-AUTHORITATIVE-EVIDENCE",
		ChangeReason: "repository 500-contract atomic generation acceptance",
		ContractTemplate: &option.CreateContractReq{
			UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
			ExerciseStyle:  option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN,
			SettlementType: option.SettlementType_SETTLEMENT_TYPE_CASH,
			ContractUnit:   "1", MinOrderQty: "1", MaxOrderQty: "1000",
			PriceTick: "0.1", QtyStep: "1", Multiplier: "1",
			AutoExerciseThreshold: "1", MaxUserLongQty: "10000", MaxUserShortQty: "10000",
			MaxOpenInterest: "10000", OrderPriceBandRatio: "0.2", CircuitBreakerRatio: "0.5",
			GreeksMaxAgeSeconds: 60, SettlementPriceSource: "authoritative-market",
			SettlementPriceMethod: "MEDIAN", SettlementWindowSeconds: 60, SettlementMinSamples: 3,
			IsAutoExercise: common.YesNo_YES_NO_YES,
			MakerFeeRate:   "0.02", TakerFeeRate: "0.04", ExerciseFeeRate: "0.1",
			FeeUserId: 96890, FeeAccountId: 96891,
			SellerMarginMode:  option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED,
			InitialMarginRate: "0.5", MaintenanceMarginRate: "0.2", MinMarginRate: "0.1",
			LiquidationFeeRate: "0.1", InsuranceUserId: 96892, InsuranceAccountId: 96893,
			LiquidationDeficitPolicy: option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW,
			TradingCalendarCode:      calendarCode,
		},
		Expiries: []*option.ContractSeriesExpiryInput{
			{SequenceNo: 1, CycleCode: "WEEKLY", ListTime: now - 1,
				ExerciseCutoffTime: now + 3600, ExpireTime: now + 3600, DeliverTime: now + 3660},
			{SequenceNo: 2, CycleCode: "MONTHLY", ListTime: now - 1,
				ExerciseCutoffTime: now + 7200, ExpireTime: now + 7200, DeliverTime: now + 7260},
		},
		StrikeBands: []*option.ContractSeriesStrikeBandInput{{
			SequenceNo: 1, LowerStrike: "1", UpperStrike: fmt.Sprintf("%d", strikeCount), StrikeStep: "1",
		}},
	}
}

func assertP2ContractSeriesGenerated(t *testing.T, ctx context.Context, db *sql.DB, seriesID int64) {
	t.Helper()
	var status, launchStatus, expected, generated, reviewedBy int64
	var details, contracts, pending, calls, puts, expiries, pairs, distinctCodes, distinctContracts int64
	if err := db.QueryRowContext(ctx, `SELECT status,launch_status,expected_contract_count,
		generated_contract_count,reviewed_by FROM t_option_contract_series WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, seriesID,
	).Scan(&status, &launchStatus, &expected, &generated, &reviewedBy); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),COUNT(DISTINCT detail.contract_code),
		COUNT(DISTINCT detail.contract_id),SUM(detail.option_type=1),SUM(detail.option_type=2),
		COUNT(DISTINCT detail.expiry_id),
		COUNT(DISTINCT CONCAT(detail.expiry_id,':',CAST(detail.strike_price AS CHAR)))
		FROM t_option_contract_series_detail detail WHERE detail.tenant_id=? AND detail.series_id=?`,
		p0AssetE2ETenantID, seriesID,
	).Scan(&details, &distinctCodes, &distinctContracts, &calls, &puts, &expiries, &pairs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(contract.status=1)
		FROM t_option_contract_series_detail detail
		JOIN t_option_contract contract ON contract.tenant_id=detail.tenant_id AND contract.id=detail.contract_id
		WHERE detail.tenant_id=? AND detail.series_id=?`,
		p0AssetE2ETenantID, seriesID,
	).Scan(&contracts, &pending); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED) ||
		launchStatus != int64(option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_PENDING_REVIEW) ||
		expected != p2SeriesContractNum || generated != p2SeriesContractNum ||
		details != p2SeriesContractNum || contracts != p2SeriesContractNum || pending != p2SeriesContractNum ||
		distinctCodes != p2SeriesContractNum || distinctContracts != p2SeriesContractNum ||
		calls != 250 || puts != 250 || expiries != 2 || pairs != 250 ||
		(reviewedBy != p2SeriesReviewerA && reviewedBy != p2SeriesReviewerB) {
		t.Fatalf("series generated status/launch/expected/generated/reviewer details/contracts/pending/codes/ids/call/put/expiries/pairs=%d/%d/%d/%d/%d %d/%d/%d/%d/%d/%d/%d/%d/%d",
			status, launchStatus, expected, generated, reviewedBy, details, contracts, pending,
			distinctCodes, distinctContracts, calls, puts, expiries, pairs)
	}
}

func assertP2ContractStatus(
	t *testing.T, ctx context.Context, db *sql.DB, contractID int64, expected option.ContractStatus,
) {
	t.Helper()
	var status int64
	if err := db.QueryRowContext(ctx, `SELECT status FROM t_option_contract WHERE tenant_id=? AND id=?`,
		p0AssetE2ETenantID, contractID,
	).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != int64(expected) {
		t.Fatalf("contract %d status=%d want=%d", contractID, status, expected)
	}
}

func insertP2ContractSeriesMarket(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, contractID, snapshotTime int64,
) int64 {
	t.Helper()
	market := &models.TOptionMarket{
		TenantId: p0AssetE2ETenantID, ContractId: contractID,
		UnderlyingPrice: decimal.NewFromInt(100), MarkPrice: decimal.NewFromInt(10),
		LastPrice: decimal.NewFromInt(10), BidPrice: decimal.NewFromInt(9), AskPrice: decimal.NewFromInt(11),
		TheoreticalPrice: decimal.NewFromInt(10), IntrinsicValue: decimal.Zero, TimeValue: decimal.NewFromInt(10),
		Iv: decimal.RequireFromString("0.5"), Delta: decimal.RequireFromString("0.5"),
		Gamma: decimal.RequireFromString("0.01"), Theta: decimal.RequireFromString("-0.01"),
		Vega: decimal.RequireFromString("0.1"), Rho: decimal.RequireFromString("0.01"),
		RiskFreeRate: decimal.RequireFromString("0.02"), PricingModel: "BLACK_SCHOLES",
		SnapshotTime: snapshotTime, UnderlyingSnapshotTime: snapshotTime,
		MarkSnapshotTime: snapshotTime, GreeksSnapshotTime: snapshotTime,
		CreateTimes: snapshotTime, UpdateTimes: snapshotTime,
	}
	result, err := serviceCtx.OptionMarketModel.Insert(ctx, market)
	if err != nil {
		t.Fatal(err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func testP2ContractSeriesCollisionRollback(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	insertP0OrderTestContract(
		t, ctx, serviceCtx, "P2COLLIDE-V1-E001-K001-P", calendarCode, 96890, now,
	)
	request := p2ContractSeriesRequest(now, calendarCode, "P2-SERIES-COLLISION-REQUEST", "P2COLLIDE", 1)
	request.Expiries = request.Expiries[:1]
	created, err := adminlogic.NewCreateContractSeriesLogic(
		p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateContractSeries(request)
	if err != nil || created == nil || created.Base == nil || created.Base.Code != 200 || created.Data == nil {
		t.Fatalf("create collision series response=%+v err=%v", created, err)
	}
	seriesID := created.Data.Id
	reviewed, reviewErr := adminlogic.NewReviewContractSeriesLogic(
		p0AdminContext(ctx, p2SeriesReviewerA, p0AssetE2ETenantID), serviceCtx,
	).ReviewContractSeries(&option.ReviewContractSeriesReq{
		TenantId: p0AssetE2ETenantID, SeriesId: seriesID, Approve: true,
		Reason: "collision must roll back the preceding call contract",
	})
	if reviewErr == nil && reviewed != nil && reviewed.Base != nil && reviewed.Base.Code == 200 {
		t.Fatal("contract series collision was unexpectedly approved")
	}
	var status, generated, details, codeRows int64
	if err = db.QueryRowContext(ctx, `SELECT
		(SELECT status FROM t_option_contract_series WHERE tenant_id=? AND id=?),
		(SELECT generated_contract_count FROM t_option_contract_series WHERE tenant_id=? AND id=?),
		(SELECT COUNT(*) FROM t_option_contract_series_detail WHERE tenant_id=? AND series_id=?),
		(SELECT COUNT(*) FROM t_option_contract WHERE tenant_id=? AND contract_code LIKE 'P2COLLIDE-V1-%')`,
		p0AssetE2ETenantID, seriesID, p0AssetE2ETenantID, seriesID,
		p0AssetE2ETenantID, seriesID, p0AssetE2ETenantID,
	).Scan(&status, &generated, &details, &codeRows); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_PENDING_REVIEW) ||
		generated != 0 || details != 0 || codeRows != 1 {
		t.Fatalf("collision rollback status/generated/details/codes=%d/%d/%d/%d", status, generated, details, codeRows)
	}
}

func testP2ContractSeriesOversizeRejected(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) {
	t.Helper()
	request := p2ContractSeriesRequest(now, calendarCode, "P2-SERIES-OVERSIZE-REQUEST", "P2OVER", 251)
	request.Expiries = request.Expiries[:1]
	response, err := adminlogic.NewCreateContractSeriesLogic(
		p0AdminContext(ctx, p2SeriesCreatorID, p0AssetE2ETenantID), serviceCtx,
	).CreateContractSeries(request)
	if err == nil && response != nil && response.Base != nil && response.Base.Code == 200 {
		t.Fatal("502-contract series was accepted")
	}
	var rows int64
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_contract_series
		WHERE tenant_id=? AND request_key=?`, p0AssetE2ETenantID, request.RequestKey).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 0 {
		t.Fatalf("oversize contract series left %d rows", rows)
	}
}

func assertP2ContractSeriesImmutable(
	t *testing.T, ctx context.Context, db *sql.DB, seriesID, contractID int64,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_contract_series_strike_band
		SET upper_strike=upper_strike+1 WHERE tenant_id=? AND series_id=? LIMIT 1`,
		p0AssetE2ETenantID, seriesID,
	); err == nil {
		t.Fatal("submitted contract-series strike evidence was mutable")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM t_option_contract_series_detail
		WHERE tenant_id=? AND series_id=? LIMIT 1`, p0AssetE2ETenantID, seriesID); err == nil {
		t.Fatal("contract-series generation lineage was deletable")
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_contract SET strike_price=strike_price+1
		WHERE tenant_id=? AND id=?`, p0AssetE2ETenantID, contractID); err == nil {
		t.Fatal("generated contract economics were mutable")
	}
}

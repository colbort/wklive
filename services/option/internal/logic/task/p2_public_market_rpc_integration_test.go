package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"wklive/proto/common"
	"wklive/proto/option"
	applogic "wklive/services/option/internal/logic/app"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"

	"github.com/shopspring/decimal"
)

const p2PublicOtherTenantID int64 = 996032

type p2PublicContractSpec struct {
	tenantID   int64
	code       string
	underlying string
	expireTime int64
	strike     string
	optionType option.OptionType
	status     option.ContractStatus
}

func testP2PublicMarketAcceptance(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	now := time.Now().Unix()
	sparseExpiry := now + 120000
	capacityExpiry := now + 180000

	sparse := []p2PublicContractSpec{
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-100-C", "P2CHAINSPARSE", sparseExpiry, "100", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_TRADING},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-100-P", "P2CHAINSPARSE", sparseExpiry, "100", option.OptionType_OPTION_TYPE_PUT, option.ContractStatus_CONTRACT_STATUS_TRADING},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-110-C", "P2CHAINSPARSE", sparseExpiry, "110", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_TRADING},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-120-C", "P2CHAINSPARSE", sparseExpiry, "120", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_PAUSED},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-120-P", "P2CHAINSPARSE", sparseExpiry, "120", option.OptionType_OPTION_TYPE_PUT, option.ContractStatus_CONTRACT_STATUS_PAUSED},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-130-C", "P2CHAINSPARSE", sparseExpiry, "130", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_PENDING},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-140-C", "P2CHAINSPARSE", sparseExpiry, "140", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_EXPIRED},
		{p0AssetE2ETenantID, "P2-PUB-SPARSE-OTHER-C", "P2CHAINSPARSE", sparseExpiry + 1, "999", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_TRADING},
	}
	sparseIDs := seedP2PublicContracts(t, ctx, db, sparse, now)
	callID, putID := sparseIDs[0], sparseIDs[1]

	capacitySpecs := make([]p2PublicContractSpec, 0, 500)
	for strike := 1; strike <= 250; strike++ {
		for _, optionType := range []option.OptionType{
			option.OptionType_OPTION_TYPE_CALL,
			option.OptionType_OPTION_TYPE_PUT,
		} {
			suffix := "C"
			if optionType == option.OptionType_OPTION_TYPE_PUT {
				suffix = "P"
			}
			capacitySpecs = append(capacitySpecs, p2PublicContractSpec{
				tenantID: p0AssetE2ETenantID, code: fmt.Sprintf("P2-PUB-500-%03d-%s", strike, suffix),
				underlying: "P2CHAIN500", expireTime: capacityExpiry, strike: fmt.Sprintf("%d", strike),
				optionType: optionType, status: option.ContractStatus_CONTRACT_STATUS_TRADING,
			})
		}
	}
	capacityStarted := time.Now()
	capacityIDs := seedP2PublicContracts(t, ctx, db, capacitySpecs, now)
	seedDuration := time.Since(capacityStarted)
	bookContractID := capacityIDs[0]

	seedP2PublicMarkets(t, ctx, db, callID, now)
	seedP2PublicTrades(t, ctx, db, callID, putID, now)
	seedP2PublicPositions(t, ctx, db, callID, putID, now)
	seedP2PublicOrderBook(t, ctx, db, bookContractID, now)

	mainCtx := p0AdminContext(ctx, 96901, p0AssetE2ETenantID)
	sparseResp := callP2PublicChain(t, mainCtx, serviceCtx, "P2CHAINSPARSE", sparseExpiry, option.ContractStatus_CONTRACT_STATUS_UNKNOWN)
	assertP2SparseChain(t, sparseResp, callID, putID)

	pausedResp := callP2PublicChain(t, mainCtx, serviceCtx, "P2CHAINSPARSE", sparseExpiry, option.ContractStatus_CONTRACT_STATUS_PAUSED)
	if len(pausedResp.Data) != 1 || pausedResp.Data[0].StrikePrice != "120" ||
		pausedResp.Data[0].Call == nil || pausedResp.Data[0].Put == nil {
		t.Fatalf("paused chain leaked or lost contracts: %+v", pausedResp.Data)
	}
	pendingResp, pendingErr := applogic.NewListOptionChainLogic(mainCtx, serviceCtx).ListOptionChain(
		&option.ListOptionChainReq{UnderlyingSymbol: "P2CHAINSPARSE", ExpireTime: sparseExpiry,
			Status: option.ContractStatus_CONTRACT_STATUS_PENDING},
	)
	if pendingErr != nil || pendingResp == nil || pendingResp.Base == nil || pendingResp.Base.Code == 200 {
		t.Fatalf("public chain accepted PENDING status resp=%+v err=%v", pendingResp, pendingErr)
	}

	capacityQueryStarted := time.Now()
	capacityResp := callP2PublicChain(t, mainCtx, serviceCtx, "P2CHAIN500", capacityExpiry, option.ContractStatus_CONTRACT_STATUS_UNKNOWN)
	capacityQueryDuration := time.Since(capacityQueryStarted)
	assertP2CapacityChain(t, capacityResp, 250, 500)

	bookResp := callP2PublicBook(t, mainCtx, serviceCtx, bookContractID, 100)
	assertP2PublicBook(t, bookResp, false)
	tooDeep, tooDeepErr := applogic.NewGetOrderBookLogic(mainCtx, serviceCtx).GetOrderBook(
		&option.GetOrderBookReq{ContractId: bookContractID, DepthLimit: 101},
	)
	if tooDeepErr != nil || tooDeep == nil || tooDeep.Base == nil || tooDeep.Base.Code == 200 || tooDeep.Data != nil {
		t.Fatalf("public book accepted depth 101 resp=%+v err=%v", tooDeep, tooDeepErr)
	}

	assertP2PublicTradeBoundary(t, ctx, serviceCtx, callID, now)
	testP2PublicSnapshotConsistency(t, ctx, db, serviceCtx, mainCtx, bookContractID, now)
	p95 := testP2PublicConcurrentCapacity(t, mainCtx, serviceCtx, capacityExpiry, bookContractID)
	testP2PublicTenantIsolation(t, ctx, db, serviceCtx, capacityExpiry, callID, bookContractID, now)

	overflowSpec := p2PublicContractSpec{
		tenantID: p0AssetE2ETenantID, code: "P2-PUB-501-251-C", underlying: "P2CHAIN500",
		expireTime: capacityExpiry, strike: "251", optionType: option.OptionType_OPTION_TYPE_CALL,
		status: option.ContractStatus_CONTRACT_STATUS_TRADING,
	}
	seedP2PublicContracts(t, ctx, db, []p2PublicContractSpec{overflowSpec}, now)
	overflowResp, overflowErr := applogic.NewListOptionChainLogic(mainCtx, serviceCtx).ListOptionChain(
		&option.ListOptionChainReq{UnderlyingSymbol: "P2CHAIN500", ExpireTime: capacityExpiry},
	)
	if overflowErr != nil || overflowResp == nil || overflowResp.Base == nil ||
		overflowResp.Base.Code == 200 || len(overflowResp.Data) != 0 {
		t.Fatalf("501-contract chain was not rejected atomically resp=%+v err=%v", overflowResp, overflowErr)
	}

	t.Logf("public_market_capacity=contracts:500 strikes:250 book_levels:100x2 concurrent:16 seed:%s single_chain:%s concurrent_p95:%s",
		seedDuration, capacityQueryDuration, p95)
}

func seedP2PublicContracts(
	t *testing.T, ctx context.Context, db *sql.DB, specs []p2PublicContractSpec, now int64,
) []int64 {
	t.Helper()
	stmt, err := db.PrepareContext(ctx, `INSERT INTO t_option_contract (
		tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
		option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
		max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
		deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
		circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
		maintenance_margin_rate,min_margin_rate,status,is_deleted,create_times,update_times
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	ids := make([]int64, 0, len(specs))
	for _, spec := range specs {
		result, execErr := stmt.ExecContext(ctx,
			spec.tenantID, spec.code, spec.underlying, "BTC", "USDT", "USDT",
			int64(spec.optionType), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
			int64(option.SettlementType_SETTLEMENT_TYPE_CASH), spec.strike, "1", "1", "1000",
			"0.1", "1", "1", now-3600, spec.expireTime-60, spec.expireTime, spec.expireTime+60,
			"10000", "10000", "10000", "0.2", "0.5", 60,
			int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED), "0.5", "0.2", "0.1",
			int64(spec.status), int64(common.YesNo_YES_NO_NO), now, now,
		)
		if execErr != nil {
			t.Fatalf("seed public contract %s: %v", spec.code, execErr)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			t.Fatal(idErr)
		}
		ids = append(ids, id)
	}
	return ids
}

func seedP2PublicMarkets(t *testing.T, ctx context.Context, db *sql.DB, contractID, now int64) {
	t.Helper()
	for _, row := range []struct {
		tenantID int64
		mark     string
	}{
		{p0AssetE2ETenantID, "10"},
		{p2PublicOtherTenantID, "999"},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO t_option_market (
			tenant_id,contract_id,underlying_price,mark_price,last_price,bid_price,ask_price,
			snapshot_time,underlying_snapshot_time,mark_snapshot_time,greeks_snapshot_time,
			create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			row.tenantID, contractID, "100", row.mark, row.mark, row.mark, row.mark,
			now, now-1, now-2, now-3, now, now,
		); err != nil {
			t.Fatalf("seed public market tenant=%d: %v", row.tenantID, err)
		}
	}
}

func seedP2PublicTrades(t *testing.T, ctx context.Context, db *sql.DB, callID, putID, now int64) {
	t.Helper()
	start := now - 86400
	rows := []struct {
		tenantID, contractID, sequence, tradeTime int64
		no, qty, turnover                         string
	}{
		{p0AssetE2ETenantID, callID, 1, start - 1, "P2-PUB-TRADE-BEFORE", "1", "10"},
		{p0AssetE2ETenantID, callID, 2, start, "P2-PUB-TRADE-START", "2", "20"},
		{p0AssetE2ETenantID, callID, 3, start + 1, "P2-PUB-TRADE-AFTER", "3", "30"},
		{p0AssetE2ETenantID, callID, 4, now, "P2-PUB-TRADE-END", "4", "40"},
		{p0AssetE2ETenantID, callID, 5, now + 1, "P2-PUB-TRADE-FUTURE", "5", "50"},
		{p0AssetE2ETenantID, putID, 1, now - 100, "P2-PUB-PUT-SAFE", "7", "70"},
		{p2PublicOtherTenantID, callID, 999, now - 100, "P2-PUB-CROSS-TENANT", "999", "9990"},
	}
	for _, row := range rows {
		if _, err := db.ExecContext(ctx, `INSERT INTO t_option_trade (
			tenant_id,trade_no,contract_id,underlying_symbol,price,qty,turnover,
			match_sequence,trade_time,create_times
		) VALUES (?,?,?,?,?,?,?,?,?,?)`,
			row.tenantID, row.no, row.contractID, "P2CHAINSPARSE", "10", row.qty, row.turnover,
			row.sequence, row.tradeTime, row.tradeTime,
		); err != nil {
			t.Fatalf("seed public trade %s: %v", row.no, err)
		}
	}
}

func seedP2PublicPositions(t *testing.T, ctx context.Context, db *sql.DB, callID, putID, now int64) {
	t.Helper()
	rows := []struct {
		tenantID, userID, accountID, contractID int64
		side                                    common.PositionSide
		qty                                     string
		status                                  option.PositionStatus
	}{
		{p0AssetE2ETenantID, 96911, 1, callID, common.PositionSide_POSITION_SIDE_LONG, "8", option.PositionStatus_POSITION_STATUS_HOLDING},
		{p0AssetE2ETenantID, 96912, 1, callID, common.PositionSide_POSITION_SIDE_SHORT, "7", option.PositionStatus_POSITION_STATUS_HOLDING},
		{p0AssetE2ETenantID, 96913, 1, callID, common.PositionSide_POSITION_SIDE_LONG, "100", option.PositionStatus_POSITION_STATUS_CLOSED},
		{p0AssetE2ETenantID, 96914, 1, putID, common.PositionSide_POSITION_SIDE_LONG, "4", option.PositionStatus_POSITION_STATUS_HOLDING},
		{p0AssetE2ETenantID, 96915, 1, putID, common.PositionSide_POSITION_SIDE_SHORT, "4", option.PositionStatus_POSITION_STATUS_HOLDING},
		{p2PublicOtherTenantID, 96916, 1, callID, common.PositionSide_POSITION_SIDE_LONG, "999", option.PositionStatus_POSITION_STATUS_HOLDING},
	}
	for _, row := range rows {
		if _, err := db.ExecContext(ctx, `INSERT INTO t_option_position (
			tenant_id,user_id,account_id,contract_id,underlying_symbol,side,position_qty,
			available_qty,status,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
			row.tenantID, row.userID, row.accountID, row.contractID, "P2CHAINSPARSE",
			int64(row.side), row.qty, row.qty, int64(row.status), now, now,
		); err != nil {
			t.Fatalf("seed public position tenant=%d contract=%d: %v", row.tenantID, row.contractID, err)
		}
	}
}

func seedP2PublicOrderBook(t *testing.T, ctx context.Context, db *sql.DB, contractID, now int64) {
	t.Helper()
	insert := func(tenantID, userID int64, no string, side common.Side, orderType option.OrderType,
		price, qty string, status option.OrderStatus, comboID, comboLeg int64,
	) int64 {
		result, err := db.ExecContext(ctx, `INSERT INTO t_option_order (
			tenant_id,order_no,user_id,account_id,contract_id,underlying_symbol,side,position_effect,
			order_type,price,qty,unfilled_qty,status,combo_order_id,combo_leg_no,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			tenantID, no, userID, 1, contractID, "P2CHAIN500", int64(side),
			int64(option.PositionEffect_POSITION_EFFECT_OPEN), int64(orderType), price, qty, qty,
			int64(status), comboID, comboLeg, now, now,
		)
		if err != nil {
			t.Fatalf("seed public order %s: %v", no, err)
		}
		id, err := result.LastInsertId()
		if err != nil {
			t.Fatal(err)
		}
		return id
	}
	for level := 0; level < 100; level++ {
		insert(p0AssetE2ETenantID, 97000+int64(level), fmt.Sprintf("P2-PUB-BID-%03d", level),
			common.Side_SIDE_BUY, option.OrderType_ORDER_TYPE_LIMIT, fmt.Sprintf("%d", 1000+level), "1",
			option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
		insert(p0AssetE2ETenantID, 97100+int64(level), fmt.Sprintf("P2-PUB-ASK-%03d", level),
			common.Side_SIDE_SELL, option.OrderType_ORDER_TYPE_POST_ONLY, fmt.Sprintf("%d", 2000+level), "1",
			option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
	}
	insert(p0AssetE2ETenantID, 97201, "P2-PUB-BID-AGGREGATE", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_POST_ONLY, "1000", "2", option.OrderStatus_ORDER_STATUS_PART_FILLED, 0, 0)
	insert(p0AssetE2ETenantID, 97202, "P2-PUB-EXCLUDE-IOC", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_IOC, "1000", "50", option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
	insert(p0AssetE2ETenantID, 97203, "P2-PUB-EXCLUDE-FOK", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_FOK, "1000", "50", option.OrderStatus_ORDER_STATUS_PART_FILLED, 0, 0)
	insert(p0AssetE2ETenantID, 97204, "P2-PUB-EXCLUDE-MARKET", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_MARKET, "1000", "50", option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
	insert(p0AssetE2ETenantID, 97205, "P2-PUB-EXCLUDE-FUNDING", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_LIMIT, "1000", "50", option.OrderStatus_ORDER_STATUS_FUNDING, 0, 0)
	insert(p0AssetE2ETenantID, 97206, "P2-PUB-EXCLUDE-CANCELED", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_LIMIT, "1000", "50", option.OrderStatus_ORDER_STATUS_CANCELED, 0, 0)
	insert(p0AssetE2ETenantID, 97207, "P2-PUB-EXCLUDE-ZERO", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_LIMIT, "1000", "0", option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
	insert(p0AssetE2ETenantID, 97208, "P2-PUB-EXCLUDE-COMBO", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_LIMIT, "1000", "50", option.OrderStatus_ORDER_STATUS_PENDING, 888888, 1)
	insert(p2PublicOtherTenantID, 97209, "P2-PUB-EXCLUDE-TENANT", common.Side_SIDE_BUY,
		option.OrderType_ORDER_TYPE_LIMIT, "1000", "999", option.OrderStatus_ORDER_STATUS_PENDING, 0, 0)
}

func callP2PublicChain(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext,
	underlying string, expiry int64, status option.ContractStatus,
) *option.ListOptionChainResp {
	t.Helper()
	resp, err := applogic.NewListOptionChainLogic(ctx, serviceCtx).ListOptionChain(
		&option.ListOptionChainReq{UnderlyingSymbol: underlying, ExpireTime: expiry, Status: status},
	)
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 {
		t.Fatalf("list public chain underlying=%s expiry=%d status=%d resp=%+v err=%v",
			underlying, expiry, status, resp, err)
	}
	return resp
}

func callP2PublicBook(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, contractID int64, depth int32,
) *option.GetOrderBookResp {
	t.Helper()
	resp, err := applogic.NewGetOrderBookLogic(ctx, serviceCtx).GetOrderBook(
		&option.GetOrderBookReq{ContractId: contractID, DepthLimit: depth},
	)
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil {
		t.Fatalf("get public book contract=%d depth=%d resp=%+v err=%v", contractID, depth, resp, err)
	}
	return resp
}

func assertP2SparseChain(t *testing.T, resp *option.ListOptionChainResp, callID, putID int64) {
	t.Helper()
	if len(resp.Data) != 2 || resp.Data[0].StrikePrice != "100" || resp.Data[1].StrikePrice != "110" ||
		resp.Data[0].Call == nil || resp.Data[0].Put == nil || resp.Data[1].Call == nil || resp.Data[1].Put != nil {
		t.Fatalf("sparse chain pair/missing-leg result=%+v", resp.Data)
	}
	call := resp.Data[0].Call
	put := resp.Data[0].Put
	if call.Contract.Id != callID || put.Contract.Id != putID || call.Market == nil ||
		call.Market.MarkPrice != "10" || call.Market.UnderlyingSnapshotTime == 0 ||
		call.Market.MarkSnapshotTime == 0 || call.Market.GreeksSnapshotTime == 0 {
		t.Fatalf("sparse chain contract/market identity call=%+v put=%+v", call, put)
	}
	if call.Statistics.OpenInterest != "8" || call.Statistics.LongOpenInterest != "8" ||
		call.Statistics.ShortOpenInterest != "7" || call.Statistics.OiBalanced ||
		put.Statistics.OpenInterest != "4" || !put.Statistics.OiBalanced ||
		put.Statistics.Volume_24H != "7" || put.Statistics.Turnover_24H != "70" ||
		put.Statistics.TradeCount_24H != 1 {
		t.Fatalf("public statistics/OI mismatch call=%+v put=%+v", call.Statistics, put.Statistics)
	}
	if resp.GeneratedAt <= 0 || resp.StatisticsWindowStart != resp.GeneratedAt-86400 ||
		call.Statistics.StatisticsAsOf != resp.GeneratedAt ||
		call.Statistics.StatisticsWindowStart != resp.StatisticsWindowStart {
		t.Fatalf("public snapshot times resp=%+v stats=%+v", resp, call.Statistics)
	}
}

func assertP2CapacityChain(t *testing.T, resp *option.ListOptionChainResp, rows, legs int) {
	t.Helper()
	if len(resp.Data) != rows {
		t.Fatalf("capacity chain rows=%d want=%d", len(resp.Data), rows)
	}
	actualLegs := 0
	for index, row := range resp.Data {
		wantStrike := fmt.Sprintf("%d", index+1)
		if row.StrikePrice != wantStrike || row.Call == nil || row.Put == nil ||
			row.Call.Contract.ExpireTime != row.Put.Contract.ExpireTime {
			t.Fatalf("capacity chain row[%d]=%+v want strike=%s paired", index, row, wantStrike)
		}
		actualLegs += 2
	}
	if actualLegs != legs {
		t.Fatalf("capacity chain legs=%d want=%d", actualLegs, legs)
	}
}

func assertP2PublicBook(t *testing.T, resp *option.GetOrderBookResp, snapshotLevel bool) {
	t.Helper()
	book := resp.Data
	if len(book.Bids) != 100 || len(book.Asks) != 100 || book.Source != "OPTION_ACTIVE_LIMIT_ORDERS" {
		t.Fatalf("public book level/source mismatch bids=%d asks=%d source=%s", len(book.Bids), len(book.Asks), book.Source)
	}
	firstBid := 1099
	if snapshotLevel {
		firstBid = 5000
	}
	if book.Bids[0].Price != fmt.Sprintf("%d", firstBid) || book.Asks[0].Price != "2000" ||
		book.Asks[len(book.Asks)-1].Price != "2099" {
		t.Fatalf("public book sort mismatch first/last=%+v/%+v/%+v",
			book.Bids[0], book.Asks[0], book.Asks[len(book.Asks)-1])
	}
	foundAggregate := false
	for _, level := range book.Bids {
		if level.Price == "1000" {
			foundAggregate = true
			if level.Qty != "3" || level.OrderCount != 2 {
				t.Fatalf("public book aggregate includes non-resting orders level=%+v", level)
			}
		}
	}
	if !snapshotLevel && !foundAggregate {
		t.Fatal("public book lost the 1000 aggregate level")
	}
}

func assertP2PublicTradeBoundary(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, contractID, now int64,
) {
	t.Helper()
	stats, err := serviceCtx.OptionTradeModel.FindStatisticsByContracts(
		ctx, p0AssetE2ETenantID, []int64{contractID}, now-86400, now,
	)
	if err != nil || len(stats) != 1 {
		t.Fatalf("public trade boundary query stats=%+v err=%v", stats, err)
	}
	if !stats[0].Volume.Equal(decimal.NewFromInt(9)) ||
		!stats[0].Turnover.Equal(decimal.NewFromInt(90)) || stats[0].TradeCount != 3 {
		t.Fatalf("24h closed boundary stats=%+v want volume=9 turnover=90 count=3", stats[0])
	}
}

func testP2PublicSnapshotConsistency(
	t *testing.T, ctx context.Context, db *sql.DB, serviceCtx *svc.ServiceContext,
	readCtx context.Context, contractID, now int64,
) {
	t.Helper()
	result, err := db.ExecContext(ctx, `INSERT INTO t_option_order (
		tenant_id,order_no,user_id,account_id,contract_id,underlying_symbol,side,position_effect,
		order_type,price,qty,unfilled_qty,status,create_times,update_times
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p0AssetE2ETenantID, "P2-PUB-SNAPSHOT-BID", 97301, 1, contractID, "P2CHAIN500",
		int64(common.Side_SIDE_BUY), int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		int64(option.OrderType_ORDER_TYPE_LIMIT), "5000", "1", "1",
		int64(option.OrderStatus_ORDER_STATUS_PENDING), now, now,
	)
	if err != nil {
		t.Fatal(err)
	}
	orderID, err := result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}

	start := make(chan struct{})
	errCh := make(chan error, 256)
	var beforeCount atomic.Int64
	var afterCount atomic.Int64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		for iteration := 0; iteration < 40; iteration++ {
			tx, txErr := db.BeginTx(ctx, nil)
			if txErr != nil {
				errCh <- txErr
				return
			}
			if iteration%2 == 0 {
				_, txErr = tx.ExecContext(ctx, `DELETE FROM t_option_trade
					WHERE tenant_id=? AND contract_id=? AND match_sequence=10000`, p0AssetE2ETenantID, contractID)
				if txErr == nil {
					_, txErr = tx.ExecContext(ctx, `UPDATE t_option_order SET status=?,update_times=? WHERE id=?`,
						int64(option.OrderStatus_ORDER_STATUS_PENDING), now+int64(iteration), orderID)
				}
			} else {
				_, txErr = tx.ExecContext(ctx, `UPDATE t_option_order SET status=?,update_times=? WHERE id=?`,
					int64(option.OrderStatus_ORDER_STATUS_CANCELED), now+int64(iteration), orderID)
				if txErr == nil {
					_, txErr = tx.ExecContext(ctx, `INSERT INTO t_option_trade (
						tenant_id,trade_no,contract_id,underlying_symbol,price,qty,turnover,
						match_sequence,trade_time,create_times
					) VALUES (?,?,?,?,?,?,?,?,?,?)`,
						p0AssetE2ETenantID, "P2-PUB-SNAPSHOT-TRADE", contractID, "P2CHAIN500",
						"1", "1", "1", 10000, now, now,
					)
				}
			}
			if txErr != nil {
				_ = tx.Rollback()
				errCh <- txErr
				return
			}
			if txErr = tx.Commit(); txErr != nil {
				errCh <- txErr
				return
			}
			time.Sleep(time.Millisecond)
		}
	}()
	const readers = 24
	for reader := 0; reader < readers; reader++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for iteration := 0; iteration < 6; iteration++ {
				resp, callErr := applogic.NewGetOrderBookLogic(readCtx, serviceCtx).GetOrderBook(
					&option.GetOrderBookReq{ContractId: contractID, DepthLimit: 100},
				)
				if callErr != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil {
					errCh <- fmt.Errorf("snapshot response=%+v err=%v", resp, callErr)
					return
				}
				hasLevel := len(resp.Data.Bids) > 0 && resp.Data.Bids[0].Price == "5000"
				switch {
				case hasLevel && resp.Data.LastMatchSequence == 0:
					beforeCount.Add(1)
				case !hasLevel && resp.Data.LastMatchSequence == 10000:
					afterCount.Add(1)
				default:
					errCh <- fmt.Errorf("torn public snapshot level=%t sequence=%d", hasLevel, resp.Data.LastMatchSequence)
					return
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	close(errCh)
	for gotErr := range errCh {
		if gotErr != nil {
			t.Fatal(gotErr)
		}
	}
	if beforeCount.Load()+afterCount.Load() != readers*6 {
		t.Fatalf("snapshot reads=%d/%d want=%d", beforeCount.Load(), afterCount.Load(), readers*6)
	}
	t.Logf("public_market_snapshot=before:%d after:%d torn:0", beforeCount.Load(), afterCount.Load())
}

func testP2PublicConcurrentCapacity(
	t *testing.T, ctx context.Context, serviceCtx *svc.ServiceContext, expiry, contractID int64,
) time.Duration {
	t.Helper()
	const concurrency = 16
	durations := make([]time.Duration, concurrency)
	errCh := make(chan error, concurrency)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for index := 0; index < concurrency; index++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			<-start
			started := time.Now()
			chain, chainErr := applogic.NewListOptionChainLogic(ctx, serviceCtx).ListOptionChain(
				&option.ListOptionChainReq{UnderlyingSymbol: "P2CHAIN500", ExpireTime: expiry},
			)
			if chainErr != nil || chain == nil || chain.Base == nil || chain.Base.Code != 200 || len(chain.Data) != 250 {
				errCh <- fmt.Errorf("concurrent chain[%d] rows=%d resp=%+v err=%v",
					index, func() int {
						if chain == nil {
							return -1
						}
						return len(chain.Data)
					}(), chain, chainErr)
				return
			}
			book, bookErr := applogic.NewGetOrderBookLogic(ctx, serviceCtx).GetOrderBook(
				&option.GetOrderBookReq{ContractId: contractID, DepthLimit: 100},
			)
			if bookErr != nil || book == nil || book.Base == nil || book.Base.Code != 200 ||
				book.Data == nil || len(book.Data.Bids) != 100 || len(book.Data.Asks) != 100 {
				errCh <- fmt.Errorf("concurrent book[%d] resp=%+v err=%v", index, book, bookErr)
				return
			}
			durations[index] = time.Since(started)
		}(index)
	}
	close(start)
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	return durations[(concurrency*95+99)/100-1]
}

func testP2PublicTenantIsolation(
	t *testing.T, ctx context.Context, db *sql.DB, serviceCtx *svc.ServiceContext,
	expiry, pollutedContractID, mainBookContractID, now int64,
) {
	t.Helper()
	seedOpenTradingCalendarForTenant(
		t, ctx, db, p2PublicOtherTenantID, logichelpers.DefaultTradingCalendarCode, now,
	)
	otherSpecs := []p2PublicContractSpec{
		{p2PublicOtherTenantID, "P2-PUB-OTHER-9999-C", "P2CHAIN500", expiry, "9999", option.OptionType_OPTION_TYPE_CALL, option.ContractStatus_CONTRACT_STATUS_TRADING},
		{p2PublicOtherTenantID, "P2-PUB-OTHER-9999-P", "P2CHAIN500", expiry, "9999", option.OptionType_OPTION_TYPE_PUT, option.ContractStatus_CONTRACT_STATUS_TRADING},
	}
	otherIDs := seedP2PublicContracts(t, ctx, db, otherSpecs, now)
	if _, err := db.ExecContext(ctx, `INSERT INTO t_option_order (
		tenant_id,order_no,user_id,account_id,contract_id,underlying_symbol,side,position_effect,
		order_type,price,qty,unfilled_qty,status,create_times,update_times
	) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		p2PublicOtherTenantID, "P2-PUB-OTHER-BOOK", 97401, 1, otherIDs[0], "P2CHAIN500",
		int64(common.Side_SIDE_BUY), int64(option.PositionEffect_POSITION_EFFECT_OPEN),
		int64(option.OrderType_ORDER_TYPE_LIMIT), "8888", "8", "8",
		int64(option.OrderStatus_ORDER_STATUS_PENDING), now, now,
	); err != nil {
		t.Fatal(err)
	}
	otherCtx := p0AdminContext(ctx, 96902, p2PublicOtherTenantID)
	otherChain := callP2PublicChain(t, otherCtx, serviceCtx, "P2CHAIN500", expiry, option.ContractStatus_CONTRACT_STATUS_UNKNOWN)
	if len(otherChain.Data) != 1 || otherChain.Data[0].StrikePrice != "9999" ||
		otherChain.Data[0].Call == nil || otherChain.Data[0].Put == nil {
		t.Fatalf("cross-tenant chain leakage: %+v", otherChain.Data)
	}
	otherBook := callP2PublicBook(t, otherCtx, serviceCtx, otherIDs[0], 100)
	if len(otherBook.Data.Bids) != 1 || otherBook.Data.Bids[0].Price != "8888" ||
		otherBook.Data.Bids[0].Qty != "8" {
		t.Fatalf("cross-tenant book mismatch: %+v", otherBook.Data)
	}
	mainWrong, mainWrongErr := applogic.NewGetOrderBookLogic(
		p0AdminContext(ctx, 96903, p0AssetE2ETenantID), serviceCtx,
	).GetOrderBook(&option.GetOrderBookReq{ContractId: otherIDs[0], DepthLimit: 100})
	if mainWrongErr != nil || mainWrong == nil || mainWrong.Base == nil || mainWrong.Base.Code == 200 || mainWrong.Data != nil {
		t.Fatalf("main tenant read other contract resp=%+v err=%v", mainWrong, mainWrongErr)
	}

	var crossFacts int64
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM t_option_market WHERE tenant_id=? AND contract_id=? AND mark_price=999) +
		(SELECT COUNT(*) FROM t_option_trade WHERE tenant_id=? AND contract_id=? AND qty=999) +
		(SELECT COUNT(*) FROM t_option_position WHERE tenant_id=? AND contract_id=? AND position_qty=999) +
		(SELECT COUNT(*) FROM t_option_order WHERE tenant_id=? AND contract_id=? AND unfilled_qty=999)`,
		p2PublicOtherTenantID, pollutedContractID,
		p2PublicOtherTenantID, pollutedContractID,
		p2PublicOtherTenantID, pollutedContractID,
		p2PublicOtherTenantID, mainBookContractID,
	).Scan(&crossFacts); err != nil {
		t.Fatal(err)
	}
	if crossFacts != 4 {
		t.Fatalf("cross-tenant polluted fact fixtures=%d want=4", crossFacts)
	}
}

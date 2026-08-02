package tasklogic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	p1RiskFaultTenantID   int64 = 996041
	p1RiskHealthyTenantID int64 = 996042
)

type riskScanFaultAssetClient struct {
	asset.AssetClient

	failTenantID int64
	failUserID   int64
}

func (c *riskScanFaultAssetClient) GetAssetBalance(
	ctx context.Context,
	in *asset.GetUserAssetDetailReq,
	opts ...grpc.CallOption,
) (*asset.GetUserAssetDetailResp, error) {
	if in.GetTenantId() == c.failTenantID && in.GetUserId() == c.failUserID {
		return nil, status.Error(codes.Unavailable, "P1 RISK-002 injected per-wallet Asset failure")
	}
	return c.AssetClient.GetAssetBalance(ctx, in, opts...)
}

func testP1RiskScanFailureIsolation(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		staleUserID       int64 = 9101
		assetFailUserID   int64 = 9102
		healthyUserID     int64 = 9103
		otherTenantUserID int64 = 9201
	)
	now := time.Now().Unix()
	seedOpenTradingCalendarForTenant(
		t, ctx, db, p1RiskFaultTenantID, logichelpers.DefaultTradingCalendarCode, now,
	)
	seedOpenTradingCalendarForTenant(
		t, ctx, db, p1RiskHealthyTenantID, logichelpers.DefaultTradingCalendarCode, now,
	)
	seeds := []struct {
		tenantID   int64
		contractID int64
		marketID   int64
		positionID int64
		userID     int64
		fresh      bool
	}{
		{p1RiskFaultTenantID, 996411, 996421, 996431, staleUserID, false},
		{p1RiskFaultTenantID, 996412, 996422, 996432, assetFailUserID, true},
		{p1RiskFaultTenantID, 996413, 996423, 996433, healthyUserID, true},
		{p1RiskHealthyTenantID, 996414, 996424, 996434, otherTenantUserID, true},
	}
	for _, seed := range seeds {
		seedP1RiskScanContractAndPosition(t, ctx, db, seed.tenantID, seed.contractID,
			seed.marketID, seed.positionID, seed.userID, seed.fresh, now)
	}
	creditP1RiskAsset(t, ctx, assetClient, p1RiskFaultTenantID, healthyUserID,
		"P1-RISK-HEALTHY-WALLET")
	creditP1RiskAsset(t, ctx, assetClient, p1RiskHealthyTenantID, otherTenantUserID,
		"P1-RISK-OTHER-TENANT")

	faultClient := &riskScanFaultAssetClient{
		AssetClient:  assetClient,
		failTenantID: p1RiskFaultTenantID, failUserID: assetFailUserID,
	}
	faultServiceCtx := *serviceCtx
	faultServiceCtx.AssetClient = faultClient
	resp, err := NewProcessRiskAccountsLogic(ctx, &faultServiceCtx).ProcessRiskAccounts(
		&option.OptionTaskReq{TenantId: 0},
	)
	if err == nil || resp != nil {
		t.Fatalf("global risk scan did not aggregate wallet failures: resp=%+v err=%v", resp, err)
	}
	if !containsAll(err.Error(), "stale option risk market", "injected per-wallet Asset failure") {
		t.Fatalf("global risk scan aggregate error lost a wallet failure: %v", err)
	}

	assertP1RiskAccountState(t, ctx, db, p1RiskFaultTenantID, staleUserID,
		option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED, false)
	assertP1RiskAccountState(t, ctx, db, p1RiskFaultTenantID, assetFailUserID,
		option.RiskAccountStatus_RISK_ACCOUNT_STATUS_RESTRICTED, false)
	assertP1RiskAccountState(t, ctx, db, p1RiskFaultTenantID, healthyUserID,
		option.RiskAccountStatus_RISK_ACCOUNT_STATUS_NORMAL, true)
	assertP1RiskAccountState(t, ctx, db, p1RiskHealthyTenantID, otherTenantUserID,
		option.RiskAccountStatus_RISK_ACCOUNT_STATUS_NORMAL, true)
}

func seedP1RiskScanContractAndPosition(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	tenantID, contractID, marketID, positionID, userID int64,
	fresh bool,
	now int64,
) {
	t.Helper()
	contractCode := fmt.Sprintf("P1-RISK-%d", contractID)
	if _, err := db.ExecContext(ctx, `
		INSERT INTO t_option_contract (
			id,tenant_id,contract_code,underlying_symbol,underlying_coin,settle_coin,quote_coin,
			option_type,exercise_style,settlement_type,strike_price,contract_unit,min_order_qty,
			max_order_qty,price_tick,qty_step,multiplier,list_time,exercise_cutoff_time,expire_time,
			deliver_time,max_user_long_qty,max_user_short_qty,max_open_interest,order_price_band_ratio,
			circuit_breaker_ratio,greeks_max_age_seconds,seller_margin_mode,initial_margin_rate,
			maintenance_margin_rate,min_margin_rate,status,is_deleted,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		contractID, tenantID, contractCode, "BTCUSDT", "BTC", "USDT", "USDT",
		int64(option.OptionType_OPTION_TYPE_CALL), int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		int64(option.SettlementType_SETTLEMENT_TYPE_CASH), "100", "1", "1", "1000", "0.1", "1", "1",
		now-3600, now+3600, now+7200, now+7200, "10000", "10000", "10000", "0.2", "0.5", 60,
		int64(option.SellerMarginMode_SELLER_MARGIN_MODE_ISOLATED), "0.2", "0.1", "0.05",
		int64(option.ContractStatus_CONTRACT_STATUS_TRADING), int64(common.YesNo_YES_NO_NO), now, now,
	); err != nil {
		t.Fatalf("seed P1 risk contract %d: %v", contractID, err)
	}
	snapshotTime := now
	if !fresh {
		snapshotTime = now - 120
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO t_option_market (
			id,tenant_id,contract_id,underlying_price,mark_price,last_price,bid_price,ask_price,
			theoretical_price,snapshot_time,underlying_snapshot_time,mark_snapshot_time,
			greeks_snapshot_time,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		marketID, tenantID, contractID, "100", "10", "10", "9.9", "10.1", "10",
		snapshotTime, snapshotTime, snapshotTime, snapshotTime, now, now,
	); err != nil {
		t.Fatalf("seed P1 risk market %d: %v", marketID, err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO t_option_position (
			id,tenant_id,user_id,account_id,contract_id,underlying_symbol,side,position_qty,
			available_qty,open_avg_price,margin_amount,exerciseable_qty,status,create_times,update_times
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		positionID, tenantID, userID, userID+10000, contractID, "BTCUSDT",
		int64(common.PositionSide_POSITION_SIDE_LONG), "1", "1", "10", "0", "1",
		int64(option.PositionStatus_POSITION_STATUS_HOLDING), now, now,
	); err != nil {
		t.Fatalf("seed P1 risk position %d: %v", positionID, err)
	}
}

func creditP1RiskAsset(
	t *testing.T,
	ctx context.Context,
	client asset.AssetClient,
	tenantID, userID int64,
	bizNo string,
) {
	t.Helper()
	resp, err := client.AddAvailable(ctx, &asset.AddAvailableReq{
		TenantId: tenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: "100",
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizNo: bizNo, Remark: "P1 risk scan isolation acceptance seed",
	})
	assertAssetOK(t, resp, err)
}

func assertP1RiskAccountState(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	tenantID, userID int64,
	wantStatus option.RiskAccountStatus,
	wantCalculated bool,
) {
	t.Helper()
	var statusValue, lastCalcTime int64
	if err := db.QueryRowContext(ctx, `SELECT status,last_calc_time FROM t_option_risk_account
		WHERE tenant_id=? AND user_id=? AND account_id=0 AND settle_coin='USDT'`,
		tenantID, userID,
	).Scan(&statusValue, &lastCalcTime); err != nil {
		t.Fatalf("load P1 risk account tenant/user=%d/%d: %v", tenantID, userID, err)
	}
	if statusValue != int64(wantStatus) || (lastCalcTime > 0) != wantCalculated {
		t.Fatalf("P1 risk account tenant/user=%d/%d status/lastCalc=%d/%d want=%d/calculated=%t",
			tenantID, userID, statusValue, lastCalcTime, wantStatus, wantCalculated)
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}

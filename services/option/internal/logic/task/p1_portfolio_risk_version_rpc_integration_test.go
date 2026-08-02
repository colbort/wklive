package tasklogic

import (
	"context"
	"database/sql"
	"strconv"
	"testing"
	"time"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	adminlogic "wklive/services/option/internal/logic/admin"
	applogic "wklive/services/option/internal/logic/app"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"google.golang.org/grpc/metadata"
)

const p1PortfolioRiskTenantID int64 = 996051

func testP1PortfolioRiskVersionGovernance(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
) {
	t.Helper()
	const (
		creatorID  int64 = 95101
		reviewerID int64 = 95102
	)
	now := time.Now().Unix()
	calendarCode := "P1_PORTFOLIO_VERSION_24_7"
	seedP1PortfolioCalendar(t, ctx, db, calendarCode, now)
	contract := insertP1PortfolioContract(t, ctx, serviceCtx, calendarCode, now)
	insertP1PortfolioMarket(t, ctx, serviceCtx, contract.Id, now)

	v1 := createP1PortfolioConfig(t, ctx, serviceCtx, creatorID, &option.CreatePortfolioRiskConfigReq{
		TenantId: p1PortfolioRiskTenantID, SettleCoin: "USDT",
		ModelMethod:      option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1,
		InitialShockRate: "0.2", MaintenanceShockRate: "0.1",
		ScenarioShocks: "-1,-0.2,0,0.2,4", ConcentrationThreshold: "100000",
		ConcentrationAddonRate: "0.1", LiquidityAddonRate: "0.02",
		EffectiveFrom: time.Now().Unix() + 2,
		ChangeReason:  "P1-004 initial governed version", EvidenceRef: "P1-004-V1-EVIDENCE",
	})
	if v1.Version != 1 || v1.SourceConfigId != 0 {
		t.Fatalf("unexpected V1 lineage: %+v", v1)
	}
	reviewP1PortfolioConfig(t, ctx, serviceCtx, reviewerID, v1.Id, true, "independent V1 approval")
	waitP1PortfolioBoundary(t, v1.EffectiveFrom)
	assertP1PortfolioPhase(t, ctx, db, assetClient, serviceCtx, contract, v1, 95111, 95211, "V1")

	v2 := createP1PortfolioConfig(t, ctx, serviceCtx, creatorID, &option.CreatePortfolioRiskConfigReq{
		TenantId: p1PortfolioRiskTenantID, SettleCoin: "USDT",
		ModelMethod:      option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1,
		InitialShockRate: "0.3", MaintenanceShockRate: "0.15",
		ScenarioShocks: "-1,-0.3,0,0.3,4", ConcentrationThreshold: "200000",
		ConcentrationAddonRate: "0.12", LiquidityAddonRate: "0.03",
		EffectiveFrom: time.Now().Unix() + 2,
		ChangeReason:  "P1-004 parameter increase", EvidenceRef: "P1-004-V2-EVIDENCE",
	})
	reviewP1PortfolioConfig(t, ctx, serviceCtx, reviewerID, v2.Id, true, "independent V2 approval")
	assertP1PortfolioActive(t, ctx, serviceCtx, v1.Id, v1.Version, time.Now().Unix())
	assertP1PortfolioClosedAt(t, ctx, db, v1.Id, v2.EffectiveFrom, v2.Id)
	waitP1PortfolioBoundary(t, v2.EffectiveFrom)
	assertP1PortfolioPhase(t, ctx, db, assetClient, serviceCtx, contract, v2, 95112, 95212, "V2")

	v3 := createP1PortfolioConfig(t, ctx, serviceCtx, creatorID, &option.CreatePortfolioRiskConfigReq{
		TenantId: p1PortfolioRiskTenantID, SettleCoin: "USDT",
		SourceConfigId: v1.Id, EffectiveFrom: time.Now().Unix() + 2,
		ChangeReason: "P1-004 rollback to validated V1", EvidenceRef: "P1-004-V3-ROLLBACK",
	})
	if v3.Version != 3 || v3.SourceConfigId != v1.Id || v3.InitialShockRate != v1.InitialShockRate ||
		v3.MaintenanceShockRate != v1.MaintenanceShockRate || v3.ScenarioShocks != v1.ScenarioShocks ||
		v3.ConcentrationThreshold != v1.ConcentrationThreshold ||
		v3.ConcentrationAddonRate != v1.ConcentrationAddonRate ||
		v3.LiquidityAddonRate != v1.LiquidityAddonRate {
		t.Fatalf("rollback V3 did not preserve V1 source lineage and parameters: V1=%+v V3=%+v", v1, v3)
	}
	reviewP1PortfolioConfig(t, ctx, serviceCtx, reviewerID, v3.Id, true, "independent rollback approval")
	assertP1PortfolioActive(t, ctx, serviceCtx, v2.Id, v2.Version, time.Now().Unix())
	assertP1PortfolioClosedAt(t, ctx, db, v2.Id, v3.EffectiveFrom, v3.Id)
	waitP1PortfolioBoundary(t, v3.EffectiveFrom)
	assertP1PortfolioPhase(t, ctx, db, assetClient, serviceCtx, contract, v3, 95113, 95213, "V3")

	retroactive := createP1PortfolioConfigResponse(t, ctx, serviceCtx, creatorID,
		&option.CreatePortfolioRiskConfigReq{
			TenantId: p1PortfolioRiskTenantID, SettleCoin: "USDT",
			ModelMethod:      option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1,
			InitialShockRate: "0.25", MaintenanceShockRate: "0.12",
			ScenarioShocks: "-1,-0.25,0,0.25,4", ConcentrationThreshold: "100000",
			ConcentrationAddonRate: "0.1", LiquidityAddonRate: "0.02",
			EffectiveFrom: time.Now().Unix(), ChangeReason: "must reject retroactive draft",
			EvidenceRef: "P1-004-RETROACTIVE-REJECT",
		})
	if retroactive.Base == nil || retroactive.Base.Code == 200 || retroactive.Data != nil {
		t.Fatalf("retroactive portfolio draft unexpectedly accepted: %+v", retroactive)
	}

	v4 := createP1PortfolioConfig(t, ctx, serviceCtx, creatorID, &option.CreatePortfolioRiskConfigReq{
		TenantId: p1PortfolioRiskTenantID, SettleCoin: "USDT",
		ModelMethod:      option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1,
		InitialShockRate: "0.25", MaintenanceShockRate: "0.12",
		ScenarioShocks: "-1,-0.25,0,0.25,4", ConcentrationThreshold: "100000",
		ConcentrationAddonRate: "0.1", LiquidityAddonRate: "0.02",
		EffectiveFrom: time.Now().Unix() + 2, ChangeReason: "approval deadline acceptance",
		EvidenceRef: "P1-004-EXPIRED-DRAFT",
	})
	waitP1PortfolioBoundary(t, v4.EffectiveFrom)
	expiredApproval := reviewP1PortfolioConfigResponse(
		t, ctx, serviceCtx, reviewerID, v4.Id, true, "must reject approval after boundary",
	)
	if expiredApproval.Base == nil || expiredApproval.Base.Code == 200 || expiredApproval.Data != nil {
		t.Fatalf("expired portfolio approval unexpectedly accepted: %+v", expiredApproval)
	}
	assertP1PortfolioDatabaseGuards(t, ctx, db, v1, v3, v4)
	reviewP1PortfolioConfig(t, ctx, serviceCtx, reviewerID, v4.Id, false, "reject expired draft")

	assertP1PortfolioActive(t, ctx, serviceCtx, v3.Id, v3.Version, time.Now().Unix())
	var configs, approved, superseded, rejected, rollbackSources, orders, riskOnV3 int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status=2),SUM(status=4),SUM(status=3),
		SUM(source_config_id=?) FROM t_option_portfolio_risk_config WHERE tenant_id=?`,
		v1.Id, p1PortfolioRiskTenantID).Scan(
		&configs, &approved, &superseded, &rejected, &rollbackSources,
	); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_order
		WHERE tenant_id=? AND portfolio_risk_config_version IN (1,2,3)`, p1PortfolioRiskTenantID).
		Scan(&orders); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM t_option_risk_account
		WHERE tenant_id=? AND portfolio_risk_config_id=? AND portfolio_risk_config_version=3`,
		p1PortfolioRiskTenantID, v3.Id).Scan(&riskOnV3); err != nil {
		t.Fatal(err)
	}
	if configs != 4 || approved != 1 || superseded != 2 || rejected != 1 ||
		rollbackSources != 1 || orders != 3 || riskOnV3 != 3 {
		t.Fatalf("P1-004 evidence configs/approved/superseded/rejected/source/orders/risk=%d/%d/%d/%d/%d/%d/%d",
			configs, approved, superseded, rejected, rollbackSources, orders, riskOnV3)
	}
	t.Logf("P1-004 versions=4 approved=1 superseded=2 rejected=1 rollback_source=%d orders=3 final_risk_v3=3",
		v1.Id)
}

func createP1PortfolioConfig(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	operatorID int64,
	req *option.CreatePortfolioRiskConfigReq,
) *option.OptionPortfolioRiskConfig {
	t.Helper()
	resp := createP1PortfolioConfigResponse(t, ctx, serviceCtx, operatorID, req)
	if resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil {
		t.Fatalf("create portfolio config rejected: %+v", resp)
	}
	return resp.Data
}

func createP1PortfolioConfigResponse(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	operatorID int64,
	req *option.CreatePortfolioRiskConfigReq,
) *option.GetPortfolioRiskConfigResp {
	t.Helper()
	resp, err := adminlogic.NewCreatePortfolioRiskConfigLogic(
		p1PortfolioAdminContext(ctx, operatorID), serviceCtx,
	).CreatePortfolioRiskConfig(req)
	if err != nil {
		t.Fatalf("create portfolio config: %v", err)
	}
	if resp == nil {
		t.Fatal("nil create portfolio config response")
	}
	return resp
}

func reviewP1PortfolioConfig(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	operatorID, configID int64,
	approve bool,
	reason string,
) *option.OptionPortfolioRiskConfig {
	t.Helper()
	resp := reviewP1PortfolioConfigResponse(t, ctx, serviceCtx, operatorID, configID, approve, reason)
	if resp.Base == nil || resp.Base.Code != 200 || resp.Data == nil {
		t.Fatalf("review portfolio config rejected: %+v", resp)
	}
	return resp.Data
}

func reviewP1PortfolioConfigResponse(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	operatorID, configID int64,
	approve bool,
	reason string,
) *option.GetPortfolioRiskConfigResp {
	t.Helper()
	resp, err := adminlogic.NewReviewPortfolioRiskConfigLogic(
		p1PortfolioAdminContext(ctx, operatorID), serviceCtx,
	).ReviewPortfolioRiskConfig(&option.ReviewPortfolioRiskConfigReq{
		TenantId: p1PortfolioRiskTenantID, ConfigId: configID, Approve: approve, Reason: reason,
	})
	if err != nil {
		t.Fatalf("review portfolio config: %v", err)
	}
	if resp == nil {
		t.Fatal("nil review portfolio config response")
	}
	return resp
}

func assertP1PortfolioPhase(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	assetClient asset.AssetClient,
	serviceCtx *svc.ServiceContext,
	contract *models.TOptionContract,
	config *option.OptionPortfolioRiskConfig,
	userID, accountID int64,
	phase string,
) {
	t.Helper()
	creditP1PortfolioAsset(t, ctx, assetClient, userID, "1000", "P1-004-"+phase+"-ASSET")
	resp, err := applogic.NewPlaceOrderLogic(
		p1PortfolioUserContext(ctx, userID), serviceCtx,
	).PlaceOrder(&option.PlaceOrderReq{
		AccountId: accountID, ContractId: contract.Id,
		Side: common.Side_SIDE_SELL, PositionEffect: option.PositionEffect_POSITION_EFFECT_OPEN,
		OrderType: option.OrderType_ORDER_TYPE_LIMIT, Price: "10", Qty: "1",
		ClientOrderId: "P1-004-" + phase + "-ORDER",
	})
	if err != nil || resp == nil || resp.Base == nil || resp.Base.Code != 200 ||
		resp.Data == nil || resp.Data.OrderId <= 0 {
		t.Fatalf("%s portfolio order response=%+v err=%v", phase, resp, err)
	}
	order, err := serviceCtx.OptionOrderModel.FindOne(ctx, resp.Data.OrderId)
	if err != nil {
		t.Fatal(err)
	}
	if order.PortfolioRiskConfigId != config.Id || order.PortfolioRiskConfigVersion != config.Version {
		t.Fatalf("%s order config=%d/%d want=%d/%d", phase,
			order.PortfolioRiskConfigId, order.PortfolioRiskConfigVersion, config.Id, config.Version)
	}
	position := &models.TOptionPosition{
		TenantId: p1PortfolioRiskTenantID, UserId: userID, AccountId: accountID,
		ContractId: contract.Id, UnderlyingSymbol: contract.UnderlyingSymbol,
		Side: int64(common.PositionSide_POSITION_SIDE_SHORT), PositionQty: decimal.NewFromInt(1),
		AvailableQty: decimal.NewFromInt(1), OpenAvgPrice: decimal.NewFromInt(10),
		Status:      int64(option.PositionStatus_POSITION_STATUS_HOLDING),
		CreateTimes: time.Now().Unix(), UpdateTimes: time.Now().Unix(),
	}
	if _, err := serviceCtx.OptionPositionModel.Insert(ctx, position); err != nil {
		t.Fatalf("insert %s risk position: %v", phase, err)
	}
	result, err := NewProcessRiskAccountsLogic(ctx, serviceCtx).ProcessRiskAccounts(&option.OptionTaskReq{
		TenantId: p1PortfolioRiskTenantID,
	})
	if err != nil || result == nil || result.Base == nil || result.Base.Code != 200 {
		t.Fatalf("%s risk scan response=%+v err=%v", phase, result, err)
	}
	var riskConfigID, riskConfigVersion int64
	if err := db.QueryRowContext(ctx, `SELECT portfolio_risk_config_id,portfolio_risk_config_version
		FROM t_option_risk_account WHERE tenant_id=? AND user_id=? AND account_id=0 AND settle_coin='USDT'`,
		p1PortfolioRiskTenantID, userID).Scan(&riskConfigID, &riskConfigVersion); err != nil {
		t.Fatal(err)
	}
	if riskConfigID != config.Id || riskConfigVersion != config.Version {
		t.Fatalf("%s risk config=%d/%d want=%d/%d", phase,
			riskConfigID, riskConfigVersion, config.Id, config.Version)
	}
	t.Logf("P1-004 %s order=%d risk=%d config=%d/%d", phase, order.Id, userID, config.Id, config.Version)
}

func assertP1PortfolioActive(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	configID, version, at int64,
) {
	t.Helper()
	active, err := serviceCtx.OptionPortfolioRiskConfigModel.FindActive(
		ctx, p1PortfolioRiskTenantID, "USDT", at,
	)
	if err != nil {
		t.Fatalf("find active portfolio config at %d: %v", at, err)
	}
	if active.Id != configID || active.Version != version {
		t.Fatalf("active portfolio config at %d=%d/%d want=%d/%d", at,
			active.Id, active.Version, configID, version)
	}
}

func assertP1PortfolioClosedAt(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	previousID, effectiveUntil, supersedingID int64,
) {
	t.Helper()
	var status, storedUntil, supersedesID int64
	if err := db.QueryRowContext(ctx, `SELECT status,effective_until,
		(SELECT supersedes_id FROM t_option_portfolio_risk_config WHERE id=?)
		FROM t_option_portfolio_risk_config WHERE id=?`, supersedingID, previousID).
		Scan(&status, &storedUntil, &supersedesID); err != nil {
		t.Fatal(err)
	}
	if status != int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED) ||
		storedUntil != effectiveUntil || supersedesID != previousID {
		t.Fatalf("portfolio interval predecessor=%d status/until/supersedes=%d/%d/%d want=4/%d/%d",
			previousID, status, storedUntil, supersedesID, effectiveUntil, previousID)
	}
}

func assertP1PortfolioDatabaseGuards(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	v1, v3, expired *option.OptionPortfolioRiskConfig,
) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `UPDATE t_option_portfolio_risk_config
		SET source_config_id=0 WHERE id=?`, v3.Id); err == nil {
		t.Fatal("database allowed immutable rollback source to be changed")
	}
	if _, err := db.ExecContext(ctx, `UPDATE t_option_portfolio_risk_config
		SET status=2,reviewed_by=95102,review_reason='late direct approval',reviewed_at=?,update_times=?
		WHERE id=?`, time.Now().Unix(), time.Now().Unix(), expired.Id); err == nil {
		t.Fatal("database allowed portfolio approval at or after effective boundary")
	}
	now := time.Now().Unix()
	if _, err := db.ExecContext(ctx, `INSERT INTO t_option_portfolio_risk_config
		(tenant_id,settle_coin,version,status,model_method,initial_shock_rate,maintenance_shock_rate,
		 scenario_shocks,concentration_threshold,concentration_addon_rate,liquidity_addon_rate,
		 effective_from,effective_until,supersedes_id,source_config_id,change_reason,evidence_ref,
		 created_by,reviewed_by,review_reason,reviewed_at,create_times,update_times)
		VALUES (?,?,5,1,1,0.21,0.1,'-1,-0.2,0,0.2,4',100000,0.1,0.02,?,0,0,?,
		'wrong source parameters','P1-004-DB-GUARD',95101,0,'',0,?,?)`,
		p1PortfolioRiskTenantID, "USDT", now+300, v1.Id, now, now); err == nil {
		t.Fatal("database guard fixture unexpectedly inserted; expected mismatch parameters")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO t_option_portfolio_risk_config
		(tenant_id,settle_coin,version,status,model_method,initial_shock_rate,maintenance_shock_rate,
		 scenario_shocks,concentration_threshold,concentration_addon_rate,liquidity_addon_rate,
		 effective_from,effective_until,supersedes_id,source_config_id,change_reason,evidence_ref,
		 created_by,reviewed_by,review_reason,reviewed_at,create_times,update_times)
		VALUES (?,?,5,1,1,0.25,0.12,'-1,-0.25,0,0.25,4',100000,0.1,0.02,?,0,0,0,
		'retroactive direct draft','P1-004-DB-GUARD',95101,0,'',0,?,?)`,
		p1PortfolioRiskTenantID, "USDT", now, now, now); err == nil {
		t.Fatal("database allowed a non-future portfolio draft")
	}
}

func waitP1PortfolioBoundary(t *testing.T, unixSecond int64) {
	t.Helper()
	delay := time.Until(time.Unix(unixSecond, 0).Add(100 * time.Millisecond))
	if delay > 0 {
		time.Sleep(delay)
	}
}

func p1PortfolioAdminContext(ctx context.Context, operatorID int64) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUid, strconv.FormatInt(operatorID, 10),
		utils.CtxKeyTenantId, strconv.FormatInt(p1PortfolioRiskTenantID, 10),
		utils.CtxKeyUserType, strconv.FormatInt(utils.SysUserTypeTenantAdmin, 10),
	))
}

func p1PortfolioUserContext(ctx context.Context, userID int64) context.Context {
	return metadata.NewIncomingContext(ctx, metadata.Pairs(
		utils.CtxKeyUid, strconv.FormatInt(userID, 10),
		utils.CtxKeyTenantId, strconv.FormatInt(p1PortfolioRiskTenantID, 10),
	))
}

func seedP1PortfolioCalendar(
	t *testing.T,
	ctx context.Context,
	db *sql.DB,
	calendarCode string,
	now int64,
) {
	t.Helper()
	seedOpenTradingCalendarForTenant(
		t, ctx, db, p1PortfolioRiskTenantID, calendarCode, now,
	)
}

func insertP1PortfolioContract(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	calendarCode string,
	now int64,
) *models.TOptionContract {
	t.Helper()
	contract := &models.TOptionContract{
		TenantId: p1PortfolioRiskTenantID, ContractCode: "P1-004-PORTFOLIO-CALL",
		UnderlyingSymbol: "BTCUSDT", UnderlyingCoin: "BTC", SettleCoin: "USDT", QuoteCoin: "USDT",
		OptionType:     int64(option.OptionType_OPTION_TYPE_CALL),
		ExerciseStyle:  int64(option.ExerciseStyle_EXERCISE_STYLE_EUROPEAN),
		SettlementType: int64(option.SettlementType_SETTLEMENT_TYPE_CASH),
		StrikePrice:    decimal.NewFromInt(100), ContractUnit: decimal.NewFromInt(1),
		MinOrderQty: decimal.NewFromInt(1), MaxOrderQty: decimal.NewFromInt(1000),
		PriceTick: decimal.RequireFromString("0.1"), QtyStep: decimal.NewFromInt(1),
		Multiplier: decimal.NewFromInt(1), ListTime: now - 3600,
		ExerciseCutoffTime: now + 3600, ExpireTime: now + 7200, DeliverTime: now + 7200,
		AutoExerciseThreshold: decimal.NewFromInt(1), MaxUserLongQty: decimal.NewFromInt(10000),
		MaxUserShortQty: decimal.NewFromInt(10000), MaxOpenInterest: decimal.NewFromInt(10000),
		OrderPriceBandRatio: decimal.RequireFromString("0.2"),
		CircuitBreakerRatio: decimal.RequireFromString("0.5"), GreeksMaxAgeSeconds: 60,
		SettlementPriceSource: "authoritative-market", SettlementPriceMethod: "MEDIAN",
		SettlementWindowSeconds: 60, SettlementMinSamples: 3,
		IsAutoExercise: int64(common.YesNo_YES_NO_NO),
		MakerFeeRate:   decimal.RequireFromString("0.02"), TakerFeeRate: decimal.RequireFromString("0.04"),
		ExerciseFeeRate: decimal.RequireFromString("0.1"), FeeUserId: 95190, FeeAccountId: 95191,
		SellerMarginMode:      int64(option.SellerMarginMode_SELLER_MARGIN_MODE_PORTFOLIO),
		InitialMarginRate:     decimal.RequireFromString("0.5"),
		MaintenanceMarginRate: decimal.RequireFromString("0.2"), MinMarginRate: decimal.RequireFromString("0.1"),
		LiquidationFeeRate: decimal.RequireFromString("0.1"), InsuranceUserId: 95192, InsuranceAccountId: 95193,
		LiquidationDeficitPolicy: int64(option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW),
		TradingCalendarCode:      calendarCode, Status: int64(option.ContractStatus_CONTRACT_STATUS_TRADING),
		IsDeleted: int64(common.YesNo_YES_NO_NO), CreateTimes: now, UpdateTimes: now,
	}
	result, err := serviceCtx.OptionContractModel.Insert(ctx, contract)
	if err != nil {
		t.Fatalf("insert P1-004 contract: %v", err)
	}
	contract.Id, err = result.LastInsertId()
	if err != nil {
		t.Fatal(err)
	}
	return contract
}

func insertP1PortfolioMarket(
	t *testing.T,
	ctx context.Context,
	serviceCtx *svc.ServiceContext,
	contractID, now int64,
) {
	t.Helper()
	_, err := serviceCtx.OptionMarketModel.Insert(ctx, &models.TOptionMarket{
		TenantId: p1PortfolioRiskTenantID, ContractId: contractID,
		UnderlyingPrice: decimal.NewFromInt(100), MarkPrice: decimal.NewFromInt(10),
		LastPrice: decimal.NewFromInt(10), BidPrice: decimal.NewFromInt(10), AskPrice: decimal.NewFromInt(10),
		TheoreticalPrice: decimal.NewFromInt(10), IntrinsicValue: decimal.NewFromInt(10),
		Iv: decimal.RequireFromString("0.5"), SnapshotTime: now,
		UnderlyingSnapshotTime: now, MarkSnapshotTime: now, GreeksSnapshotTime: now,
		CreateTimes: now, UpdateTimes: now,
	})
	if err != nil {
		t.Fatalf("insert P1-004 market: %v", err)
	}
}

func creditP1PortfolioAsset(
	t *testing.T,
	ctx context.Context,
	client asset.AssetClient,
	userID int64,
	amount, bizNo string,
) {
	t.Helper()
	resp, err := client.AddAvailable(ctx, &asset.AddAvailableReq{
		TenantId: p1PortfolioRiskTenantID, UserId: userID,
		WalletType: common.WalletType_WALLET_TYPE_OPTION, Coin: "USDT", Amount: amount,
		BizType: asset.BizType_BIZ_TYPE_OPTION, SceneType: asset.SceneType_SCENE_TYPE_TRADE_MATCH,
		BizNo: bizNo, Remark: "P1-004 portfolio version acceptance seed",
	})
	assertAssetOK(t, resp, err)
}

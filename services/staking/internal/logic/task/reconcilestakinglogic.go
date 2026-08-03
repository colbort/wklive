package tasklogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	stakeReconciliationStatusMatched = int64(1)
	stakeReconciliationStatusDiff    = int64(2)
)

type ReconcileStakingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type reconciliationKey struct {
	TenantId   int64  `db:"tenant_id"`
	CoinSymbol string `db:"coin_symbol"`
}

type reconciliationAmounts struct {
	ActivePrincipal      decimal.Decimal `db:"active_principal"`
	ProductStaked        decimal.Decimal `db:"product_staked"`
	PositionStaked       decimal.Decimal `db:"position_staked"`
	AssetLocked          decimal.Decimal `db:"asset_locked"`
	RewardLogAmount      decimal.Decimal `db:"reward_log_amount"`
	RewardPlatformAmount decimal.Decimal `db:"reward_platform_amount"`
	FeeLogAmount         decimal.Decimal `db:"fee_log_amount"`
	FeePlatformAmount    decimal.Decimal `db:"fee_platform_amount"`
}

func NewReconcileStakingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReconcileStakingLogic {
	return &ReconcileStakingLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ReconcileStaking records one cumulative accounting snapshot per tenant and
// coin. Cumulative business logs are compared with cumulative Asset platform
// flows, while all current principal aggregates are compared at the same read
// point. A difference remains visible until a later run proves it is closed.
func (l *ReconcileStakingLogic) ReconcileStaking(in *staking.StakingTaskReq) (*staking.StakingTaskResp, error) {
	lockName := fmt.Sprintf("reconcile_staking:%d", in.GetTenantId())
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, lockName, func() (*staking.StakingTaskResp, error) {
		keys, err := l.findKeys(in.GetTenantId())
		if err != nil {
			return nil, err
		}
		now := utils.NowMillis()
		date := utcDate(now)
		for _, key := range keys {
			amounts, err := l.aggregate(key)
			if err != nil {
				return nil, fmt.Errorf("aggregate staking reconciliation tenant=%d coin=%s: %w", key.TenantId, key.CoinSymbol, err)
			}
			record := buildReconciliationRecord(key, date, now, amounts)
			if err := l.svcCtx.StakeReconciliationModel.Upsert(l.ctx, record); err != nil {
				return nil, fmt.Errorf("save staking reconciliation tenant=%d coin=%s: %w", key.TenantId, key.CoinSymbol, err)
			}
			if record.Status == stakeReconciliationStatusDiff {
				l.Errorf("staking reconciliation difference tenant=%d coin=%s detail=%s", key.TenantId, key.CoinSymbol, record.Detail)
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ReconcileStakingLogic) findKeys(tenantID int64) ([]reconciliationKey, error) {
	query := `SELECT DISTINCT tenant_id, UPPER(coin_symbol) AS coin_symbol FROM (
		SELECT tenant_id,coin_symbol FROM t_stake_product
		UNION SELECT tenant_id,coin_symbol FROM t_stake_order
		UNION SELECT tenant_id,reward_coin_symbol AS coin_symbol FROM t_stake_reward_log
		UNION SELECT tenant_id,coin AS coin_symbol FROM t_asset_lock WHERE biz_type='staking'
		UNION SELECT tenant_id,coin AS coin_symbol FROM t_asset_platform_flow WHERE account_type IN ('STAKING_REWARD','FEE_REVENUE')
	) reconciliation_coins WHERE coin_symbol<>'' AND (?=0 OR tenant_id=?) ORDER BY tenant_id,coin_symbol`
	var keys []reconciliationKey
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &keys, query, tenantID, tenantID); err != nil {
		return nil, err
	}
	return keys, nil
}

func (l *ReconcileStakingLogic) aggregate(key reconciliationKey) (reconciliationAmounts, error) {
	query := `SELECT
		(SELECT COALESCE(SUM(stake_amount),0) FROM t_stake_order WHERE tenant_id=? AND UPPER(coin_symbol)=? AND status=1) AS active_principal,
		(SELECT COALESCE(SUM(staked_amount),0) FROM t_stake_product WHERE tenant_id=? AND UPPER(coin_symbol)=?) AS product_staked,
		(SELECT COALESCE(SUM(p.staked_amount),0) FROM t_stake_user_position p JOIN t_stake_product s ON s.id=p.product_id AND s.tenant_id=p.tenant_id WHERE p.tenant_id=? AND UPPER(s.coin_symbol)=?) AS position_staked,
		(SELECT COALESCE(SUM(remain_amount),0) FROM t_asset_lock WHERE tenant_id=? AND biz_type='staking' AND UPPER(coin)=?) AS asset_locked,
		(SELECT COALESCE(SUM(reward_amount),0) FROM t_stake_reward_log WHERE tenant_id=? AND UPPER(reward_coin_symbol)=? AND reward_status=2) AS reward_log_amount,
		(SELECT COALESCE(SUM(amount),0) FROM t_asset_platform_flow WHERE tenant_id=? AND UPPER(coin)=? AND account_type='STAKING_REWARD' AND op_type=2) AS reward_platform_amount,
		(SELECT COALESCE(SUM(r.fee_amount),0) FROM t_stake_redeem_log r JOIN t_stake_order o ON o.id=r.order_id AND o.tenant_id=r.tenant_id WHERE r.tenant_id=? AND UPPER(o.coin_symbol)=? AND r.redeem_status=2) AS fee_log_amount,
		(SELECT COALESCE(SUM(amount),0) FROM t_asset_platform_flow WHERE tenant_id=? AND UPPER(coin)=? AND account_type='FEE_REVENUE' AND op_type=1) AS fee_platform_amount`
	args := make([]any, 0, 16)
	for i := 0; i < 8; i++ {
		args = append(args, key.TenantId, key.CoinSymbol)
	}
	var amounts reconciliationAmounts
	if err := l.svcCtx.DB.QueryRowCtx(l.ctx, &amounts, query, args...); err != nil {
		return reconciliationAmounts{}, err
	}
	return amounts, nil
}

func buildReconciliationRecord(key reconciliationKey, date, now int64, a reconciliationAmounts) *models.TStakeReconciliation {
	productDiff := a.ProductStaked.Sub(a.ActivePrincipal)
	positionDiff := a.PositionStaked.Sub(a.ActivePrincipal)
	lockDiff := a.AssetLocked.Sub(a.ActivePrincipal)
	rewardDiff := a.RewardPlatformAmount.Sub(a.RewardLogAmount)
	feeDiff := a.FeePlatformAmount.Sub(a.FeeLogAmount)
	status := stakeReconciliationStatusMatched
	parts := make([]string, 0, 5)
	appendDiff := func(name string, value decimal.Decimal) {
		if value.IsZero() {
			return
		}
		status = stakeReconciliationStatusDiff
		parts = append(parts, name+"="+value.String())
	}
	appendDiff("product", productDiff)
	appendDiff("position", positionDiff)
	appendDiff("asset_lock", lockDiff)
	appendDiff("reward", rewardDiff)
	appendDiff("fee", feeDiff)
	return &models.TStakeReconciliation{
		TenantId: key.TenantId, ReconciliationDate: date, CoinSymbol: key.CoinSymbol,
		ActivePrincipal: a.ActivePrincipal, ProductStaked: a.ProductStaked, PositionStaked: a.PositionStaked,
		AssetLocked: a.AssetLocked, RewardLogAmount: a.RewardLogAmount,
		RewardPlatformAmount: a.RewardPlatformAmount, FeeLogAmount: a.FeeLogAmount,
		FeePlatformAmount: a.FeePlatformAmount, ProductDiff: productDiff, PositionDiff: positionDiff,
		LockDiff: lockDiff, RewardDiff: rewardDiff, FeeDiff: feeDiff, Status: status,
		Detail: strings.Join(parts, ", "), CreateTimes: now, UpdateTimes: now,
	}
}

func utcDate(nowMillis int64) int64 {
	value := time.UnixMilli(nowMillis).UTC()
	return int64(value.Year()*10000 + int(value.Month())*100 + value.Day())
}

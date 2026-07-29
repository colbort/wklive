package tasklogic

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const crossMarginRiskRateMax = "9999999999.9999999999"

type crossMarginAggregate = models.CrossMarginAggregate
type crossMarginCursor = models.CrossMarginAggregateCursor

func (l *ProcessPositionsLogic) refreshCrossMarginSnapshots(tenantID int64) error {
	cursor := crossMarginCursor{}
	for {
		groups, err := l.findCrossMarginAggregates(tenantID, cursor, 100)
		if err != nil {
			return err
		}
		if len(groups) == 0 {
			return nil
		}
		for _, group := range groups {
			cursor = crossMarginCursor{
				TenantID: group.TenantID, UserID: group.UserID,
				MarginAsset: group.MarginAsset, Valid: true,
			}
			if err = l.projectCrossMarginAccount(group); err != nil {
				return err
			}
		}
		if len(groups) < 100 {
			return nil
		}
	}
}

func (l *ProcessPositionsLogic) findCrossMarginAggregates(tenantID int64, cursor crossMarginCursor, limit int) ([]*crossMarginAggregate, error) {
	return l.svcCtx.ContractMarginSnapshotModel.FindRiskProjectionAggregates(
		l.ctx, tenantID, cursor, limit,
	)
}

func (l *ProcessPositionsLogic) projectCrossMarginAccount(group *crossMarginAggregate) error {
	if group == nil {
		return errors.New("cross margin aggregate is nil")
	}
	resp, err := l.svcCtx.AssetAdminClient.GetUserAssetDetail(l.ctx, &asset.GetUserAssetDetailReq{
		TenantId: group.TenantID, UserId: group.UserID,
		WalletType: common.WalletType_WALLET_TYPE_CONTRACT, Coin: group.MarginAsset,
	})
	if err != nil {
		return fmt.Errorf("query cross margin Asset wallet: %w", err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return errors.New("cross margin Asset wallet query rejected")
	}
	wallet := resp.GetData()
	total, err := decimal.NewFromString(wallet.GetTotalAmount())
	if err != nil {
		return fmt.Errorf("parse cross margin wallet total: %w", err)
	}
	available, err := decimal.NewFromString(wallet.GetAvailableAmount())
	if err != nil {
		return fmt.Errorf("parse cross margin wallet available: %w", err)
	}
	frozen, err := decimal.NewFromString(wallet.GetFrozenAmount())
	if err != nil {
		return fmt.Errorf("parse cross margin wallet frozen: %w", err)
	}
	equity, availableMargin, riskRate := calculateCrossAccountRisk(
		total, available, group.PositionMargin, group.UnrealizedPnl, group.MaintenanceMargin,
	)
	now := utils.NowMillis()
	snapshotTime := group.PositionUpdateTime
	if group.OrderUpdateTime > snapshotTime {
		snapshotTime = group.OrderUpdateTime
	}
	if wallet.GetUpdateTimes() > snapshotTime {
		snapshotTime = wallet.GetUpdateTimes()
	}
	sourceEventNo := crossMarginProjectionSource(
		group.TenantID, group.UserID, group.MarginAsset, wallet.GetVersion(),
		group.PositionVersionSum, group.OrderVersionSum,
	)
	_, err = l.svcCtx.ContractMarginSnapshotModel.UpsertRiskProjection(l.ctx, &models.TContractMarginSnapshot{
		TenantId:          group.TenantID,
		UserId:            group.UserID,
		MarginAsset:       group.MarginAsset,
		WalletBalance:     total,
		AvailableBalance:  available,
		FrozenBalance:     frozen,
		PositionMargin:    group.PositionMargin,
		OrderMargin:       group.OrderMargin,
		MaintenanceMargin: group.MaintenanceMargin,
		AccountEquity:     equity,
		AvailableMargin:   availableMargin,
		RiskRate:          riskRate,
		PositionCount:     group.PositionCount,
		AssetVersion:      wallet.GetVersion(),
		UnrealizedPnl:     group.UnrealizedPnl,
		RealizedPnl:       group.RealizedPnl,
		SourceEventNo:     sql.NullString{String: sourceEventNo, Valid: true},
		SnapshotTime:      snapshotTime,
		CreateTimes:       now,
		UpdateTimes:       now,
	})
	return err
}

func calculateCrossAccountRisk(walletTotal, walletAvailable, positionMargin, unrealizedPnl, maintenanceMargin decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal) {
	equity := walletTotal.Add(positionMargin).Add(unrealizedPnl)
	availableMargin := walletAvailable.Add(unrealizedPnl)
	if !maintenanceMargin.IsPositive() {
		return equity, availableMargin, decimal.Zero
	}
	if !equity.IsPositive() {
		return equity, availableMargin, decimal.RequireFromString(crossMarginRiskRateMax)
	}
	return equity, availableMargin, maintenanceMargin.Div(equity).Round(10)
}

func crossMarginProjectionSource(tenantID, userID int64, marginAsset string, assetVersion, positionVersionSum, orderVersionSum int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf(
		"%d|%d|%s|%d|%d|%d",
		tenantID, userID, marginAsset, assetVersion, positionVersionSum, orderVersionSum,
	)))
	return "CR-" + hex.EncodeToString(sum[:24])
}

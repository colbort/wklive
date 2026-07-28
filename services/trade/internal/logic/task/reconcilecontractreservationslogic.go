package tasklogic

import (
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/trade"
	"wklive/services/trade/models"

	"github.com/shopspring/decimal"
)

const (
	reservationFreezeCheck     = "RESERVATION_ASSET_FREEZE"
	reservationFreezeScanLimit = 100
)

type contractReservationAudit struct {
	Id             int64           `db:"id"`
	TenantId       int64           `db:"tenant_id"`
	OrderId        int64           `db:"order_id"`
	OrderNo        string          `db:"order_no"`
	UserId         int64           `db:"user_id"`
	ReservationNo  string          `db:"reservation_no"`
	Asset          string          `db:"asset"`
	ReservedAmount decimal.Decimal `db:"reserved_amount"`
	ConsumedAmount decimal.Decimal `db:"consumed_amount"`
	ReleasedAmount decimal.Decimal `db:"released_amount"`
	Status         int64           `db:"status"`
}

func (l *ReconcileContractAssetFlowsLogic) reconcileReservations(tenantID int64) error {
	now := utils.NowMillis()
	cursor, err := l.loadReconciliationCursor(tenantID, reservationFreezeCheck, now)
	if err != nil {
		return err
	}
	rows, err := l.findContractReservationAudits(
		tenantID, cursor, now-orderFillStableDelay, reservationFreezeScanLimit,
	)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return l.completeReconciliationCycle(tenantID, reservationFreezeCheck, now)
	}
	for _, row := range rows {
		if err = l.reconcileReservation(row, now); err != nil {
			return err
		}
	}
	return l.advanceReconciliationCursor(tenantID, reservationFreezeCheck, rows[len(rows)-1].Id, now)
}

func (l *ReconcileContractAssetFlowsLogic) findContractReservationAudits(tenantID, cursor, cutoff int64, limit int) ([]*contractReservationAudit, error) {
	tenantClause := ""
	args := []any{int64(common.ProductType_PRODUCT_TYPE_DERIVATIVE), cursor, cutoff}
	if tenantID > 0 {
		tenantClause = " AND r.tenant_id=?"
		args = append(args, tenantID)
	}
	args = append(args, limit)
	query := `
SELECT r.id,r.tenant_id,r.order_id,o.order_no,o.user_id,r.reservation_no,r.asset,
       r.reserved_amount,r.consumed_amount,r.released_amount,r.status
FROM t_trade_asset_reservation r
JOIN t_trade_order o ON o.tenant_id=r.tenant_id AND o.id=r.order_id
WHERE o.product_type=? AND r.id>? AND r.update_times<=?` + tenantClause + `
ORDER BY r.id
LIMIT ?`
	var rows []*contractReservationAudit
	if err := l.svcCtx.DB.QueryRowsCtx(l.ctx, &rows, query, args...); err != nil {
		return nil, err
	}
	return rows, nil
}

func (l *ReconcileContractAssetFlowsLogic) reconcileReservation(row *contractReservationAudit, now int64) error {
	resp, err := l.svcCtx.AssetAdminClient.PageAssetFreezes(l.ctx, &asset.PageAssetFreezesReq{
		TenantId:   row.TenantId,
		UserId:     row.UserId,
		WalletType: common.WalletType_WALLET_TYPE_CONTRACT,
		Coin:       row.Asset,
		BizType:    asset.BizType_BIZ_TYPE_TRADE,
		BizNo:      row.ReservationNo,
		Page:       &common.PageReq{Limit: 2},
	})
	if err != nil {
		return fmt.Errorf("query Asset freeze for reservation %s: %w", row.ReservationNo, err)
	}
	if resp == nil || resp.GetBase() == nil || resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("Asset freeze query rejected for reservation %s", row.ReservationNo)
	}
	var freeze *asset.AssetFreeze
	if len(resp.GetData()) == 1 {
		freeze = resp.GetData()[0]
	}
	issueKey := "RESERVATION_FREEZE:" + row.ReservationNo
	matched, detail := assetFreezeMatchesReservation(row, resp.GetData())
	if matched {
		return l.svcCtx.ContractReconcileIssueModel.ResolveByKey(
			l.ctx, row.TenantId, issueKey, "Trade reservation matches Asset freeze fact", now,
		)
	}
	expected := expectedReservationSummary(row)
	actual := actualAssetFreezeSummary(freeze, len(resp.GetData()))
	if err = l.recordContractReconciliationFinding(&models.TContractReconciliationIssue{
		TenantId:      row.TenantId,
		IssueKey:      issueKey,
		CheckType:     reservationFreezeCheck,
		BizType:       "reservation",
		BizNo:         row.ReservationNo,
		ExpectedValue: expected,
		ActualValue:   actual,
		Detail:        detail,
		FirstSeenAt:   now,
		LastSeenAt:    now,
		CreateTimes:   now,
		UpdateTimes:   now,
	}); err != nil {
		return err
	}
	return nil
}

func assetFreezeMatchesReservation(row *contractReservationAudit, freezes []*asset.AssetFreeze) (bool, string) {
	if row == nil {
		return false, "reservation is nil"
	}
	if row.Status == int64(trade.AssetReservationStatus_ASSET_RESERVATION_STATUS_FAILED) && len(freezes) == 0 {
		return true, ""
	}
	if len(freezes) != 1 || freezes[0] == nil {
		return false, fmt.Sprintf("expected exactly one Asset freeze, got %d", len(freezes))
	}
	freeze := freezes[0]
	if freeze.GetTenantId() != row.TenantId || freeze.GetUserId() != row.UserId ||
		freeze.GetCoin() != row.Asset || freeze.GetBizType() != asset.BizType_BIZ_TYPE_TRADE ||
		freeze.GetBizNo() != row.ReservationNo {
		return false, "Asset freeze identity does not match reservation"
	}
	amount, amountErr := decimal.NewFromString(freeze.GetAmount())
	used, usedErr := decimal.NewFromString(freeze.GetUsedAmount())
	unfrozen, unfreezeErr := decimal.NewFromString(freeze.GetUnfreezeAmount())
	remain, remainErr := decimal.NewFromString(freeze.GetRemainAmount())
	if errors.Join(amountErr, usedErr, unfreezeErr, remainErr) != nil {
		return false, "Asset freeze contains invalid decimal amount"
	}
	expectedRemain := row.ReservedAmount.Sub(row.ConsumedAmount).Sub(row.ReleasedAmount)
	if !amount.Equal(row.ReservedAmount) || !used.Equal(row.ConsumedAmount) ||
		!unfrozen.Equal(row.ReleasedAmount) || !remain.Equal(expectedRemain) {
		return false, "Asset freeze amounts do not match reservation ledger"
	}
	if !amount.Equal(used.Add(unfrozen).Add(remain)) || expectedRemain.IsNegative() {
		return false, "freeze or reservation amount conservation is violated"
	}
	expectedStatus := expectedAssetFreezeStatus(row.ConsumedAmount, row.ReleasedAmount, expectedRemain)
	if freeze.GetStatus() != expectedStatus {
		return false, fmt.Sprintf("Asset freeze status=%s expected=%s", freeze.GetStatus(), expectedStatus)
	}
	return true, ""
}

func expectedAssetFreezeStatus(consumed, released, remain decimal.Decimal) asset.FreezeStatus {
	if remain.IsPositive() {
		if consumed.IsPositive() || released.IsPositive() {
			return asset.FreezeStatus_FREEZE_STATUS_PARTIAL_RELEASED
		}
		return asset.FreezeStatus_FREEZE_STATUS_FREEZING
	}
	if consumed.IsPositive() && released.IsPositive() {
		return asset.FreezeStatus_FREEZE_STATUS_CLOSED
	}
	if consumed.IsPositive() {
		return asset.FreezeStatus_FREEZE_STATUS_DEDUCTED
	}
	return asset.FreezeStatus_FREEZE_STATUS_UNFROZEN
}

func expectedReservationSummary(row *contractReservationAudit) string {
	remain := row.ReservedAmount.Sub(row.ConsumedAmount).Sub(row.ReleasedAmount)
	return fmt.Sprintf("user=%d asset=%s total=%s consumed=%s released=%s remain=%s status=%d",
		row.UserId, row.Asset, row.ReservedAmount, row.ConsumedAmount, row.ReleasedAmount, remain, row.Status)
}

func actualAssetFreezeSummary(freeze *asset.AssetFreeze, count int) string {
	if freeze == nil {
		return fmt.Sprintf("freeze_count=%d", count)
	}
	return fmt.Sprintf("freeze_count=%d freeze=%s user=%d asset=%s total=%s used=%s released=%s remain=%s status=%s",
		count, freeze.GetFreezeNo(), freeze.GetUserId(), freeze.GetCoin(), freeze.GetAmount(),
		freeze.GetUsedAmount(), freeze.GetUnfreezeAmount(), freeze.GetRemainAmount(), freeze.GetStatus())
}

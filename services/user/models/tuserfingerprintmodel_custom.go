package models

import (
	"context"
	"database/sql"
	"fmt"
	"wklive/common/sqlutil"
)

type UserFingerprintMatchRow struct {
	Id          int64  `db:"id"`
	UserId      int64  `db:"user_id"`
	Fingerprint string `db:"fingerprint"`
	SourceIp    string `db:"source_ip"`
}

type UserFingerprintModel interface {
	tUserFingerprintModel
	UpsertSeen(ctx context.Context, data *TUserFingerprint) error
	FindGuestFingerprintCandidates(ctx context.Context, tenantId int64, matchKey string, cursor int64, limit int64) ([]*UserFingerprintMatchRow, error)
}

func (m *defaultTUserFingerprintModel) UpsertSeen(ctx context.Context, data *TUserFingerprint) error {
	query := fmt.Sprintf(`
		INSERT INTO %s (%s)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			device_id = VALUES(device_id),
			match_key = VALUES(match_key),
			fingerprint = VALUES(fingerprint),
			source_ip = VALUES(source_ip),
			last_seen_time = VALUES(last_seen_time),
			update_times = VALUES(update_times)
	`, m.table, tUserFingerprintRowsExpectAutoSet)

	_, err := m.ExecNoCacheCtx(ctx, query, data.TenantId, data.UserId, data.DeviceId, data.FingerprintHash, data.MatchKey, data.Fingerprint, data.SourceIp, data.FirstSeenTime, data.LastSeenTime, data.CreateTimes, data.UpdateTimes)
	return err
}

func (m *defaultTUserFingerprintModel) FindGuestFingerprintCandidates(ctx context.Context, tenantId int64, matchKey string, cursor int64, limit int64) ([]*UserFingerprintMatchRow, error) {
	if limit <= 0 || limit > 500 {
		limit = 500
	}

	builder := sqlutil.NewPageQueryBuilder()
	builder.EqInt64("f.tenant_id", tenantId)
	if cursor > 0 {
		builder.And("f.id < ?", cursor)
	}
	builder.EqInt64("u.is_guest", int64(2))
	builder.EqInt64("u.deleted", int64(0))
	builder.EqString("f.match_key", matchKey)

	where := builder.Where()
	args := builder.Args()
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT f.id, f.user_id, f.fingerprint
		FROM %s f
		JOIN t_user u ON u.tenant_id = f.tenant_id AND u.id = f.user_id
		WHERE %s
		ORDER BY f.id DESC
		LIMIT ?
	`, m.table, where)

	var list []*UserFingerprintMatchRow
	if err := m.QueryRowsNoCacheCtx(ctx, &list, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return list, nil
}

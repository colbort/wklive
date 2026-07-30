package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SysAdminNotificationIncidentModel = (*customSysAdminNotificationIncidentModel)(nil)

const (
	AdminNotificationStatusOpen         int64 = 1
	AdminNotificationStatusAcknowledged int64 = 2
	AdminNotificationStatusResolved     int64 = 3
)

type (
	RecordAdminNotificationIncidentInput struct {
		TenantId     int64
		EventType    string
		AlertKey     string
		EventId      string
		State        string
		Severity     string
		Title        string
		Message      string
		Source       string
		PayloadJson  string
		EventTime    int64
		AckTimeoutMs int64
	}

	AcknowledgeAdminNotificationIncidentInput struct {
		TenantId  int64
		EventType string
		AlertKey  string
		UserId    int64
		Username  string
		Reason    string
		Now       int64
	}

	// SysAdminNotificationIncidentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSysAdminNotificationIncidentModel.
	SysAdminNotificationIncidentModel interface {
		sysAdminNotificationIncidentModel
		Record(
			ctx context.Context,
			input RecordAdminNotificationIncidentInput,
		) (*SysAdminNotificationIncident, error)
		Acknowledge(
			ctx context.Context,
			input AcknowledgeAdminNotificationIncidentInput,
		) (*SysAdminNotificationIncident, error)
		ClaimDue(
			ctx context.Context,
			now int64,
			repeatIntervalMs int64,
			maxLevel int64,
			limit int64,
		) ([]*SysAdminNotificationIncident, error)
		ReleaseEscalation(
			ctx context.Context,
			id int64,
			claimedLevel int64,
			retryAt int64,
			now int64,
		) (bool, error)
	}

	customSysAdminNotificationIncidentModel struct {
		*defaultSysAdminNotificationIncidentModel
	}
)

// NewSysAdminNotificationIncidentModel returns a model for the database table.
func NewSysAdminNotificationIncidentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) SysAdminNotificationIncidentModel {
	return &customSysAdminNotificationIncidentModel{
		defaultSysAdminNotificationIncidentModel: newSysAdminNotificationIncidentModel(conn, c, opts...),
	}
}

func (m *defaultSysAdminNotificationIncidentModel) Record(
	ctx context.Context,
	input RecordAdminNotificationIncidentInput,
) (*SysAdminNotificationIncident, error) {
	var result *SysAdminNotificationIncident
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		current, err := m.findForUpdate(ctx, session, input.TenantId, input.EventType, input.AlertKey)
		if err != nil && !isAdminNotificationNotFound(err) {
			return err
		}
		if current == nil {
			status := AdminNotificationStatusOpen
			nextEscalateAt := input.EventTime + input.AckTimeoutMs
			resolvedAt := int64(0)
			if input.State == "resolved" {
				status = AdminNotificationStatusResolved
				nextEscalateAt = 0
				resolvedAt = input.EventTime
			}
			insert := fmt.Sprintf(`INSERT INTO %s
(tenant_id, event_type, alert_key, last_event_id, severity, title, message,
 source, payload_json, status, first_seen_at, last_seen_at,
 next_escalate_at, resolved_at, version, create_times, update_times)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?)`, m.table)
			insertResult, err := session.ExecCtx(
				ctx,
				insert,
				input.TenantId,
				input.EventType,
				input.AlertKey,
				input.EventId,
				input.Severity,
				input.Title,
				input.Message,
				input.Source,
				input.PayloadJson,
				status,
				input.EventTime,
				input.EventTime,
				nextEscalateAt,
				resolvedAt,
				input.EventTime,
				input.EventTime,
			)
			if err != nil {
				return err
			}
			id, err := insertResult.LastInsertId()
			if err != nil {
				return err
			}
			result, err = m.findById(ctx, session, id)
			return err
		}

		if input.State == "resolved" {
			update := fmt.Sprintf(`UPDATE %s
SET last_event_id=?, severity=?, title=?, message=?, source=?, payload_json=?,
    status=?, last_seen_at=?, next_escalate_at=0, resolved_at=?,
    version=version+1, update_times=?
WHERE id=?`, m.table)
			if _, err := session.ExecCtx(
				ctx,
				update,
				input.EventId,
				input.Severity,
				input.Title,
				input.Message,
				input.Source,
				input.PayloadJson,
				AdminNotificationStatusResolved,
				input.EventTime,
				input.EventTime,
				input.EventTime,
				current.Id,
			); err != nil {
				return err
			}
		} else if current.Status == AdminNotificationStatusResolved {
			update := fmt.Sprintf(`UPDATE %s
SET last_event_id=?, severity=?, title=?, message=?, source=?, payload_json=?,
    status=?, first_seen_at=?, last_seen_at=?, acknowledged_at=0,
    acknowledged_by=0, acknowledged_username='', acknowledge_reason='',
    escalation_level=0, last_escalated_at=0, next_escalate_at=?,
    resolved_at=0, version=version+1, update_times=?
WHERE id=?`, m.table)
			if _, err := session.ExecCtx(
				ctx,
				update,
				input.EventId,
				input.Severity,
				input.Title,
				input.Message,
				input.Source,
				input.PayloadJson,
				AdminNotificationStatusOpen,
				input.EventTime,
				input.EventTime,
				input.EventTime+input.AckTimeoutMs,
				input.EventTime,
				current.Id,
			); err != nil {
				return err
			}
		} else {
			update := fmt.Sprintf(`UPDATE %s
SET last_event_id=?, severity=?, title=?, message=?, source=?, payload_json=?,
    last_seen_at=?, version=version+1, update_times=?
WHERE id=?`, m.table)
			if _, err := session.ExecCtx(
				ctx,
				update,
				input.EventId,
				input.Severity,
				input.Title,
				input.Message,
				input.Source,
				input.PayloadJson,
				input.EventTime,
				input.EventTime,
				current.Id,
			); err != nil {
				return err
			}
		}
		result, err = m.findById(ctx, session, current.Id)
		return err
	})
	if err != nil || result == nil {
		return result, err
	}
	return result, m.clearAdminNotificationCache(ctx, result)
}

func (m *defaultSysAdminNotificationIncidentModel) Acknowledge(
	ctx context.Context,
	input AcknowledgeAdminNotificationIncidentInput,
) (*SysAdminNotificationIncident, error) {
	var result *SysAdminNotificationIncident
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		current, err := m.findForUpdate(ctx, session, input.TenantId, input.EventType, input.AlertKey)
		if err != nil {
			return err
		}
		if current.Status == AdminNotificationStatusResolved {
			return errors.New("resolved admin notification cannot be acknowledged")
		}
		if current.Status != AdminNotificationStatusAcknowledged {
			update := fmt.Sprintf(`UPDATE %s
SET status=?, acknowledged_at=?, acknowledged_by=?,
    acknowledged_username=?, acknowledge_reason=?, next_escalate_at=0,
    version=version+1, update_times=?
WHERE id=?`, m.table)
			if _, err := session.ExecCtx(
				ctx,
				update,
				AdminNotificationStatusAcknowledged,
				input.Now,
				input.UserId,
				input.Username,
				input.Reason,
				input.Now,
				current.Id,
			); err != nil {
				return err
			}
		}
		result, err = m.findById(ctx, session, current.Id)
		return err
	})
	if err != nil || result == nil {
		return result, err
	}
	return result, m.clearAdminNotificationCache(ctx, result)
}

func (m *defaultSysAdminNotificationIncidentModel) ClaimDue(
	ctx context.Context,
	now int64,
	repeatIntervalMs int64,
	maxLevel int64,
	limit int64,
) ([]*SysAdminNotificationIncident, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var claimed []*SysAdminNotificationIncident
	err := m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		query := fmt.Sprintf(`SELECT %s FROM %s
WHERE status=? AND next_escalate_at>0 AND next_escalate_at<=?
  AND escalation_level<?
ORDER BY next_escalate_at ASC, id ASC
LIMIT ? FOR UPDATE SKIP LOCKED`, sysAdminNotificationIncidentRows, m.table)
		var due []*SysAdminNotificationIncident
		if err := session.QueryRowsCtx(
			ctx,
			&due,
			query,
			AdminNotificationStatusOpen,
			now,
			maxLevel,
			limit,
		); err != nil {
			return err
		}
		for _, incident := range due {
			nextLevel := incident.EscalationLevel + 1
			nextEscalateAt := int64(0)
			if nextLevel < maxLevel {
				nextEscalateAt = now + repeatIntervalMs
			}
			update := fmt.Sprintf(`UPDATE %s
SET escalation_level=?, last_escalated_at=?, next_escalate_at=?,
    version=version+1, update_times=?
WHERE id=?`, m.table)
			if _, err := session.ExecCtx(
				ctx,
				update,
				nextLevel,
				now,
				nextEscalateAt,
				now,
				incident.Id,
			); err != nil {
				return err
			}
			incident.EscalationLevel = nextLevel
			incident.LastEscalatedAt = now
			incident.NextEscalateAt = nextEscalateAt
			incident.Version++
			incident.UpdateTimes = now
			claimed = append(claimed, incident)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, incident := range claimed {
		if err := m.clearAdminNotificationCache(ctx, incident); err != nil {
			return nil, err
		}
	}
	return claimed, nil
}

func (m *defaultSysAdminNotificationIncidentModel) ReleaseEscalation(
	ctx context.Context,
	id int64,
	claimedLevel int64,
	retryAt int64,
	now int64,
) (bool, error) {
	query := fmt.Sprintf(`UPDATE %s
SET escalation_level=GREATEST(escalation_level-1, 0),
    next_escalate_at=?, version=version+1, update_times=?
WHERE id=? AND status=? AND escalation_level=?`, m.table)
	result, err := m.ExecNoCacheCtx(
		ctx,
		query,
		retryAt,
		now,
		id,
		AdminNotificationStatusOpen,
		claimedLevel,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil || affected != 1 {
		return false, err
	}
	var row SysAdminNotificationIncident
	if err := m.QueryRowNoCacheCtx(
		ctx,
		&row,
		fmt.Sprintf("SELECT %s FROM %s WHERE id=? LIMIT 1", sysAdminNotificationIncidentRows, m.table),
		id,
	); err != nil {
		return false, err
	}
	return true, m.clearAdminNotificationCache(ctx, &row)
}

func (m *defaultSysAdminNotificationIncidentModel) findForUpdate(
	ctx context.Context,
	session sqlx.Session,
	tenantId int64,
	eventType string,
	alertKey string,
) (*SysAdminNotificationIncident, error) {
	query := fmt.Sprintf(`SELECT %s FROM %s
WHERE tenant_id=? AND event_type=? AND alert_key=?
LIMIT 1 FOR UPDATE`, sysAdminNotificationIncidentRows, m.table)
	var result SysAdminNotificationIncident
	if err := session.QueryRowCtx(ctx, &result, query, tenantId, eventType, alertKey); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *defaultSysAdminNotificationIncidentModel) findById(
	ctx context.Context,
	session sqlx.Session,
	id int64,
) (*SysAdminNotificationIncident, error) {
	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE id=? LIMIT 1",
		sysAdminNotificationIncidentRows,
		m.table,
	)
	var result SysAdminNotificationIncident
	if err := session.QueryRowCtx(ctx, &result, query, id); err != nil {
		return nil, err
	}
	return &result, nil
}

func (m *defaultSysAdminNotificationIncidentModel) clearAdminNotificationCache(
	ctx context.Context,
	value *SysAdminNotificationIncident,
) error {
	return m.DelCacheCtx(
		ctx,
		fmt.Sprintf("%s%v", cacheSysAdminNotificationIncidentIdPrefix, value.Id),
		fmt.Sprintf(
			"%s%v:%v:%v",
			cacheSysAdminNotificationIncidentTenantIdEventTypeAlertKeyPrefix,
			value.TenantId,
			value.EventType,
			value.AlertKey,
		),
	)
}

func isAdminNotificationNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || errors.Is(err, sqlc.ErrNotFound)
}

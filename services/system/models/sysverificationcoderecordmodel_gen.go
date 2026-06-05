package models

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	sysVerificationCodeRecordFieldNames          = builder.RawFieldNames(&SysVerificationCodeRecord{})
	sysVerificationCodeRecordRows                = strings.Join(sysVerificationCodeRecordFieldNames, ",")
	sysVerificationCodeRecordRowsExpectAutoSet   = strings.Join(stringx.Remove(sysVerificationCodeRecordFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	sysVerificationCodeRecordRowsWithPlaceHolder = strings.Join(stringx.Remove(sysVerificationCodeRecordFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"

	cacheSysVerificationCodeRecordIdPrefix = "cache:sysVerificationCodeRecord:id:"
)

type (
	sysVerificationCodeRecordModel interface {
		Insert(ctx context.Context, data *SysVerificationCodeRecord) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*SysVerificationCodeRecord, error)
		Update(ctx context.Context, data *SysVerificationCodeRecord) error
		Delete(ctx context.Context, id int64) error
	}

	defaultSysVerificationCodeRecordModel struct {
		sqlc.CachedConn
		table string
	}

	SysVerificationCodeRecord struct {
		Id           int64          `db:"id"`
		TenantId     int64          `db:"tenant_id"` // 所属租户ID：0=系统侧，>0=租户ID
		Channel      int64          `db:"channel"`   // 发送渠道：1邮箱 2手机短信
		Target       string         `db:"target"`    // 发送目标：邮箱或手机号
		Scene        int64          `db:"scene"`     // 业务场景
		Code         string         `db:"code"`      // 验证码
		Status       int64          `db:"status"`    // 发送状态：1成功 2失败
		Provider     sql.NullString `db:"provider"`  // 服务商
		ErrorMessage sql.NullString `db:"error_message"`
		CreateTimes  int64          `db:"create_times"`
		UpdateTimes  int64          `db:"update_times"`
	}
)

func newSysVerificationCodeRecordModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultSysVerificationCodeRecordModel {
	return &defaultSysVerificationCodeRecordModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`sys_verification_code_record`",
	}
}

func (m *defaultSysVerificationCodeRecordModel) Delete(ctx context.Context, id int64) error {
	sysVerificationCodeRecordIdKey := fmt.Sprintf("%s%v", cacheSysVerificationCodeRecordIdPrefix, id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, sysVerificationCodeRecordIdKey)
	return err
}

func (m *defaultSysVerificationCodeRecordModel) FindOne(ctx context.Context, id int64) (*SysVerificationCodeRecord, error) {
	sysVerificationCodeRecordIdKey := fmt.Sprintf("%s%v", cacheSysVerificationCodeRecordIdPrefix, id)
	var resp SysVerificationCodeRecord
	err := m.QueryRowCtx(ctx, &resp, sysVerificationCodeRecordIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sysVerificationCodeRecordRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultSysVerificationCodeRecordModel) Insert(ctx context.Context, data *SysVerificationCodeRecord) (sql.Result, error) {
	sysVerificationCodeRecordIdKey := fmt.Sprintf("%s%v", cacheSysVerificationCodeRecordIdPrefix, data.Id)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", m.table, sysVerificationCodeRecordRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.TenantId, data.Channel, data.Target, data.Scene, data.Code, data.Status, data.Provider, data.ErrorMessage, data.CreateTimes, data.UpdateTimes)
	}, sysVerificationCodeRecordIdKey)
	return ret, err
}

func (m *defaultSysVerificationCodeRecordModel) Update(ctx context.Context, data *SysVerificationCodeRecord) error {
	sysVerificationCodeRecordIdKey := fmt.Sprintf("%s%v", cacheSysVerificationCodeRecordIdPrefix, data.Id)
	_, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, sysVerificationCodeRecordRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, data.TenantId, data.Channel, data.Target, data.Scene, data.Code, data.Status, data.Provider, data.ErrorMessage, data.CreateTimes, data.UpdateTimes, data.Id)
	}, sysVerificationCodeRecordIdKey)
	return err
}

func (m *defaultSysVerificationCodeRecordModel) formatPrimary(primary any) string {
	return fmt.Sprintf("%s%v", cacheSysVerificationCodeRecordIdPrefix, primary)
}

func (m *defaultSysVerificationCodeRecordModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", sysVerificationCodeRecordRows, m.table)
	return conn.QueryRowCtx(ctx, v, query, primary)
}

func (m *defaultSysVerificationCodeRecordModel) tableName() string {
	return m.table
}

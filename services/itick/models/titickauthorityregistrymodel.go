package models

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TItickAuthorityRegistryModel = (*customTItickAuthorityRegistryModel)(nil)

type (
	// TItickAuthorityRegistryModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTItickAuthorityRegistryModel.
	TItickAuthorityRegistryModel interface {
		tItickAuthorityRegistryModel
		FindEnabled(context.Context, string) (*TItickAuthorityRegistry, error)
	}

	customTItickAuthorityRegistryModel struct {
		*defaultTItickAuthorityRegistryModel
	}
)

// NewTItickAuthorityRegistryModel returns a model for the database table.
func NewTItickAuthorityRegistryModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TItickAuthorityRegistryModel {
	return &customTItickAuthorityRegistryModel{
		defaultTItickAuthorityRegistryModel: newTItickAuthorityRegistryModel(conn, c, opts...),
	}
}

func (r *TItickAuthorityRegistry) Allows(kind string) bool {
	if r == nil || r.Status != 1 {
		return false
	}
	var kinds []string
	if json.Unmarshal([]byte(r.AllowedKinds), &kinds) != nil {
		return false
	}
	kind = strings.ToUpper(strings.TrimSpace(kind))
	for _, allowed := range kinds {
		if strings.ToUpper(strings.TrimSpace(allowed)) == kind {
			return true
		}
	}
	return false
}

func (m *defaultTItickAuthorityRegistryModel) FindEnabled(ctx context.Context, authority string) (*TItickAuthorityRegistry, error) {
	var row TItickAuthorityRegistry
	err := m.QueryRowNoCacheCtx(ctx, &row, `SELECT id,authority,producer_type,allowed_kinds,status,version,create_times,update_times
FROM t_itick_authority_registry WHERE authority=? AND status=1 LIMIT 1`, strings.ToLower(strings.TrimSpace(authority)))
	if errors.Is(err, sqlx.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

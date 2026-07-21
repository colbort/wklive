package models

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TContractFundingDifferenceAccountModel = (*customTContractFundingDifferenceAccountModel)(nil)

type (
	// TContractFundingDifferenceAccountModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTContractFundingDifferenceAccountModel.
	TContractFundingDifferenceAccountModel interface {
		tContractFundingDifferenceAccountModel
		FindEnabled(context.Context, int64, string) (*TContractFundingDifferenceAccount, error)
	}

	customTContractFundingDifferenceAccountModel struct {
		*defaultTContractFundingDifferenceAccountModel
	}
)

// NewTContractFundingDifferenceAccountModel returns a model for the database table.
func NewTContractFundingDifferenceAccountModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TContractFundingDifferenceAccountModel {
	return &customTContractFundingDifferenceAccountModel{
		defaultTContractFundingDifferenceAccountModel: newTContractFundingDifferenceAccountModel(conn, c, opts...),
	}
}

func (m *defaultTContractFundingDifferenceAccountModel) FindEnabled(ctx context.Context, tenantID int64, asset string) (*TContractFundingDifferenceAccount, error) {
	var row TContractFundingDifferenceAccount
	err := m.QueryRowNoCacheCtx(ctx, &row, "SELECT "+tContractFundingDifferenceAccountRows+" FROM t_contract_funding_difference_account WHERE tenant_id=? AND settle_asset=? AND status=1 LIMIT 1", tenantID, asset)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

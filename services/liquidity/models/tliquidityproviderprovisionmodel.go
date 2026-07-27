package models

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	ProvisionStepPending        int64 = 1
	ProvisionStepAccountCreated int64 = 2
	ProvisionStepFunded         int64 = 3
	ProvisionStepCompleted      int64 = 4
	ProvisionStepFailed         int64 = 5
)

var _ TLiquidityProviderProvisionModel = (*customTLiquidityProviderProvisionModel)(nil)

type (
	// TLiquidityProviderProvisionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTLiquidityProviderProvisionModel.
	TLiquidityProviderProvisionModel interface {
		tLiquidityProviderProvisionModel
		Reserve(ctx context.Context, providerCode, requestHash string, now int64) (*TLiquidityProviderProvision, error)
		UpdateProgress(ctx context.Context, providerCode string, tradeUserID, step int64, lastError string, now int64) error
	}

	customTLiquidityProviderProvisionModel struct {
		*defaultTLiquidityProviderProvisionModel
	}
)

// NewTLiquidityProviderProvisionModel returns a model for the database table.
func NewTLiquidityProviderProvisionModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TLiquidityProviderProvisionModel {
	return &customTLiquidityProviderProvisionModel{
		defaultTLiquidityProviderProvisionModel: newTLiquidityProviderProvisionModel(conn, c, opts...),
	}
}

func (m *customTLiquidityProviderProvisionModel) Reserve(ctx context.Context, providerCode, requestHash string, now int64) (*TLiquidityProviderProvision, error) {
	_, insertErr := m.Insert(ctx, &TLiquidityProviderProvision{
		ProviderCode: providerCode, RequestHash: requestHash,
		Step: ProvisionStepPending, CreateTimes: now, UpdateTimes: now,
	})
	row, findErr := m.FindOneByProviderCode(ctx, providerCode)
	if findErr != nil {
		if insertErr != nil {
			return nil, insertErr
		}
		return nil, findErr
	}
	if row.RequestHash != requestHash {
		return nil, fmt.Errorf("provider_code already has a different provisioning request")
	}
	return row, nil
}

func (m *customTLiquidityProviderProvisionModel) UpdateProgress(ctx context.Context, providerCode string, tradeUserID, step int64, lastError string, now int64) error {
	row, err := m.FindOneByProviderCode(ctx, providerCode)
	if err != nil {
		return err
	}
	row.TradeUserId = tradeUserID
	row.Step = step
	row.LastErrorMsg = lastError
	row.UpdateTimes = now
	return m.Update(ctx, row)
}

package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	transactionMaxAttempts = 3
	transactionRetryBase   = 10 * time.Millisecond
)

type TransactionModel interface {
	Transact(ctx context.Context, fn func(context.Context, *TransactionModels) error) error
	TransactOnce(ctx context.Context, fn func(context.Context, *TransactionModels) error) error
}

type TransactionModels struct {
	BizTradeEvent                  TBizTradeEventModel
	ContractAccountLiquidation     TContractAccountLiquidationModel
	ContractAccountLiquidationItem TContractAccountLiquidationItemModel
	ContractAdlExecution           TContractAdlExecutionModel
	ContractDeliveryBatch          TContractDeliveryBatchModel
	ContractDeliverySettlement     TContractDeliverySettlementModel
	ContractFundingBatch           TContractFundingBatchModel
	ContractFundingSettlement      TContractFundingSettlementModel
	ContractLiquidation            TContractLiquidationModel
	ContractPosition               TContractPositionModel
	ContractPositionHistory        TContractPositionHistoryModel
	ContractReconciliationIssue    TContractReconciliationIssueModel
	ContractRiskLimitTier          TContractRiskLimitTierModel
	TradeAssetReservation          TTradeAssetReservationModel
	TradeCancelLog                 TTradeCancelLogModel
	TradeFill                      TTradeFillModel
	TradeOrder                     TTradeOrderModel
	TradeOrderContract             TTradeOrderContractModel
	TradeOrderSeconds              TTradeOrderSecondsModel
	TradeOrderSpot                 TTradeOrderSpotModel
	TradeSecondsPriceSnapshot      TTradeSecondsPriceSnapshotModel
	TradeSettlementInstruction     TTradeSettlementInstructionModel
	TradeSymbol                    TTradeSymbolModel
	TradeSymbolContract            TTradeSymbolContractModel
}

type transactionModel struct {
	conn  sqlx.SqlConn
	cache cache.CacheConf
}

func NewTransactionModel(conn sqlx.SqlConn, cacheConfig cache.CacheConf) TransactionModel {
	return &transactionModel{conn: conn, cache: cacheConfig}
}

func (m *transactionModel) Transact(
	ctx context.Context, fn func(context.Context, *TransactionModels) error,
) error {
	if m == nil || m.conn == nil {
		return errors.New("transaction database is nil")
	}
	if fn == nil {
		return errors.New("transaction callback is nil")
	}
	return transactWithRetry(
		ctx, transactionMaxAttempts, transactionRetryBase,
		m.conn.TransactCtx,
		func(txCtx context.Context, session sqlx.Session) error {
			return fn(txCtx, newTransactionModels(
				sqlx.NewSqlConnFromSession(session), m.cache,
			))
		},
	)
}

func (m *transactionModel) TransactOnce(
	ctx context.Context, fn func(context.Context, *TransactionModels) error,
) error {
	if m == nil || m.conn == nil {
		return errors.New("transaction database is nil")
	}
	if fn == nil {
		return errors.New("transaction callback is nil")
	}
	return m.conn.TransactCtx(ctx, func(txCtx context.Context, session sqlx.Session) error {
		return fn(txCtx, newTransactionModels(
			sqlx.NewSqlConnFromSession(session), m.cache,
		))
	})
}

func newTransactionModels(conn sqlx.SqlConn, cacheConfig cache.CacheConf) *TransactionModels {
	return &TransactionModels{
		BizTradeEvent:                  NewTBizTradeEventModel(conn, cacheConfig),
		ContractAccountLiquidation:     NewTContractAccountLiquidationModel(conn, cacheConfig),
		ContractAccountLiquidationItem: NewTContractAccountLiquidationItemModel(conn, cacheConfig),
		ContractAdlExecution:           NewTContractAdlExecutionModel(conn, cacheConfig),
		ContractDeliveryBatch:          NewTContractDeliveryBatchModel(conn, cacheConfig),
		ContractDeliverySettlement:     NewTContractDeliverySettlementModel(conn, cacheConfig),
		ContractFundingBatch:           NewTContractFundingBatchModel(conn, cacheConfig),
		ContractFundingSettlement:      NewTContractFundingSettlementModel(conn, cacheConfig),
		ContractLiquidation:            NewTContractLiquidationModel(conn, cacheConfig),
		ContractPosition:               NewTContractPositionModel(conn, cacheConfig),
		ContractPositionHistory:        NewTContractPositionHistoryModel(conn, cacheConfig),
		ContractReconciliationIssue:    NewTContractReconciliationIssueModel(conn, cacheConfig),
		ContractRiskLimitTier:          NewTContractRiskLimitTierModel(conn, cacheConfig),
		TradeAssetReservation:          NewTTradeAssetReservationModel(conn, cacheConfig),
		TradeCancelLog:                 NewTTradeCancelLogModel(conn, cacheConfig),
		TradeFill:                      NewTTradeFillModel(conn, cacheConfig),
		TradeOrder:                     NewTTradeOrderModel(conn, cacheConfig),
		TradeOrderContract:             NewTTradeOrderContractModel(conn, cacheConfig),
		TradeOrderSeconds:              NewTTradeOrderSecondsModel(conn, cacheConfig),
		TradeOrderSpot:                 NewTTradeOrderSpotModel(conn, cacheConfig),
		TradeSecondsPriceSnapshot:      NewTTradeSecondsPriceSnapshotModel(conn, cacheConfig),
		TradeSettlementInstruction:     NewTTradeSettlementInstructionModel(conn, cacheConfig),
		TradeSymbol:                    NewTTradeSymbolModel(conn, cacheConfig),
		TradeSymbolContract:            NewTTradeSymbolContractModel(conn, cacheConfig),
	}
}

type transactionRunner func(context.Context, func(context.Context, sqlx.Session) error) error

func transactWithRetry(
	ctx context.Context,
	maxAttempts int,
	baseDelay time.Duration,
	transact transactionRunner,
	fn func(context.Context, sqlx.Session) error,
) error {
	if maxAttempts <= 0 || baseDelay < 0 || transact == nil || fn == nil {
		return errors.New("invalid deadlock retry configuration")
	}
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = transact(ctx, fn)
		if lastErr == nil || !isRetryableTransactionError(lastErr) {
			return lastErr
		}
		if attempt == maxAttempts {
			break
		}
		timer := time.NewTimer(baseDelay * time.Duration(attempt))
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
	return fmt.Errorf(
		"mysql transaction concurrency failure after %d attempts: %w",
		maxAttempts, lastErr,
	)
}

func isRetryableTransactionError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if !errors.As(err, &mysqlErr) {
		return false
	}
	return mysqlErr.Number == 1213 || mysqlErr.Number == 1205
}

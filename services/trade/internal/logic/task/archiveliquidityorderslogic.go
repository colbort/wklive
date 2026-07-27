package tasklogic

import (
	"context"
	"time"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultLiquidityArchiveRetentionDays = int64(30)
	defaultLiquidityArchiveBatchSize     = int64(500)
)

type ArchiveLiquidityOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArchiveLiquidityOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveLiquidityOrdersLogic {
	return &ArchiveLiquidityOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 归档零成交且已撤销的做市订单
func (l *ArchiveLiquidityOrdersLogic) ArchiveLiquidityOrders(_ *trade.TradeTaskReq) (*trade.TradeTaskResp, error) {
	if !l.svcCtx.Config.LiquidityOrderArchive.Enabled {
		return helpers.OkTaskResp(), nil
	}
	retentionDays := l.svcCtx.Config.LiquidityOrderArchive.RetentionDays
	if retentionDays <= 0 {
		retentionDays = defaultLiquidityArchiveRetentionDays
	}
	batchSize := l.svcCtx.Config.LiquidityOrderArchive.BatchSize
	if batchSize <= 0 {
		batchSize = defaultLiquidityArchiveBatchSize
	}

	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, "archive_liquidity_orders", func() (*trade.TradeTaskResp, error) {
		now := utils.NowMillis()
		cutoff := now - retentionDays*int64(24*time.Hour/time.Millisecond)
		for {
			count, err := l.svcCtx.TradeOrderModel.ArchiveZeroFillLiquidityOrders(
				l.ctx,
				int64(trade.OrderSourceType_ORDER_SOURCE_TYPE_LIQUIDITY),
				int64(trade.OrderStatus_ORDER_STATUS_CANCELED),
				cutoff,
				batchSize,
				now,
			)
			if err != nil {
				return nil, err
			}
			if count < batchSize {
				break
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

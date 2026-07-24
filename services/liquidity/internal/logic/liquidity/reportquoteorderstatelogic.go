package liquiditylogic

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/proto/trade"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ReportQuoteOrderStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReportQuoteOrderStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReportQuoteOrderStateLogic {
	return &ReportQuoteOrderStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ReportQuoteOrderStateLogic) ReportQuoteOrderState(in *liquidity.ReportQuoteOrderStateReq) (*liquidity.CommonResp, error) {
	if in.TenantId <= 0 || strings.TrimSpace(in.EventNo) == "" {
		return nil, fmt.Errorf("tenant_id and event_no are required")
	}
	row, err := l.svcCtx.QuoteOrderModel.FindByInternalIdentity(l.ctx, in.TenantId, in.InternalOrderId, in.InternalOrderNo, in.ClientOrderId)
	if err != nil {
		return nil, err
	}
	filled, err := strconv.ParseFloat(strings.TrimSpace(in.FilledQty), 64)
	if err != nil || filled < 0 {
		return nil, fmt.Errorf("filled_qty is invalid")
	}
	switch trade.OrderStatus(in.OrderStatus) {
	case trade.OrderStatus_ORDER_STATUS_PENDING:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_OPEN)
	case trade.OrderStatus_ORDER_STATUS_PART_FILLED:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_PART_FILLED)
	case trade.OrderStatus_ORDER_STATUS_FILLED:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FILLED)
	case trade.OrderStatus_ORDER_STATUS_CANCELING:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELING)
	case trade.OrderStatus_ORDER_STATUS_CANCELED, trade.OrderStatus_ORDER_STATUS_EXPIRED:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_CANCELED)
	case trade.OrderStatus_ORDER_STATUS_REJECTED:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_FAILED)
	default:
		row.Status = int64(liquidity.QuoteOrderStatus_QUOTE_ORDER_STATUS_UNCERTAIN)
	}
	row.InternalOrderId, row.InternalOrderNo = in.InternalOrderId, strings.TrimSpace(in.InternalOrderNo)
	row.FilledQty, row.CancelReason = filled, strings.TrimSpace(in.Reason)
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
	if err := l.svcCtx.QuoteOrderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}

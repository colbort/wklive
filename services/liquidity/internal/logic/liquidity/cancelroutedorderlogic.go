package liquiditylogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelRoutedOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelRoutedOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelRoutedOrderLogic {
	return &CancelRoutedOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelRoutedOrderLogic) CancelRoutedOrder(in *liquidity.CancelRoutedOrderReq) (*liquidity.CommonResp, error) {
	if in.ExternalOrderId <= 0 || strings.TrimSpace(in.RequestNo) == "" {
		return nil, fmt.Errorf("external_order_id and request_no are required")
	}
	row, err := l.svcCtx.ExternalOrderModel.FindOne(l.ctx, in.ExternalOrderId)
	if err != nil {
		return nil, err
	}
	switch liquidity.ExternalOrderStatus(row.Status) {
	case liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FILLED,
		liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELED,
		liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_REJECTED,
		liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FAILED:
		return &liquidity.CommonResp{Base: helper.OkResp()}, nil
	}
	provider, err := l.svcCtx.ProviderModel.FindOne(l.ctx, row.ProviderId)
	if err != nil {
		return nil, err
	}
	adapter, err := l.svcCtx.ProviderAdapters.Get(provider.VenueCode)
	if err != nil {
		return nil, err
	}
	row.Status = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELING)
	row.LastErrorMsg = strings.TrimSpace(in.Reason)
	row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
	if err := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	result, err := adapter.CancelOrder(l.ctx, provider, row)
	if err != nil {
		row.Status = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_UNCERTAIN)
		row.LastErrorCode, row.LastErrorMsg = "CANCEL_FAILED", err.Error()
		row.UpdateTimes, row.Version = time.Now().UnixMilli(), row.Version+1
		if updateErr := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); updateErr != nil {
			return nil, updateErr
		}
		return nil, err
	}
	applyExternalResult(row, result.Status, result.ExternalOrderID, result.FilledQty, result.AvgPrice, result.FeeAmount, result.FeeAsset, result.RawResponse)
	if err := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}

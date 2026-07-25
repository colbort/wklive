package liquiditylogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type RouteExternalOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRouteExternalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RouteExternalOrderLogic {
	return &RouteExternalOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RouteExternalOrderLogic) RouteExternalOrder(in *liquidity.RouteExternalOrderReq) (*liquidity.ExternalOrderResp, error) {
	if in.SymbolId <= 0 || strings.TrimSpace(in.RequestNo) == "" {
		return nil, fmt.Errorf("symbol_id and request_no are required")
	}
	if existing, err := l.svcCtx.ExternalOrderModel.FindOneByRequestNo(l.ctx, in.RequestNo); err == nil {
		return &liquidity.ExternalOrderResp{Base: helper.OkResp(), Data: helpers.ExternalOrderToProto(existing)}, nil
	} else if err != models.ErrNotFound {
		return nil, err
	}
	if in.Side != common.Side_SIDE_BUY && in.Side != common.Side_SIDE_SELL {
		return nil, fmt.Errorf("invalid side")
	}
	qty, err := parsePositive("qty", in.Qty)
	if err != nil {
		return nil, err
	}
	price := decimal.Zero
	if in.OrderType == liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_LIMIT {
		price, err = parsePositive("price", in.Price)
		if err != nil {
			return nil, err
		}
	} else if in.OrderType != liquidity.ExternalOrderType_EXTERNAL_ORDER_TYPE_MARKET {
		return nil, fmt.Errorf("invalid external order type")
	}
	config, provider, err := loadExternalRoute(l.ctx, l.svcCtx, in.SymbolId)
	if err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	row := &models.TLiquidityExternalOrder{
		OrderNo:   fmt.Sprintf("EXT%d", time.Now().UnixNano()),
		RequestNo: strings.TrimSpace(in.RequestNo), ProviderId: provider.Id, ConfigId: config.Id,
		SymbolId: in.SymbolId, ExternalSymbol: config.ExternalSymbol, Purpose: int64(in.Purpose),
		ReferenceType: strings.TrimSpace(in.ReferenceType), ReferenceId: in.ReferenceId,
		Side: int64(in.Side), OrderType: int64(in.OrderType), TimeInForce: int64(in.TimeInForce),
		Price: price, Qty: qty, ExternalClientOrderId: strings.TrimSpace(in.RequestNo),
		Status:  int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_PENDING_SUBMIT),
		Version: 1, CreateTimes: now, UpdateTimes: now,
	}
	insertResult, err := l.svcCtx.ExternalOrderModel.Insert(l.ctx, row)
	if err != nil {
		if existing, findErr := l.svcCtx.ExternalOrderModel.FindOneByRequestNo(l.ctx, in.RequestNo); findErr == nil {
			return &liquidity.ExternalOrderResp{Base: helper.OkResp(), Data: helpers.ExternalOrderToProto(existing)}, nil
		}
		return nil, err
	}
	row.Id, err = insertResult.LastInsertId()
	if err != nil {
		return nil, err
	}
	adapter, err := l.svcCtx.ProviderAdapters.Get(provider.VenueCode)
	if err != nil {
		row.Status, row.LastErrorCode, row.LastErrorMsg = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FAILED), "ADAPTER_UNAVAILABLE", err.Error()
		row.FinishedAt, row.UpdateTimes, row.Version = now, now, row.Version+1
		if updateErr := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); updateErr != nil {
			return nil, updateErr
		}
		return &liquidity.ExternalOrderResp{Base: helper.ErrResp(503, err.Error()), Data: helpers.ExternalOrderToProto(row)}, nil
	}
	result, err := adapter.SubmitOrder(l.ctx, provider, row)
	if err != nil {
		row.Status, row.LastErrorCode, row.LastErrorMsg = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_FAILED), "SUBMIT_FAILED", err.Error()
		row.FinishedAt, row.UpdateTimes, row.Version = time.Now().UnixMilli(), time.Now().UnixMilli(), row.Version+1
		if updateErr := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); updateErr != nil {
			return nil, updateErr
		}
		return &liquidity.ExternalOrderResp{Base: helper.ErrResp(502, err.Error()), Data: helpers.ExternalOrderToProto(row)}, nil
	}
	applyExternalResult(row, result.Status, result.ExternalOrderID, result.FilledQty, result.AvgPrice, result.FeeAmount, result.FeeAsset, result.RawResponse)
	if err := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.ExternalOrderResp{Base: helper.OkResp(), Data: helpers.ExternalOrderToProto(row)}, nil
}

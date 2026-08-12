package adminlogic

import (
	"context"
	"errors"
	"wklive/proto/common"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOrderListAdminLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOrderListAdminLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOrderListAdminLogic {
	return &GetOrderListAdminLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取后台订单列表
func (l *GetOrderListAdminLogic) GetOrderListAdmin(in *trade.GetOrderListAdminReq) (*trade.GetOrderListAdminResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.TradeOrderModel.FindPage(l.ctx, models.TradeOrderPageFilter{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		SymbolId:    in.SymbolId,
		ProductType: int64(in.ProductType),
		Status:      int64(in.Status),
		TimeStart:   in.TimeRange.StartTime,
		TimeEnd:     in.TimeRange.EndTime,
		Keyword:     in.Keyword,
		ExcludeSources: []int64{
			int64(trade.OrderSourceType_ORDER_SOURCE_TYPE_LIQUIDITY),
		},
	}, cursor, limit, pageutil.Count(in.Page))
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(data) > 0 {
		lastID = int64(data[len(data)-1].Id)
	}
	resp := &trade.GetOrderListAdminResp{Base: pageutil.Base(cursor, limit, len(data), total, lastID)}
	for _, item := range data {
		order := helpers.OrderToProto(item)
		if item.ProductType == int64(common.ProductType_PRODUCT_TYPE_SECONDS) {
			seconds, findErr := l.svcCtx.TradeOrderSecondsModel.FindOneByTenantIdOrderId(l.ctx, item.TenantId, item.Id)
			if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
				return nil, findErr
			}
			if findErr == nil {
				order.SecondsDirection = trade.SecondsDirection(seconds.Direction)
				order.DurationSeconds = seconds.DurationSeconds
				order.DisplayStatus = helpers.SecondsOrderDisplayStatus(seconds.SettlementStatus)
			}
		}
		resp.Data = append(resp.Data, order)
	}
	return resp, nil
}

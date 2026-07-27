package adminlogic

import (
	"context"
	"wklive/common/pageutil"
	helpers "wklive/services/trade/internal/logic/helpers"

	"wklive/proto/trade"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeliveryBatchListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeliveryBatchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeliveryBatchListLogic {
	return &GetDeliveryBatchListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 交割批次与结算明细（只读）
func (l *GetDeliveryBatchListLogic) GetDeliveryBatchList(in *trade.GetDeliveryBatchListReq) (*trade.GetDeliveryBatchListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.ContractDeliveryBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: helpers.AdminTenantID(l.ctx, in.TenantId), SymbolId: in.SymbolId, Status: int64(in.Status), TimeStart: in.TimeRange.GetStartTime(), TimeEnd: in.TimeRange.GetEndTime()}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetDeliveryBatchListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, mapper.DeliveryBatchProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

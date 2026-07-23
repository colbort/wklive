package adminlogic

import (
	"context"
	"wklive/common/pageutil"

	"wklive/proto/trade"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetDeliverySettlementListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetDeliverySettlementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetDeliverySettlementListLogic {
	return &GetDeliverySettlementListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetDeliverySettlementListLogic) GetDeliverySettlementList(in *trade.GetDeliverySettlementListReq) (*trade.GetDeliverySettlementListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.ContractDeliverySettleModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), BatchId: in.BatchId, UserId: in.UserId, PositionId: in.PositionId, Status: int64(in.Status)}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetDeliverySettlementListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, mapper.DeliverySettlementProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

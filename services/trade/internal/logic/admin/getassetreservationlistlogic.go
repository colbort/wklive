package adminlogic

import (
	"context"
	"wklive/common/pageutil"
	"wklive/services/trade/internal/logic/helpers"

	"wklive/proto/trade"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetReservationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAssetReservationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetReservationListLogic {
	return &GetAssetReservationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAssetReservationListLogic) GetAssetReservationList(in *trade.GetAssetReservationListReq) (*trade.GetAssetReservationListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.TradeAssetReservationModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: helpers.AdminTenantID(l.ctx, in.TenantId), OrderId: in.OrderId, Status: int64(in.Status)}, cursor, limit, pageutil.Count(in.Page))
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetAssetReservationListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, mapper.ReservationProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

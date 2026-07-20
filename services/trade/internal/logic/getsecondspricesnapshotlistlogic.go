package logic

import (
	"context"
	"wklive/common/pageutil"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecondsPriceSnapshotListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSecondsPriceSnapshotListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecondsPriceSnapshotListLogic {
	return &GetSecondsPriceSnapshotListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetSecondsPriceSnapshotListLogic) GetSecondsPriceSnapshotList(in *trade.GetSecondsPriceSnapshotListReq) (*trade.GetSecondsPriceSnapshotListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.TradeSecondsPriceModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), OrderId: in.OrderId, SnapshotType: int64(in.SnapshotType)}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetSecondsPriceSnapshotListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, secondsPriceProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

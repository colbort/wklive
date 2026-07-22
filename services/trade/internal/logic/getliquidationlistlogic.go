package logic

import (
	"context"
	"wklive/common/pageutil"

	"wklive/proto/trade"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLiquidationListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLiquidationListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLiquidationListLogic {
	return &GetLiquidationListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 强平、秒合约价格与结算异常（只读）
func (l *GetLiquidationListLogic) GetLiquidationList(in *trade.GetLiquidationListReq) (*trade.GetLiquidationListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.ContractLiquidationModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), UserId: in.UserId, SymbolId: in.SymbolId, PositionId: in.PositionId, Status: int64(in.Status), TimeStart: in.TimeRange.GetStartTime(), TimeEnd: in.TimeRange.GetEndTime()}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetLiquidationListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, mapper.LiquidationProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

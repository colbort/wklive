package logic

import (
	"context"
	"wklive/common/pageutil"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFundingSettlementListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFundingSettlementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFundingSettlementListLogic {
	return &GetFundingSettlementListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFundingSettlementListLogic) GetFundingSettlementList(in *trade.GetFundingSettlementListReq) (*trade.GetFundingSettlementListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	d, total, err := l.svcCtx.ContractFundingSettleModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), BatchId: in.BatchId, UserId: in.UserId, PositionId: in.PositionId, Status: int64(in.Status)}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetFundingSettlementListResp{}
	for _, v := range d {
		resp.Data = append(resp.Data, fundingSettlementProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(d), total, last)
	return resp, nil
}

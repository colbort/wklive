package tradeadminlogic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/trade"
	"wklive/services/trade/internal/mapper"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFundingBatchListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFundingBatchListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFundingBatchListLogic {
	return &GetFundingBatchListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 资金费批次与结算明细（只读）
func (l *GetFundingBatchListLogic) GetFundingBatchList(in *trade.GetFundingBatchListReq) (*trade.GetFundingBatchListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.ContractFundingBatchModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), SymbolId: in.SymbolId, Status: int64(in.Status), TimeStart: in.TimeRange.GetStartTime(), TimeEnd: in.TimeRange.GetEndTime()}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetFundingBatchListResp{}
	for _, v := range data {
		resp.Data = append(resp.Data, mapper.FundingBatchProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(data), total, last)
	return resp, nil
}

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

type GetContractRiskLimitTierListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetContractRiskLimitTierListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractRiskLimitTierListLogic {
	return &GetContractRiskLimitTierListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取合约风险限额档位列表
func (l *GetContractRiskLimitTierListLogic) GetContractRiskLimitTierList(in *trade.GetContractRiskLimitTierListReq) (*trade.GetContractRiskLimitTierListResp, error) {
	cursor, limit := pageutil.Input(in.Page)
	data, total, err := l.svcCtx.ContractRiskLimitTierModel.FindPage(l.ctx, models.AdminPageFilter{TenantId: adminTenantID(l.ctx, in.TenantId), SymbolId: in.SymbolId, Enabled: int64(in.Enabled)}, cursor, limit)
	if err != nil {
		return nil, err
	}
	last := int64(0)
	resp := &trade.GetContractRiskLimitTierListResp{}
	for _, v := range data {
		resp.Data = append(resp.Data, mapper.RiskTierProto(v))
		last = v.Id
	}
	resp.Base = pageutil.Base(cursor, limit, len(data), total, last)
	return resp, nil
}

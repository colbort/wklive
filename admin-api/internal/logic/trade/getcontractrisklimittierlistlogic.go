// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetContractRiskLimitTierListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetContractRiskLimitTierListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetContractRiskLimitTierListLogic {
	return &GetContractRiskLimitTierListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetContractRiskLimitTierListLogic) GetContractRiskLimitTierList(req *types.GetContractRiskLimitTierListReq) (resp *types.GetContractRiskLimitTierListResp, err error) {
	return logicutil.Proxy[types.GetContractRiskLimitTierListResp](l.ctx, req, l.svcCtx.TradeCli.GetContractRiskLimitTierList)
}

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

type SetContractRiskLimitTierLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetContractRiskLimitTierLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractRiskLimitTierLogic {
	return &SetContractRiskLimitTierLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetContractRiskLimitTierLogic) SetContractRiskLimitTier(req *types.SetContractRiskLimitTierReq) (resp *types.AdminCommonResp, err error) {
	return logicutil.Proxy[types.AdminCommonResp](l.ctx, req, l.svcCtx.TradeCli.SetContractRiskLimitTier)
}

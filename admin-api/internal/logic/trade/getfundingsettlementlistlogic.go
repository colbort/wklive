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

type GetFundingSettlementListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFundingSettlementListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFundingSettlementListLogic {
	return &GetFundingSettlementListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFundingSettlementListLogic) GetFundingSettlementList(req *types.GetFundingSettlementListReq) (resp *types.GetFundingSettlementListResp, err error) {
	return logicutil.Proxy[types.GetFundingSettlementListResp](l.ctx, req, l.svcCtx.TradeCli.GetFundingSettlementList)
}

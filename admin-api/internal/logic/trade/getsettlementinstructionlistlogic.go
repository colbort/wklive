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

type GetSettlementInstructionListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSettlementInstructionListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSettlementInstructionListLogic {
	return &GetSettlementInstructionListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSettlementInstructionListLogic) GetSettlementInstructionList(req *types.GetSettlementInstructionListReq) (resp *types.GetSettlementInstructionListResp, err error) {
	return logicutil.Proxy[types.GetSettlementInstructionListResp](l.ctx, req, l.svcCtx.TradeCli.GetSettlementInstructionList)
}

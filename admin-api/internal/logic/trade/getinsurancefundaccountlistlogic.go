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

type GetInsuranceFundAccountListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInsuranceFundAccountListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInsuranceFundAccountListLogic {
	return &GetInsuranceFundAccountListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInsuranceFundAccountListLogic) GetInsuranceFundAccountList(req *types.GetInsuranceFundAccountListReq) (resp *types.GetInsuranceFundAccountListResp, err error) {
	return logicutil.Proxy[types.GetInsuranceFundAccountListResp](l.ctx, req, l.svcCtx.TradeCli.GetInsuranceFundAccountList)
}

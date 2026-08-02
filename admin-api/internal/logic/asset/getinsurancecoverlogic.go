// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package asset

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInsuranceCoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetInsuranceCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInsuranceCoverLogic {
	return &GetInsuranceCoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetInsuranceCoverLogic) GetInsuranceCover(req *types.GetInsuranceCoverReq) (resp *types.InsuranceCoverResp, err error) {
	return logicutil.Proxy[types.InsuranceCoverResp](l.ctx, req, l.svcCtx.AssetCli.GetInsuranceCover)
}

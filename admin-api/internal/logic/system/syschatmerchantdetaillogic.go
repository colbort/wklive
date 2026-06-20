// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysChatMerchantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysChatMerchantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantDetailLogic {
	return &SysChatMerchantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysChatMerchantDetailLogic) SysChatMerchantDetail(req *types.SysChatMerchantDetailReq) (resp *types.SysChatMerchantDetailResp, err error) {
	return logicutil.Proxy[types.SysChatMerchantDetailResp](l.ctx, req, l.svcCtx.SystemCli.SysChatMerchantDetail)
}

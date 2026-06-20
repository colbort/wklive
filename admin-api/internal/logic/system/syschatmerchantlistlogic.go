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

type SysChatMerchantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysChatMerchantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantListLogic {
	return &SysChatMerchantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysChatMerchantListLogic) SysChatMerchantList(req *types.SysChatMerchantListReq) (resp *types.SysChatMerchantListResp, err error) {
	return logicutil.Proxy[types.SysChatMerchantListResp](l.ctx, req, l.svcCtx.SystemCli.SysChatMerchantList)
}

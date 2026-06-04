// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPayPlatformsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayPlatformsLogic {
	return &GetPayPlatformsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayPlatformsLogic) GetPayPlatforms() (resp *types.GetPayPlatformsResp, err error) {
	return logicutil.Proxy[types.GetPayPlatformsResp](l.ctx, nil, l.svcCtx.PaymentCli.GetPayPlatforms)
}

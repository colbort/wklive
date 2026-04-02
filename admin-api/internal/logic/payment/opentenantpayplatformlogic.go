// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOpenTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenTenantPayPlatformLogic {
	return &OpenTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OpenTenantPayPlatformLogic) OpenTenantPayPlatform(req *types.OpenTenantPayPlatformReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.OpenTenantPayPlatform(l.ctx, &payment.OpenTenantPayPlatformReq{
		TenantId:   req.TenantId,
		PlatformId: req.PlatformId,
		Status:     payment.CommonStatus(req.Status),
		OpenStatus: payment.OpenStatus(req.OpenStatus),
		Remark:     req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return resp, nil
}

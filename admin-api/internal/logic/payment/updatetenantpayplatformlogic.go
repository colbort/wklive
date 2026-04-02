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

type UpdateTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateTenantPayPlatformLogic {
	return &UpdateTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateTenantPayPlatformLogic) UpdateTenantPayPlatform(req *types.UpdateTenantPayPlatformReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdateTenantPayPlatform(l.ctx, &payment.UpdateTenantPayPlatformReq{
		Id:         req.Id,
		TenantId:   req.TenantId,
		Status:     payment.CommonStatus(req.Status),
		OpenStatus: payment.OpenStatus(req.OpenStatus),
		Remark:     req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

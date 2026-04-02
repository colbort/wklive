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

type UpdatePayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayPlatformLogic {
	return &UpdatePayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePayPlatformLogic) UpdatePayPlatform(req *types.UpdatePayPlatformReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdatePayPlatform(l.ctx, &payment.UpdatePayPlatformReq{
		Id:           req.Id,
		PlatformName: req.PlatformName,
		PlatformType: payment.PlatformType(req.PlatformType),
		NotifyUrl:    req.NotifyUrl,
		ReturnUrl:    req.ReturnUrl,
		Icon:         req.Icon,
		Status:       payment.CommonStatus(req.Status),
		Remark:       req.Remark,
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

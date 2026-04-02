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

type CreatePayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePayPlatformLogic {
	return &CreatePayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePayPlatformLogic) CreatePayPlatform(req *types.CreatePayPlatformReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.CreatePayPlatform(l.ctx, &payment.CreatePayPlatformReq{
		PlatformCode: req.PlatformCode,
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

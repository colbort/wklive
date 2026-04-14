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

type GetPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayPlatformLogic {
	return &GetPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayPlatformLogic) GetPayPlatform(req *types.GetPayPlatformReq) (resp *types.GetPayPlatformResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetPayPlatform(l.ctx, &payment.GetPayPlatformReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetPayPlatformResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.PayPlatform{
			Id:           result.Data.Id,
			PlatformCode: result.Data.PlatformCode,
			PlatformName: result.Data.PlatformName,
			PlatformType: int64(result.Data.PlatformType),
			NotifyUrl:    result.Data.NotifyUrl,
			ReturnUrl:    result.Data.ReturnUrl,
			Icon:         result.Data.Icon,
			Status:       int64(result.Data.Status),
			Remark:       result.Data.Remark,
			CreateTimes:  result.Data.CreateTimes,
			UpdateTimes:  result.Data.UpdateTimes,
		},
	}
	return resp, nil
}

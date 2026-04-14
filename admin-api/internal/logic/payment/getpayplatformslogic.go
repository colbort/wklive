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
	result, err := l.svcCtx.PaymentCli.GetPayPlatforms(l.ctx, &payment.AdminEmpty{})
	if err != nil {
		return nil, err
	}

	data := make([]types.PayPlatformItem, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.PayPlatformItem{
			PlatformCode: item.PlatformCode,
			PlatformName: item.PlatformName,
		}
	}

	return &types.GetPayPlatformsResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}

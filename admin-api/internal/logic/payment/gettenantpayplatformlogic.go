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

type GetTenantPayPlatformLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetTenantPayPlatformLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetTenantPayPlatformLogic {
	return &GetTenantPayPlatformLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetTenantPayPlatformLogic) GetTenantPayPlatform(req *types.GetTenantPayPlatformReq) (resp *types.GetTenantPayPlatformResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetTenantPayPlatform(l.ctx, &payment.GetTenantPayPlatformReq{
		Id:       req.Id,
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetTenantPayPlatformResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.TenantPayPlatform{
			Id:         result.Data.Id,
			TenantId:   result.Data.TenantId,
			PlatformId: result.Data.PlatformId,
			Status:     int64(result.Data.Status),
			OpenStatus: int64(result.Data.OpenStatus),
			Remark:     result.Data.Remark,
			CreateTime: result.Data.CreateTime,
			UpdateTime: result.Data.UpdateTime,
		},
	}
	return resp, nil
}

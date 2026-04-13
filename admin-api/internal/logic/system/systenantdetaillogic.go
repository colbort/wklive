// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDetailLogic {
	return &SysTenantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantDetailLogic) SysTenantDetail(req *types.SysTenantDetailReq) (resp *types.SysTenantDetailResp, err error) {
	result, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantId:   req.TenantId,
		TenantCode: req.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	return &types.SysTenantDetailResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.SysTenantItem{
			Id:           result.Data.Id,
			TenantCode:   result.Data.TenantCode,
			TenantName:   result.Data.TenantName,
			Status:       result.Data.Status,
			ExpireTime:   result.Data.ExpireTime,
			ContactName:  result.Data.ContactName,
			ContactPhone: result.Data.ContactPhone,
			Remark:       result.Data.Remark,
			CreateTimes:  result.Data.CreateTimes,
			UpdateTimes:  result.Data.UpdateTimes,
		},
	}, nil
}

package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysTenantByCodeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantByCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantByCodeLogic {
	return &SysTenantByCodeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据 code 获取租户
func (l *SysTenantByCodeLogic) SysTenantByCode(in *system.SysTenantByCodeReq) (*system.SysTenantByCodeResp, error) {
	result, err := l.svcCtx.TenantMode.FindByTenantCode(l.ctx, in.TenantCode)
	if err != nil {
		return nil, err
	}
	return &system.SysTenantByCodeResp{
		Base: &common.RespBase{
			Code: 200,
			Msg:  "获取租户成功",
		},
		Data: &system.SysTenantItem{
			Id:           result.Id,
			TenantCode:   result.TenantCode,
			TenantName:   result.TenantName,
			Status:       result.Status,
			ExpireTime:   result.ExpireTime,
			ContactName:  result.ContactName.String,
			ContactPhone: result.ContactPhone.String,
			Remark:       result.Remark.String,
			CreateTimes:  result.CreateTimes,
			UpdateTimes:  result.UpdateTimes,
		},
	}, nil
}

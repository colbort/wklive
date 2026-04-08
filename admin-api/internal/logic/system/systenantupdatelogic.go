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

type SysTenantUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysTenantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantUpdateLogic {
	return &SysTenantUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysTenantUpdateLogic) SysTenantUpdate(req *types.SysTenantUpdateReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.SystemCli.SysTenantUpdate(l.ctx, &system.SysTenantUpdateReq{
		Id:           req.Id,
		TenantName:   req.TenantName,
		Status:       req.Status,
		ExpireTime:   req.ExpireTime,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		Remark:       req.Remark,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

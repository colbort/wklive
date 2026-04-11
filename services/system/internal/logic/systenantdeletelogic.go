package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
)

type SysTenantDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantDeleteLogic {
	return &SysTenantDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除租户
func (l *SysTenantDeleteLogic) SysTenantDelete(in *system.SysTenantDeleteReq) (*system.RespBase, error) {
	tenant, err := l.svcCtx.TenantMode.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}
	err = l.svcCtx.TenantMode.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

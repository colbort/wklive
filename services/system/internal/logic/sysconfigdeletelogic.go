package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigDeleteLogic {
	return &SysConfigDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除系统配置
func (l *SysConfigDeleteLogic) SysConfigDelete(in *system.SysConfigDeleteReq) (*system.RespBase, error) {
	config, err := l.svcCtx.ConfigModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || config == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.BusinessDataNotFound, i18n.Translate(i18n.BusinessDataNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if base, err := adminTenantWriteScopeResp(l.ctx, config.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	if err := l.svcCtx.ConfigModel.Delete(l.ctx, in.Id); err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

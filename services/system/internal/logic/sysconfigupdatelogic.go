package logic

import (
	"context"
	"database/sql"
	uc "wklive/common/utils"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/internal/utils"
)

type SysConfigUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigUpdateLogic {
	return &SysConfigUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新系统配置
func (l *SysConfigUpdateLogic) SysConfigUpdate(in *system.SysConfigUpdateReq) (*system.RespBase, error) {
	config, err := l.svcCtx.ConfigModel.FindOne(l.ctx, in.Id)
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

	configKey := config.ConfigKey.String
	configValue := config.ConfigValue.String
	if in.ConfigKey != "" {
		configKey = in.ConfigKey
	}
	if in.ConfigValue != "" {
		configValue = in.ConfigValue
	}
	if err := utils.CheckConfig(configKey, configValue); err != nil {
		return nil, errorx.Wrap(err, i18n.Translate(i18n.ConfigValidationFailed, l.ctx))
	}

	if in.ConfigKey != "" {
		config.ConfigKey = sql.NullString{String: in.ConfigKey, Valid: true}
	}
	if in.ConfigValue != "" {
		config.ConfigValue = sql.NullString{String: in.ConfigValue, Valid: true}
	}
	if in.Remark != "" {
		config.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	config.UpdateTimes = uc.NowMillis()
	err = l.svcCtx.ConfigModel.Update(l.ctx, config)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

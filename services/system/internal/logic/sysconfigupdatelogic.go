package logic

import (
	"context"
	"database/sql"
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
	if err := utils.CheckConfig(in.ConfigKey, in.ConfigValue); err != nil {
		return nil, errorx.Wrap(err, i18n.Translate(i18n.ConfigValidationFailed, l.ctx))
	}
	config, err := l.svcCtx.ConfigModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
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
	err = l.svcCtx.ConfigModel.Update(l.ctx, config)
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

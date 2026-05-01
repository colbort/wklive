package logic

import (
	"context"
	"database/sql"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"

	uc "wklive/common/utils"
	"wklive/services/system/internal/utils"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigCreateLogic {
	return &SysConfigCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 新增系统配置
func (l *SysConfigCreateLogic) SysConfigCreate(in *system.SysConfigCreateReq) (*system.RespBase, error) {
	if err := utils.CheckConfig(in.ConfigKey, in.ConfigValue); err != nil {
		return nil, errorx.Wrap(err, i18n.Translate(i18n.ConfigValidationFailed, l.ctx))
	}
	config, err := l.svcCtx.ConfigModel.FindOneByTenantIdConfigKey(l.ctx, in.TenantId, sql.NullString{String: in.ConfigKey, Valid: true})
	if err != nil && err != models.ErrNotFound {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}
	if config != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.ConfigAlreadyExists, l.ctx)),
		}, nil
	}
	_, err = l.svcCtx.ConfigModel.Insert(l.ctx, &models.SysConfig{
		TenantId:    in.TenantId,
		ConfigKey:   sql.NullString{String: in.ConfigKey, Valid: true},
		ConfigValue: sql.NullString{String: in.ConfigValue, Valid: true},
		Remark:      sql.NullString{String: in.Remark, Valid: true},
		CreateTimes: uc.NowMillis(),
		UpdateTimes: uc.NowMillis(),
	})
	if err != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(500, err.Error()),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

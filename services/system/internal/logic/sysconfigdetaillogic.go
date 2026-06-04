package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysConfigDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigDetailLogic {
	return &SysConfigDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取系统配置详情
func (l *SysConfigDetailLogic) SysConfigDetail(in *system.SysConfigDetailReq) (*system.SysConfigDetailResp, error) {
	var config *models.SysConfig
	var err error
	if in.Id != nil && *in.Id > 0 {
		config, err = l.svcCtx.ConfigModel.FindOne(l.ctx, *in.Id)

	} else if in.ConfigKey != nil && in.ConfigKey.String() != "" {
		config, err = l.svcCtx.ConfigModel.FindOneByTenantIdConfigKey(l.ctx, *in.TenantId, sql.NullString{String: in.ConfigKey.String(), Valid: true})
	} else {
		err = i18n.StatusError(l.ctx, i18n.InvalidQueryCondition)
	}
	if err != nil {
		return nil, err
	}
	return &system.SysConfigDetailResp{
		Base: helper.OkResp(),
		Data: &system.SysConfigItem{
			Id:          config.Id,
			ConfigKey:   config.ConfigKey.String,
			ConfigValue: config.ConfigValue.String,
			Remark:      config.Remark.String,
			CreateTimes: config.CreateTimes,
			UpdateTimes: config.UpdateTimes,
			TenantId:    config.TenantId,
		},
	}, nil
}

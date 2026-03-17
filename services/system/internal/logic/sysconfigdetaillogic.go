package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/utils"
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

	} else if in.ConfigKey != nil {
		config, err = l.svcCtx.ConfigModel.FindOneByConfigKey(l.ctx, sql.NullString{String: in.ConfigKey.String(), Valid: true})
	} else {
		err = errors.New("无效的查询条件")
	}
	if err != nil {
		return nil, err
	}

	value, err := utils.StringToStruct(config.ConfigValue.String)
	if err != nil {
		return nil, err
	}
	return &system.SysConfigDetailResp{
		Base: &system.RespBase{
			Code: 200,
			Msg:  "查询成功",
		},
		Data: &system.SysConfigItem{
			Id:          config.Id,
			ConfigKey:   config.ConfigKey.String,
			ConfigValue: value,
			Remark:      config.Remark.String,
			CreatedAt:   config.CreatedAt.Unix(),
		},
	}, nil
}

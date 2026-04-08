package logic

import (
	"context"
	"database/sql"

	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"

	"wklive/services/system/internal/utils"
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
		return nil, errorx.Wrap(err, "配置项校验失败")
	}
	config, err := l.svcCtx.ConfigModel.FindOneByConfigKey(l.ctx, sql.NullString{String: in.ConfigKey, Valid: true})
	if err != nil && err != models.ErrNotFound {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 500,
				Msg:  err.Error(),
			},
		}, nil
	}
	if config != nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 400,
				Msg:  "配置项已存在",
			},
		}, nil
	}
	_, err = l.svcCtx.ConfigModel.Insert(l.ctx, &models.SysConfig{
		ConfigKey:   sql.NullString{String: in.ConfigKey, Valid: true},
		ConfigValue: sql.NullString{String: in.ConfigValue, Valid: true},
		Remark:      sql.NullString{String: in.Remark, Valid: true},
	})
	if err != nil {
		return &system.RespBase{
			Base: &common.RespBase{
				Code: 500,
				Msg:  err.Error(),
			},
		}, nil
	}

	return &system.RespBase{
		Base: &common.RespBase{
			Code: 200,
			Msg:  "新增成功",
		},
	}, nil
}

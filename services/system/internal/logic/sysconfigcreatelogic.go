package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

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
	config, err := l.svcCtx.ConfigModel.FindOneByConfigKey(l.ctx, sql.NullString{String: in.ConfigKey, Valid: true})
	if err != nil {
		return nil, err
	}
	if config != nil {
		return nil, errors.New("配置已存在")
	}
	_, err = l.svcCtx.ConfigModel.Insert(l.ctx, &models.SysConfig{
		ConfigKey:   sql.NullString{String: in.ConfigKey, Valid: true},
		ConfigValue: sql.NullString{String: in.ConfigValue.String(), Valid: true},
		Remark:      sql.NullString{String: in.Remark, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &system.RespBase{
		Code: 200,
		Msg:  "新增成功",
	}, nil
}

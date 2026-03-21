// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigCreateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysConfigCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigCreateLogic {
	return &SysConfigCreateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysConfigCreateLogic) SysConfigCreate(req *types.SysConfigCreateReq) (resp *types.RespBase, err error) {
	if utils.CheckConfig(req.ConfigKey, req.ConfigValue) != nil {
		return nil, errorx.Wrap(err, "配置项校验失败")
	}
	value, err := utils.StringToStruct(req.ConfigValue)
	if err != nil {
		return nil, err
	}

	out, err := l.svcCtx.SystemCli.SysConfigCreate(l.ctx, &system.SysConfigCreateReq{
		ConfigKey:   req.ConfigKey,
		ConfigValue: value,
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}
	resp = &types.RespBase{
		Code: out.Code,
		Msg:  out.Msg,
	}
	return
}

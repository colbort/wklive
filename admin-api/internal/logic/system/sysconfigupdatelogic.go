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

type SysConfigUpdateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysConfigUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigUpdateLogic {
	return &SysConfigUpdateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysConfigUpdateLogic) SysConfigUpdate(req *types.SysConfigUpdateReq) (resp *types.RespBase, err error) {
	if utils.CheckConfig(req.ConfigKey, req.ConfigValue) != nil {
		return nil, errorx.Wrap(err, "配置项校验失败")
	}
	value, err := utils.StringToStruct(req.ConfigValue)
	if err != nil {
		return nil, err
	}
	out, err := l.svcCtx.SystemCli.SysConfigUpdate(l.ctx, &system.SysConfigUpdateReq{
		Id:          req.Id,
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

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package system

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysConfigKeysLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSysConfigKeysLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysConfigKeysLogic {
	return &SysConfigKeysLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SysConfigKeysLogic) SysConfigKeys() (resp *types.SysConfigKeysResp, err error) {
	out, err := l.svcCtx.SystemCli.SysConfigKeys(l.ctx, &system.Empty{})
	if err != nil {
		return nil, err
	}

	resp = &types.SysConfigKeysResp{
		RespBase: types.RespBase{
			Code: out.Base.Code,
			Msg:  out.Base.Msg,
		},
		Data: out.Data,
	}
	return
}

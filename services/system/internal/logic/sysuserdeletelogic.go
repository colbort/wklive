package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserDeleteLogic {
	return &SysUserDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserDeleteLogic) SysUserDelete(in *system.SysUserDeleteReq) (*system.RespBase, error) {
	err := l.svcCtx.UserModel.Delete(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "删除成功",
	}, nil
}

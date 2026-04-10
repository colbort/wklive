package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type SysUserUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserUpdateLogic {
	return &SysUserUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserUpdateLogic) SysUserUpdate(in *system.SysUserUpdateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.UserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return &system.RespBase{
			Base: helper.GetErrResp(400, "用户不存在"),
		}, nil
	}
	var data models.SysUser
	_ = copier.Copy(&data, one)
	_ = copier.Copy(&data, in)

	err = l.svcCtx.UserModel.Update(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

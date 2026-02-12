package logic

import (
	"context"

	"wklive/rpc/system"
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
			Code: 400,
			Msg:  "用户不存在",
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
		Code: 200,
		Msg:  "更新成功",
	}, nil
}

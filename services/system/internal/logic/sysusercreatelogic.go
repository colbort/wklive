package logic

import (
	"context"

	"wklive/rpc/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type SysUserCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysUserCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysUserCreateLogic {
	return &SysUserCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SysUserCreateLogic) SysUserCreate(in *system.SysUserCreateReq) (*system.RespBase, error) {
	one, err := l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if one != nil {
		return &system.RespBase{
			Code: 400,
			Msg:  "用户名已存在",
		}, nil
	}
	var data models.SysUser
	_ = copier.Copy(&data, in)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	data.Password = string(hashedPassword)

	_, err = l.svcCtx.UserModel.Insert(l.ctx, &data)
	if err != nil {
		return nil, err
	}
	return &system.RespBase{
		Code: 200,
		Msg:  "创建成功",
	}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) (resp *types.CreateUserResp, err error) {
	result, err := l.svcCtx.UserCli.CreateUser(l.ctx, &user.CreateUserReq{
		TenantId:       req.TenantId,
		Username:       req.Username,
		Nickname:       req.Nickname,
		Avatar:         req.Avatar,
		Phone:          req.Phone,
		Email:          req.Email,
		Password:       req.Password,
		RegisterType:   user.RegisterType(req.RegisterType),
		Status:         user.UserStatus(req.Status),
		MemberLevel:    req.MemberLevel,
		Language:       req.Language,
		Timezone:       req.Timezone,
		InviteCode:     req.InviteCode,
		Signature:      req.Signature,
		Source:         req.Source,
		ReferrerUserId: req.ReferrerUserId,
		Remark:         req.Remark,
	})
	if err != nil {
		return nil, err
	}

	return &types.CreateUserResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		UserId: result.UserId,
	}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/system"
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
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &req.TenantCode,
	})
	if err != nil {
		return nil, err
	}
	if tenant.Base.Code != 200 || tenant.Data == nil {
		return &types.CreateUserResp{
			RespBase: types.RespBase{
				Code: tenant.Base.Code,
				Msg:  tenant.Base.Msg,
			},
		}, nil
	}

	referrerUserId := req.ReferrerUserId
	if strings.TrimSpace(req.ReferrerInviteCode) != "" {
		referrer, err := resolveReferrerByInviteCode(l.svcCtx, l.ctx, tenant.Data.Id, req.ReferrerInviteCode)
		if err != nil {
			return nil, err
		}
		if referrer == nil {
			return &types.CreateUserResp{
				RespBase: types.RespBase{
					Code: 404,
					Msg:  "推荐人不存在",
				},
			}, nil
		}
		referrerUserId = referrer.Id
	}

	result, err := l.svcCtx.UserCli.CreateUser(l.ctx, &user.CreateUserReq{
		TenantCode:     req.TenantCode,
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
		ReferrerUserId: referrerUserId,
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

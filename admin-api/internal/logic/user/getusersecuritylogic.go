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

type GetUserSecurityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityLogic {
	return &GetUserSecurityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserSecurityLogic) GetUserSecurity(req *types.GetUserSecurityReq) (resp *types.GetUserSecurityResp, err error) {
	result, err := l.svcCtx.UserCli.GetUserSecurity(l.ctx, &user.GetUserSecurityReq{
		TenantId: req.TenantId,
		UserId:   req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &types.GetUserSecurityResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Security: types.UserSecurity{
			Id:              result.Security.Id,
			TenantId:        result.Security.TenantId,
			UserId:          result.Security.UserId,
			HasPayPassword:  result.Security.HasPayPassword,
			GoogleEnabled:   result.Security.GoogleEnabled,
			LoginErrorCount: result.Security.LoginErrorCount,
			PayErrorCount:   result.Security.PayErrorCount,
			LockUntil:       result.Security.LockUntil,
			RiskLevel:       int64(result.Security.RiskLevel),
			CreateTime:      result.Security.CreateTime,
			UpdateTime:      result.Security.UpdateTime,
		},
	}, nil
}

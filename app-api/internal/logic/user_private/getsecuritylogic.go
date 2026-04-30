// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSecurityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSecurityLogic {
	return &GetSecurityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSecurityLogic) GetSecurity() (resp *types.GetSecurityResp, err error) {
	result, err := l.svcCtx.UserCli.GetSecurity(l.ctx, &user.GetSecurityReq{})
	if err != nil {
		return nil, err
	}
	return &types.GetSecurityResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.UserSecurity{
			Id:              result.Security.Id,
			TenantId:        result.Security.TenantId,
			UserId:          result.Security.UserId,
			HasPayPassword:  result.Security.PayPasswordHash != "",
			GoogleEnabled:   result.Security.GoogleEnabled,
			LoginErrorCount: result.Security.LoginErrorCount,
			PayErrorCount:   result.Security.PayErrorCount,
			LockUntil:       result.Security.LockUntil,
			RiskLevel:       int64(result.Security.RiskLevel.Number()),
			CreateTimes:     result.Security.CreateTimes,
			UpdateTimes:     result.Security.UpdateTimes,
		},
	}, nil
}

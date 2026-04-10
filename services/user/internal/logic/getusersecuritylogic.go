package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserSecurityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserSecurityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserSecurityLogic {
	return &GetUserSecurityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户安全设置
func (l *GetUserSecurityLogic) GetUserSecurity(in *user.GetUserSecurityReq) (*user.GetUserSecurityResp, error) {
	result, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil {
		return nil, err
	}

	return &user.GetUserSecurityResp{
		Base: helper.OkResp(),
		Security: &user.UserSecurity{
			Id:              result.Id,
			TenantId:        result.TenantId,
			UserId:          result.UserId,
			HasPayPassword:  result.PayPasswordHash.Valid && result.PayPasswordHash.String != "",
			GoogleEnabled:   result.GoogleEnabled == 1,
			LoginErrorCount: result.LoginErrorCount,
			PayErrorCount:   result.PayErrorCount,
			LockUntil:       result.LockUntil,
			RiskLevel:       user.RiskLevel(result.RiskLevel),
			CreateTimes:     result.CreateTimes,
			UpdateTimes:     result.UpdateTimes,
		},
	}, nil
}

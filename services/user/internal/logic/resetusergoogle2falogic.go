package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetUserGoogle2FALogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetUserGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetUserGoogle2FALogic {
	return &ResetUserGoogle2FALogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 重置用户谷歌2FA
func (l *ResetUserGoogle2FALogic) ResetUserGoogle2FA(in *user.ResetUserGoogle2FAReq) (*user.AdminCommonResp, error) {
	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{Code: 404, Msg: "用户安全信息不存在"},
		}, nil
	}

	// 禁用 Google 2FA
	err = l.svcCtx.UserSecurityModel.Update(l.ctx, &models.TUserSecurity{
		Id:          userSecurity.Id,
		GoogleEnabled: 0,
		UpdateTimes: time.Now().UnixMilli(),
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员重置用户 %d 的 Google2FA", in.UserId)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}

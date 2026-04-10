package logic

import (
	"context"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlockUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlockUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlockUserLogic {
	return &UnlockUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 解锁用户（解除登录锁定）
func (l *UnlockUserLogic) UnlockUser(in *user.UnlockUserReq) (*user.AdminCommonResp, error) {
	// 获取用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if userSecurity == nil {
		return &user.AdminCommonResp{
			Base: helper.GetErrResp(404, "用户安全信息不存在"),
		}, nil
	}

	// 重置登录失败计数和解除锁定
	userSecurity.LoginErrorCount = 0
	userSecurity.LockUntil = 0
	userSecurity.UpdateTimes = time.Now().UnixMilli()

	err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员解锁用户 %d", in.UserId)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}

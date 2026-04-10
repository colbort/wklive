package logic

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ResetPayPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewResetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPayPasswordLogic {
	return &ResetPayPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员重置支付密码
func (l *ResetPayPasswordLogic) ResetPayPassword(in *user.ResetPayPasswordReq) (*user.AdminCommonResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.AdminCommonResp{
			Base: &common.RespBase{
				Code: 404,
				Msg:  "用户不存在",
			},
		}, nil
	}

	// 获取或创建用户安全信息
	userSecurity, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	now := time.Now().UnixMilli()
	if userSecurity != nil {
		userSecurity.PayPasswordHash = sql.NullString{String: in.NewPassword, Valid: true}
		userSecurity.UpdateTimes = now

		err = l.svcCtx.UserSecurityModel.Update(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	} else {
		userSecurity = &models.TUserSecurity{
			Id:              l.svcCtx.Node.Generate().Int64(),
			TenantId:        in.TenantId,
			UserId:          in.UserId,
			PayPasswordHash: sql.NullString{String: in.NewPassword, Valid: true},
			CreateTimes:    now,
			UpdateTimes:    now,
		}

		_, err = l.svcCtx.UserSecurityModel.Insert(l.ctx, userSecurity)
		if err != nil {
			return nil, err
		}
	}

	l.Logger.Infof("管理员为用户 %d 重置支付密码成功", in.UserId)

	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}
package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户
func (l *DeleteUserLogic) DeleteUser(in *user.DeleteUserReq) (*user.AdminCommonResp, error) {
	tuser, err := l.svcCtx.UserModel.FindByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.FailWithCode(401),
		}, nil
	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserIdentityModel)
		userSecurityModel := models.NewTUserSecurityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserSecurityModel)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)

		if err := userIdentityModel.DeleteByUserId(ctx, tuser.Id); err != nil {
			return err
		}
		if err := userSecurityModel.DeleteByUserId(ctx, tuser.Id); err != nil {
			return err
		}
		if err := userModel.Delete(ctx, tuser.Id); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}

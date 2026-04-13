package logic

import (
	"context"
	"database/sql"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UpdateUserBaseLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserBaseLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBaseLogic {
	return &UpdateUserBaseLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员更新用户基本信息
func (l *UpdateUserBaseLogic) UpdateUserBase(in *user.UpdateUserBaseReq) (*user.UpdateUserBaseResp, error) {
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.UpdateUserBaseResp{
			Base: helper.GetErrResp(404, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()

	// 更新用户基本信息
	if in.Username != "" {
		tuser.Username = in.Username
	}
	if in.Nickname != "" {
		tuser.Nickname = sql.NullString{String: in.Nickname, Valid: true}
	}
	if in.Avatar != "" {
		tuser.Avatar = sql.NullString{String: in.Avatar, Valid: true}
	}
	if in.Language != "" {
		tuser.Language = sql.NullString{String: in.Language, Valid: true}
	}
	if in.Timezone != "" {
		tuser.Timezone = sql.NullString{String: in.Timezone, Valid: true}
	}
	if in.Signature != "" {
		tuser.Signature = sql.NullString{String: in.Signature, Valid: true}
	}
	if in.Source != "" {
		tuser.Source = sql.NullString{String: in.Source, Valid: true}
	}
	if in.ReferrerUserId > 0 {
		tuser.ReferrerUserId = sql.NullInt64{Int64: in.ReferrerUserId, Valid: true}
	}
	if in.Remark != "" {
		tuser.Remark = sql.NullString{String: in.Remark, Valid: true}
	}

	tuser.UpdateTimes = now

	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity != nil {
		if in.Phone != "" {
			identity.Phone = sql.NullString{String: in.Phone, Valid: true}
		}
		if in.Email != "" {
			identity.Email = sql.NullString{String: in.Email, Valid: true}
		}
		identity.UpdateTimes = now

	}

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserIdentityModel)

		if err := userModel.Update(ctx, tuser); err != nil {
			return err
		}
		if identity != nil {
			if err := userIdentityModel.Update(ctx, identity); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员更新用户 %d 基本信息成功", in.UserId)

	// 返回完整用户详情
	security, err := l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil {
		return nil, err
	}

	userDetailResp := toUserDetailProto(tuser, identity, security, nil)

	return &user.UpdateUserBaseResp{
		Base:   helper.OkResp(),
		Detail: userDetailResp,
	}, nil
}

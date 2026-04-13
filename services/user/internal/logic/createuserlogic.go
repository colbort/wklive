package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员创建用户
func (l *CreateUserLogic) CreateUser(in *user.CreateUserReq) (*user.CreateUserResp, error) {
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant.Base.Code != 200 {
		return &user.CreateUserResp{
			Base: helper.FailWithCode(i18n.TenantNotFound),
		}, nil
	}

	now := utils.NowMillis()
	userNo := l.svcCtx.Node.Generate().Int64()

	// 创建用户基本信息
	tuser := &models.TUser{
		TenantId:       tenant.Data.Id,
		UserNo:         generateUserNo(userNo),
		Username:       in.Username,
		Nickname:       sql.NullString{String: in.Nickname, Valid: in.Nickname != ""},
		Avatar:         sql.NullString{String: in.Avatar, Valid: in.Avatar != ""},
		PasswordHash:   in.Password,
		RegisterType:   int64(in.RegisterType),
		Status:         int64(in.Status),
		MemberLevel:    int64(in.MemberLevel),
		Language:       sql.NullString{String: in.Language, Valid: in.Language != ""},
		Timezone:       sql.NullString{String: in.Timezone, Valid: in.Timezone != ""},
		InviteCode:     sql.NullString{String: in.InviteCode, Valid: in.InviteCode != ""},
		Signature:      sql.NullString{String: in.Signature, Valid: in.Signature != ""},
		Source:         sql.NullString{String: in.Source, Valid: in.Source != ""},
		ReferrerUserId: sql.NullInt64{Int64: in.ReferrerUserId, Valid: in.ReferrerUserId > 0},
		RegisterTime:   now,
		RegisterIp:     sql.NullString{},
		Remark:         sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		Deleted:        0,
		CreateTimes:    now,
		UpdateTimes:    now,
		IsGuest:        1,
		IsRecharge:     0,
		DeviceId:       "",
		Fingerprint:    "",
	}

	userIdentity := &models.TUserIdentity{
		Id:          l.svcCtx.Node.Generate().Int64(),
		CreateTimes: now,
		UpdateTimes: now,
	}

	userSecurity := &models.TUserSecurity{
		Id:          l.svcCtx.Node.Generate().Int64(),
		TenantId:    tenant.Data.Id,
		CreateTimes: now,
		UpdateTimes: now,
	}

	userId := int64(0)

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserIdentityModel)
		userSecurityModel := models.NewTUserSecurityModel(conn, l.svcCtx.Config.CacheRedis).(models.UserSecurityModel)

		result, err := userModel.Insert(ctx, tuser)

		if err != nil {
			return err
		}

		userId, err = result.LastInsertId()
		if err != nil {
			return err
		}
		userIdentity.UserId = userId
		userSecurity.UserId = userId
		if _, err := userIdentityModel.Insert(ctx, userIdentity); err != nil {
			return err
		}
		if _, err := userSecurityModel.Insert(ctx, userSecurity); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	l.Logger.Infof("管理员创建用户成功，用户ID：%d，用户名：%s", userId, in.Username)

	return &user.CreateUserResp{
		Base:   helper.OkResp(),
		UserId: userId,
	}, nil
}

func generateUserNo(userId int64) string {
	return fmt.Sprintf("%06d", userId)
}

package logic

import (
	"context"
	"database/sql"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
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
	now := utils.NowMillis()
	userId := l.svcCtx.Node.Generate().Int64()

	// 创建用户基本信息
	tuser := &models.TUser{
		Id:             userId,
		TenantId:       in.TenantId,
		UserNo:         generateUserNo(userId),
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

	_, err := l.svcCtx.UserModel.Insert(l.ctx, tuser)
	if err != nil {
		return nil, err
	}

	// 创建用户身份信息
	userIdentity := &models.TUserIdentity{
		Id:          l.svcCtx.Node.Generate().Int64(),
		TenantId:    in.TenantId,
		UserId:      userId,
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err = l.svcCtx.UserIdentityModel.Insert(l.ctx, userIdentity)
	if err != nil {
		return nil, err
	}

	// 创建用户安全信息
	userSecurity := &models.TUserSecurity{
		Id:          l.svcCtx.Node.Generate().Int64(),
		TenantId:    in.TenantId,
		UserId:      userId,
		CreateTimes: now,
		UpdateTimes: now,
	}

	_, err = l.svcCtx.UserSecurityModel.Insert(l.ctx, userSecurity)
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
	// TODO: 实现用户编号生成逻辑
	return ""
}

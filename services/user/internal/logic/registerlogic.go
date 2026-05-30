package logic

import (
	"context"
	"database/sql"
	"encoding/json"
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
	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 用户注册
func (l *RegisterLogic) Register(in *user.RegisterReq) (*user.RegisterResp, error) {
	tenant, err := l.svcCtx.SystemCli.SysTenantDetail(l.ctx, &system.SysTenantDetailReq{
		TenantCode: &in.TenantCode,
	})
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tenant == nil {
		return &user.RegisterResp{
			Base: helper.GetErrResp(401, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}
	if tenant.Base.Code != 200 {
		return &user.RegisterResp{
			Base: helper.GetErrResp(tenant.Base.Code, tenant.Base.Msg),
		}, nil
	}
	if tenant.Data.Status != 1 {
		return &user.RegisterResp{
			Base: helper.GetErrResp(502, i18n.Translate(i18n.TenantDisabled, l.ctx)),
		}, nil
	}
	if tenant.Data.ExpireTime < utils.NowMillis() {
		return &user.RegisterResp{
			Base: helper.GetErrResp(502, i18n.Translate(i18n.TenantExpired, l.ctx)),
		}, nil
	}

	var tuser *models.TUser
	var userIdentify *models.TUserIdentity
	switch in.RegisterType {
	case user.RegisterType_REGISTER_TYPE_EMAIL:
		userIdentify, err = l.svcCtx.UserIdentityModel.FindByEmail(l.ctx, tenant.Data.Id, in.Email)
	case user.RegisterType_REGISTER_TYPE_PHONE:
		userIdentify, err = l.svcCtx.UserIdentityModel.FindByPhone(l.ctx, tenant.Data.Id, in.Phone)
	case user.RegisterType_REGISTER_TYPE_USERNAME:
		if in.InviteCode == "" {
			return &user.RegisterResp{
				Base: helper.GetErrResp(201, i18n.Translate(i18n.InviteCodeRequired, l.ctx)),
			}, nil
		}
		parent, err := l.svcCtx.UserModel.FindByInviteCode(l.ctx, in.InviteCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if parent == nil {
			return &user.RegisterResp{
				Base: helper.GetErrResp(201, i18n.Translate(i18n.InviterNotFound, l.ctx)),
			}, nil
		}
		count, err := l.svcCtx.UserModel.CountRecentNoRecharge(l.ctx, parent.Id)
		if err != nil {
			return nil, err
		}
		if count > 7 {
			return &user.RegisterResp{
				Base: helper.GetErrResp(201, i18n.Translate(i18n.RegistrationTooFrequent, l.ctx)),
			}, nil
		}
		tuser, err = l.svcCtx.UserModel.FindByUsername(l.ctx, in.TenantCode, in.Username)
		// 如果是用户名密码注册的 必须要邀请码，同一个 邀请码 的最近 一周内超过7个注册的用户如果没有一个充值的不给注册，直到 有用户充值
	case user.RegisterType_REGISTER_TYPE_GUEST:

	}
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, nil
	}
	if tuser != nil || userIdentify != nil {
		return &user.RegisterResp{
			Base: helper.GetErrResp(201, i18n.Translate(i18n.UserAlreadyExists, l.ctx)),
		}, nil
	}
	referrerUserId := int64(-1)
	if in.InviteCode != "" {
		parent, err := l.svcCtx.UserModel.FindByInviteCode(l.ctx, in.InviteCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if parent == nil {
			return &user.RegisterResp{
				Base: helper.GetErrResp(201, i18n.Translate(i18n.InviterNotFound, l.ctx)),
			}, nil
		}
		referrerUserId = parent.Id
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(hashedPassword)
	userNo := l.svcCtx.Node.Generate().Int64()

	now := utils.NowMillis()
	tuser = &models.TUser{
		TenantId:       tenant.Data.Id,
		UserNo:         fmt.Sprintf("U%d", userNo),
		Username:       fmt.Sprintf("U%d", userNo),
		Nickname:       sql.NullString{String: "", Valid: true},
		Avatar:         sql.NullString{String: "", Valid: true},
		PasswordHash:   passwordHash,
		RegisterType:   int64(in.RegisterType),
		Status:         1,
		MemberLevel:    0,
		Language:       sql.NullString{String: "", Valid: true},
		Timezone:       sql.NullString{String: "", Valid: true},
		InviteCode:     sql.NullString{String: "", Valid: true},
		Signature:      sql.NullString{String: "", Valid: true},
		Source:         sql.NullString{String: "", Valid: true},
		ReferrerUserId: sql.NullInt64{Int64: referrerUserId, Valid: true},
		LastLoginIp:    sql.NullString{String: in.RegisterIp, Valid: true},
		LastLoginTime:  now,
		RegisterIp:     sql.NullString{String: in.RegisterIp, Valid: true},
		RegisterTime:   now,
		IsGuest:        1,
		IsRecharge:     0,
		DeviceId:       "",
		Fingerprint:    sql.NullString{String: "", Valid: true},
		Remark:         sql.NullString{String: "", Valid: true},
		Deleted:        0,
		CreateTimes:    now,
		UpdateTimes:    now,
	}

	identity := &models.TUserIdentity{
		TenantId:    tenant.Data.Id,
		Phone:       sql.NullString{String: in.Phone, Valid: in.RegisterType == user.RegisterType_REGISTER_TYPE_PHONE && in.Phone != ""},
		Email:       sql.NullString{String: in.Email, Valid: in.RegisterType == user.RegisterType_REGISTER_TYPE_EMAIL && in.Email != ""},
		CreateTimes: now,
		UpdateTimes: now,
	}

	userId := int64(0)
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		userModel := models.NewTUserModel(conn, l.svcCtx.Config.CacheRedis)
		userIdentityModel := models.NewTUserIdentityModel(conn, l.svcCtx.Config.CacheRedis)

		result, err := userModel.Insert(ctx, tuser)
		if err != nil {
			return err
		}

		userId, err = result.LastInsertId()
		if err != nil {
			return err
		}

		identity.UserId = userId
		if _, err := userIdentityModel.Insert(ctx, identity); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	str := make(map[string]any, 0)
	str["tid"] = tenant.Data.Id
	expand, err := json.Marshal(str)
	if err != nil {
		return nil, err
	}
	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		l.svcCtx.Config.Jwt.AccessExpire,
		userId, "", string(expand),
	)
	if err != nil {
		return nil, err
	}

	return &user.RegisterResp{
		Base:    helper.OkResp(),
		UserId:  userId,
		Token:   token,
		Profile: &user.UserProfile{},
	}, nil
}

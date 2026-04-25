package logic

import (
	"context"
	"database/sql"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type SysTenantUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysTenantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysTenantUpdateLogic {
	return &SysTenantUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新租户
func (l *SysTenantUpdateLogic) SysTenantUpdate(in *system.SysTenantUpdateReq) (*system.RespBase, error) {
	var tenantNotFound bool
	err := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		tenantModel := models.NewSysTenantModel(conn, l.svcCtx.Config.CacheRedis).(models.TenantModel)
		userModel := models.NewSysUserModel(conn, l.svcCtx.Config.CacheRedis).(models.UserModel)

		tenant, err := tenantModel.FindOne(ctx, in.Id)
		if err != nil {
			return err
		}
		if tenant == nil {
			tenantNotFound = true
			return nil
		}
		if in.TenantName != "" {
			tenant.TenantName = in.TenantName
		}
		if in.TenantPassword != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.TenantPassword), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			user, err := userModel.FindOneByTenantIdUsername(ctx, tenant.Id, "")
			if err != nil {
				return err
			}
			if user != nil {
				user.Password = string(hashedPassword)
			}
			err = userModel.Update(ctx, user)
			if err != nil {
				return err
			}
		}
		if in.Status != 0 {
			tenant.Status = commonStatusToModel(in.Status)
		}
		if in.ExpireTime != 0 {
			tenant.ExpireTime = in.ExpireTime
		}
		if in.ContactName != "" {
			tenant.ContactName = sql.NullString{String: in.ContactName, Valid: true}
		}
		if in.ContactPhone != "" {
			tenant.ContactPhone = sql.NullString{String: in.ContactPhone, Valid: true}
		}
		if in.Remark != "" {
			tenant.Remark = sql.NullString{String: in.Remark, Valid: true}
		}
		tenant.UpdateTimes = utils.NowMillis()
		return tenantModel.Update(ctx, tenant)
	})
	if err != nil {
		return nil, err
	}
	if tenantNotFound {
		return &system.RespBase{
			Base: helper.GetErrResp(400, i18n.Translate(i18n.TenantNotFound, l.ctx)),
		}, nil
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

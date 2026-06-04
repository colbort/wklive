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
)

type UpdateIdentityLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateIdentityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateIdentityLogic {
	return &UpdateIdentityLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改实名认证信息
func (l *UpdateIdentityLogic) UpdateIdentity(in *user.UpdateIdentityReq) (*user.UpdateIdentityResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	// 获取用户信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.UpdateIdentityResp{
			Base: helper.GetErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	now := utils.NowMillis()
	identity, err := l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, userId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if identity != nil {
		// 更新现有身份信息
		changed := false
		if in.Phone != "" {
			identity.Phone = sql.NullString{String: in.Phone, Valid: true}
			changed = true
		}
		if in.Email != "" {
			identity.Email = sql.NullString{String: in.Email, Valid: true}
			changed = true
		}
		if in.RealName != "" {
			identity.RealName = sql.NullString{String: in.RealName, Valid: true}
			changed = true
		}
		if in.Gender != 0 {
			identity.Gender = int64(in.Gender)
			changed = true
		}
		if in.Birthday != 0 {
			identity.Birthday = in.Birthday
			changed = true
		}
		if in.CountryCode != "" {
			identity.CountryCode = sql.NullString{String: in.CountryCode, Valid: true}
			changed = true
		}
		if in.Province != "" {
			identity.Province = sql.NullString{String: in.Province, Valid: true}
			changed = true
		}
		if in.City != "" {
			identity.City = sql.NullString{String: in.City, Valid: true}
			changed = true
		}
		if in.Address != "" {
			identity.Address = sql.NullString{String: in.Address, Valid: true}
			changed = true
		}

		if changed {
			identity.UpdateTimes = now
			err = l.svcCtx.UserIdentityModel.Update(l.ctx, identity)
			if err != nil {
				return nil, err
			}
		}
	} else {
		return nil, i18n.StatusError(l.ctx, i18n.UserIdentityInfoNotFound)
	}

	l.Logger.Infof("用户 %d 更新实名认证信息成功，状态为待审核", userId)

	return &user.UpdateIdentityResp{
		Base:     helper.OkResp(),
		Identity: toUserIdentityProto(identity),
	}, nil
}

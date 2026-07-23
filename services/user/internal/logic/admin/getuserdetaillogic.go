package adminlogic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type GetUserDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDetailLogic {
	return &GetUserDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员获取用户详情（聚合查询）
func (l *GetUserDetailLogic) GetUserDetail(in *user.GetUserDetailReq) (*user.GetUserDetailResp, error) {
	// 获取用户基本信息
	tuser, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	if tuser == nil {
		return &user.GetUserDetailResp{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}

	var identity *models.TUserIdentity
	var security *models.TUserSecurity
	var userBanks []*models.TUserBank
	err = mr.Finish(
		func() error {
			var queryErr error
			identity, queryErr = l.svcCtx.UserIdentityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			security, queryErr = l.svcCtx.UserSecurityModel.FindOneByTenantIdUserId(l.ctx, tuser.TenantId, in.UserId)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
		func() error {
			var queryErr error
			userBanks, _, queryErr = l.svcCtx.UserBankModel.FindPage(l.ctx, models.UserBankPageFilter{
				TenantId: tuser.TenantId, UserId: in.UserId,
			}, 0, 100)
			if errors.Is(queryErr, models.ErrNotFound) {
				return nil
			}
			return queryErr
		},
	)
	if err != nil {
		return nil, err
	}
	return &user.GetUserDetailResp{
		Base: helper.OkResp(),
		Data: toUserDetailProto(tuser, identity, security, userBanks),
	}, nil
}

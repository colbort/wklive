// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUsersLogic) ListUsers(req *types.ListUsersReq) (resp *types.ListUsersResp, err error) {
	result, err := l.svcCtx.UserCli.ListUsers(l.ctx, &user.ListUsersReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		TenantId:          req.TenantId,
		TenantCode:        req.TenantCode,
		Keyword:           req.Keyword,
		UserId:            req.UserId,
		UserNo:            req.UserNo,
		Username:          req.Username,
		Phone:             req.Phone,
		Email:             req.Email,
		Status:            user.UserStatus(req.Status),
		MemberLevel:       req.MemberLevel,
		VerifyStatus:      user.VerifyStatus(req.VerifyStatus),
		KycLevel:          user.KycLevel(req.KycLevel),
		InviteCode:        req.InviteCode,
		RegisterTimeStart: req.RegisterTimeStart,
		RegisterTimeEnd:   req.RegisterTimeEnd,
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.UserItem, len(result.List))
	for i, item := range result.List {
		data[i] = types.UserItem{
			Id:             item.Id,
			TenantId:       item.TenantId,
			UserNo:         item.UserNo,
			Username:       item.Username,
			Nickname:       item.Nickname,
			Avatar:         item.Avatar,
			PasswordHash:   item.PasswordHash,
			RegisterType:   int64(item.RegisterType),
			Status:         int64(item.Status),
			MemberLevel:    item.MemberLevel,
			Language:       item.Language,
			Timezone:       item.Timezone,
			InviteCode:     item.InviteCode,
			Signature:      item.Signature,
			Source:         item.Source,
			ReferrerUserId: item.ReferrerUserId,
			LastLoginIp:    item.LastLoginIp,
			LastLoginTime:  item.LastLoginTime,
			RegisterIp:     item.RegisterIp,
			RegisterTime:   item.RegisterTime,
			IsGuest:        item.IsGuest,
			IsRecharge:     item.IsRecharge,
			DeviceId:       item.DeviceId,
			Fingerprint:    item.Fingerprint,
			Remark:         item.Remark,
			Deleted:        item.Deleted,
			CreateTimes:    item.CreateTimes,
			UpdateTimes:    item.UpdateTimes,
		}
	}

	return &types.ListUsersResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: data,
	}, nil
}

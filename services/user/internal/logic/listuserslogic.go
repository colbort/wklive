package logic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUsersLogic {
	return &ListUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员查询用户列表
func (l *ListUsersLogic) ListUsers(in *user.ListUsersReq) (*user.ListUsersResp, error) {
	items, total, err := l.svcCtx.UserModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	userList := make([]*user.UserItem, 0, len(items))
	for _, item := range items {
		userList = append(userList, &user.UserItem{
			Id:             item.Id,
			TenantId:       item.TenantId,
			UserNo:         item.UserNo,
			Username:       item.Username,
			Nickname:       item.Nickname.String,
			Avatar:         item.Avatar.String,
			PasswordHash:   item.PasswordHash,
			RegisterType:   user.RegisterType(item.RegisterType),
			Status:         user.UserStatus(item.Status),
			MemberLevel:    item.MemberLevel,
			Language:       item.Language.String,
			Timezone:       item.Timezone.String,
			InviteCode:     item.InviteCode.String,
			Signature:      item.Signature.String,
			Source:         item.Source.String,
			ReferrerUserId: item.ReferrerUserId.Int64,
			LastLoginIp:    item.LastLoginIp.String,
			LastLoginTime:  item.LastLoginTime,
			RegisterIp:     item.RegisterIp.String,
			RegisterTime:   item.RegisterTime,
			IsGuest:        item.IsGuest,
			IsRecharge:     item.IsRecharge,
			DeviceId:       item.DeviceId,
			Fingerprint:    item.Fingerprint,
			Remark:         item.Remark.String,
			Deleted:        item.Deleted,
			CreateTimes:    item.CreateTimes,
			UpdateTimes:    item.UpdateTimes,
		})
	}

	return &user.ListUsersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		List: userList,
	}, nil
}

func matchUserFilters(in *user.ListUsersReq, item *models.TUser, identity *models.TUserIdentity) bool {
	if in.UserId != 0 && item.Id != in.UserId {
		return false
	}
	if in.UserNo != "" && item.UserNo != in.UserNo {
		return false
	}
	if in.Username != "" && item.Username != in.Username {
		return false
	}
	if in.Status != 0 && item.Status != int64(in.Status) {
		return false
	}
	if in.MemberLevel != 0 && item.MemberLevel != int64(in.MemberLevel) {
		return false
	}
	if in.InviteCode != "" && item.InviteCode.String != in.InviteCode {
		return false
	}
	if in.RegisterTimeStart != 0 && item.RegisterTime < in.RegisterTimeStart {
		return false
	}
	if in.RegisterTimeEnd != 0 && item.RegisterTime > in.RegisterTimeEnd {
		return false
	}
	if in.Phone != "" && identityString(identity, func(i *models.TUserIdentity) string { return i.Phone.String }) != in.Phone {
		return false
	}
	if in.Email != "" && identityString(identity, func(i *models.TUserIdentity) string { return i.Email.String }) != in.Email {
		return false
	}
	if in.VerifyStatus != 0 && identityInt(identity, func(i *models.TUserIdentity) int64 { return i.VerifyStatus }) != int64(in.VerifyStatus) {
		return false
	}
	if in.KycLevel != 0 && identityInt(identity, func(i *models.TUserIdentity) int64 { return i.KycLevel }) != int64(in.KycLevel) {
		return false
	}
	if in.Keyword != "" {
		keyword := strings.ToLower(in.Keyword)
		if !strings.Contains(strings.ToLower(item.Username), keyword) &&
			!strings.Contains(strings.ToLower(item.UserNo), keyword) &&
			!strings.Contains(strings.ToLower(item.Nickname.String), keyword) &&
			!strings.Contains(strings.ToLower(identityString(identity, func(i *models.TUserIdentity) string { return i.Phone.String })), keyword) &&
			!strings.Contains(strings.ToLower(identityString(identity, func(i *models.TUserIdentity) string { return i.Email.String })), keyword) {
			return false
		}
	}
	return true
}

func identityString(identity *models.TUserIdentity, getter func(*models.TUserIdentity) string) string {
	if identity == nil {
		return ""
	}
	return getter(identity)
}

func identityInt(identity *models.TUserIdentity, getter func(*models.TUserIdentity) int64) int64 {
	if identity == nil {
		return 0
	}
	return getter(identity)
}

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

type ListUserIdentitiesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListUserIdentitiesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserIdentitiesLogic {
	return &ListUserIdentitiesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 管理员查询用户实名认证信息列表
func (l *ListUserIdentitiesLogic) ListUserIdentities(in *user.ListUserIdentitiesReq) (*user.ListUserIdentitiesResp, error) {
	items, total, err := l.svcCtx.UserIdentityModel.FindPage(l.ctx, in.TenantId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	identityList := make([]*user.UserIdentityListItem, 0, len(items))
	for _, item := range items {
		u, err := l.svcCtx.UserModel.FindOne(l.ctx, item.UserId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if u == nil || !matchIdentityFilters(in, item, u) {
			continue
		}

		identityList = append(identityList, &user.UserIdentityListItem{
			UserId:       item.UserId,
			UserNo:       u.UserNo,
			Username:     u.Username,
			Phone:        item.Phone.String,
			Email:        item.Email.String,
			RealName:     item.RealName.String,
			IdType:       user.IdType(item.IdType),
			IdNo:         item.IdNo.String,
			KycLevel:     user.KycLevel(item.KycLevel),
			VerifyStatus: user.VerifyStatus(item.VerifyStatus),
			RejectReason: item.RejectReason.String,
			SubmitTime:   item.SubmitTime,
			VerifyTime:   item.VerifyTime,
			VerifyBy:     item.VerifyBy.Int64,
		})
	}

	return &user.ListUserIdentitiesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		List: identityList,
	}, nil
}

func matchIdentityFilters(in *user.ListUserIdentitiesReq, item *models.TUserIdentity, u *models.TUser) bool {
	if in.UserId != 0 && item.UserId != in.UserId {
		return false
	}
	if in.UserNo != "" && u.UserNo != in.UserNo {
		return false
	}
	if in.Username != "" && u.Username != in.Username {
		return false
	}
	if in.Phone != "" && item.Phone.String != in.Phone {
		return false
	}
	if in.Email != "" && item.Email.String != in.Email {
		return false
	}
	if in.RealName != "" && item.RealName.String != in.RealName {
		return false
	}
	if in.VerifyStatus != 0 && item.VerifyStatus != int64(in.VerifyStatus) {
		return false
	}
	if in.KycLevel != 0 && item.KycLevel != int64(in.KycLevel) {
		return false
	}
	if in.IdType != 0 && item.IdType != int64(in.IdType) {
		return false
	}
	if in.Keyword != "" {
		keyword := strings.ToLower(in.Keyword)
		if !strings.Contains(strings.ToLower(u.UserNo), keyword) &&
			!strings.Contains(strings.ToLower(u.Username), keyword) &&
			!strings.Contains(strings.ToLower(item.Phone.String), keyword) &&
			!strings.Contains(strings.ToLower(item.Email.String), keyword) &&
			!strings.Contains(strings.ToLower(item.RealName.String), keyword) {
			return false
		}
	}
	return true
}

package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/common/utils"
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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	items, total, err := l.svcCtx.UserModel.FindPage(l.ctx, models.UserPageFilter{
		TenantId:          in.TenantId,
		UserId:            in.UserId,
		UserNo:            in.UserNo,
		Username:          in.Username,
		Nickname:          in.Nickname,
		Phone:             in.Phone,
		Email:             in.Email,
		Status:            int64(in.Status),
		MemberLevel:       int64(in.MemberLevel),
		VerifyStatus:      int64(in.VerifyStatus),
		KycLevel:          int64(in.KycLevel),
		InviteCode:        in.InviteCode,
		RegisterTimeStart: in.RegisterTimeStart,
		RegisterTimeEnd:   in.RegisterTimeEnd,
		Keyword:           in.Keyword,
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*user.UserItem, 0, len(items))
	for _, item := range items {
		data = append(data, toUserItemProto(item))
	}

	return &user.ListUsersResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}

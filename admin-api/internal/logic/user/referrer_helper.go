package user

import (
	"context"
	"strings"

	"wklive/admin-api/internal/svc"
	"wklive/proto/common"
	"wklive/proto/user"
)

func resolveReferrerByInviteCode(svcCtx *svc.ServiceContext, ctx context.Context, tenantId int64, inviteCode string) (*user.UserItem, error) {
	code := strings.TrimSpace(inviteCode)
	if code == "" {
		return nil, nil
	}

	if tenantId <= 0 {
		result, err := svcCtx.UserCli.ListUsers(ctx, &user.ListUsersReq{
			Page: &common.PageReq{
				Cursor: 0,
				Limit:  1,
			},
			InviteCode: code,
		})
		if err != nil {
			return nil, err
		}
		if result.Base.Code != 200 || len(result.List) == 0 {
			return nil, nil
		}
		return result.List[0], nil
	}

	result, err := svcCtx.UserCli.ListUsers(ctx, &user.ListUsersReq{
		Page: &common.PageReq{
			Cursor: 0,
			Limit:  1,
		},
		TenantId:   tenantId,
		InviteCode: code,
	})
	if err != nil {
		return nil, err
	}
	if result.Base.Code != 200 || len(result.List) == 0 {
		return nil, nil
	}

	return result.List[0], nil
}

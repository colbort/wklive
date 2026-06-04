package user

import (
	"context"
	"strings"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/user"
)

func resolveReferrerByInviteCode(svcCtx *svc.ServiceContext, ctx context.Context, tenantId int64, inviteCode string) (*types.UserItem, error) {
	code := strings.TrimSpace(inviteCode)
	if code == "" {
		return nil, nil
	}

	req := &user.ListUsersReq{
		Page: &common.PageReq{
			Cursor: 0,
			Limit:  1,
		},
		InviteCode: code,
	}
	if tenantId > 0 {
		req.TenantId = tenantId
	}

	result, err := logicutil.Proxy[types.ListUsersResp](ctx, req, svcCtx.UserCli.ListUsers)
	if err != nil {
		return nil, err
	}
	if result.Code != 200 || len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

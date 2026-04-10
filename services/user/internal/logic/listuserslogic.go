package logic

import (
	"context"

	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

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
	// TODO: 实现复杂查询逻辑
	// 需要支持多个过滤条件：keyword, user_id, user_no, username, phone, email, status, verify_status, kyc_level等
	// 使用 FindPage 或自定义查询方法

	userList := []*user.UserListItem{}

	return &user.ListUsersResp{
		Base: &common.RespBase{
			Code:  200,
			Msg:   "OK",
			Total: int64(len(userList)),
		},
		List: userList,
	}, nil
}
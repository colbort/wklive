package logic

import (
	"context"

	"wklive/proto/user"
	"wklive/services/user/internal/svc"

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

// 实名认证相关接口
func (l *ListUserIdentitiesLogic) ListUserIdentities(in *user.ListUserIdentitiesReq) (*user.ListUserIdentitiesResp, error) {
	// todo: add your logic here and delete this line

	return &user.ListUserIdentitiesResp{}, nil
}

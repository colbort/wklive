package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserLogic {
	return &DeleteUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除用户
func (l *DeleteUserLogic) DeleteUser(in *user.DeleteUserReq) (*user.AdminCommonResp, error) {
	tuser, err := l.svcCtx.UserModel.FindByTenantIdUserId(l.ctx, in.TenantId, in.UserId)
	if err != nil && errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if tuser == nil {
		return &user.AdminCommonResp{
			Base: helper.FailWithCode(401),
		}, nil
	}
	err = l.svcCtx.UserModel.Delete(l.ctx, tuser.Id)
	if err != nil {
		return nil, err
	}
	return &user.AdminCommonResp{
		Base: helper.OkResp(),
	}, nil
}

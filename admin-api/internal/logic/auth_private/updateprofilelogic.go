// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/system"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateProfileLogic) UpdateProfile(req *types.UpdateProfileReq) (resp *types.RespBase, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	result, err := l.svcCtx.SystemCli.UpdateProfile(l.ctx, &system.UpdateProfileReq{
		Id:       userId,
		Avatar:   &req.Avatar,
		Nickname: &req.Nickname,
		Password: &req.Password,
	})
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.InternalServerError)
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

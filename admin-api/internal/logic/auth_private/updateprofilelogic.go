// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package auth_private

import (
	"context"
	"fmt"

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
	uid, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", i18n.Translate(i18n.InternalServerError, l.ctx), err)
	}
	result, err := l.svcCtx.SystemCli.UpdateProfile(l.ctx, &system.UpdateProfileReq{
		Id:       uid,
		Avatar:   &req.Avatar,
		Nickname: &req.Nickname,
		Password: &req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateUserLevelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserLevelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLevelLogic {
	return &UpdateUserLevelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserLevelLogic) UpdateUserLevel(req *types.UpdateUserLevelReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.UpdateUserLevel(l.ctx, &user.UpdateUserLevelReq{
		TenantId:    req.TenantId,
		UserId:      req.UserId,
		MemberLevel: req.MemberLevel,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

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

type ResetLoginPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetLoginPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetLoginPasswordLogic {
	return &ResetLoginPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetLoginPasswordLogic) ResetLoginPassword(req *types.ResetLoginPasswordReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.ResetLoginPassword(l.ctx, &user.ResetLoginPasswordReq{
		TenantId:    req.TenantId,
		UserId:      req.UserId,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

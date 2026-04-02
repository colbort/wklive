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

type ResetPayPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPayPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPayPasswordLogic {
	return &ResetPayPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPayPasswordLogic) ResetPayPassword(req *types.ResetPayPasswordReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.ResetPayPassword(l.ctx, &user.ResetPayPasswordReq{
		TenantId: req.TenantId,
		UserId:   req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

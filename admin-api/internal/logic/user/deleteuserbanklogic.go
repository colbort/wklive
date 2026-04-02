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

type DeleteUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteUserBankLogic {
	return &DeleteUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteUserBankLogic) DeleteUserBank(req *types.DeleteUserBankReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.DeleteUserBank(l.ctx, &user.DeleteUserBankReq{
		TenantId: req.TenantId,
		Id:       req.Id,
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

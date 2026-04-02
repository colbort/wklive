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

type SetDefaultUserBankLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetDefaultUserBankLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetDefaultUserBankLogic {
	return &SetDefaultUserBankLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetDefaultUserBankLogic) SetDefaultUserBank(req *types.SetDefaultUserBankReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.SetDefaultUserBank(l.ctx, &user.SetDefaultUserBankReq{
		TenantId: req.TenantId,
		Id:       req.Id,
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

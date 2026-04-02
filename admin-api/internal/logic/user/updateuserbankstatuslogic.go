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

type UpdateUserBankStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateUserBankStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserBankStatusLogic {
	return &UpdateUserBankStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateUserBankStatusLogic) UpdateUserBankStatus(req *types.UpdateUserBankStatusReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.UserCli.UpdateUserBankStatus(l.ctx, &user.UpdateUserBankStatusReq{
		TenantId: req.TenantId,
		Id:       req.Id,
		Status:   user.BankStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

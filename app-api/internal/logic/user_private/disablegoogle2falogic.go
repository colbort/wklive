// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_private

import (
	"context"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableGoogle2FALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDisableGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableGoogle2FALogic {
	return &DisableGoogle2FALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DisableGoogle2FALogic) DisableGoogle2FA(req *types.VerifyGoogle2FAReq) (resp *types.RespBase, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.DisableGoogle2FA(l.ctx, &user.DisableGoogle2FAReq{
		UserId:     userId,
		GoogleCode: req.GoogleCode,
	})
	if err != nil {
		return nil, err
	}
	return &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}, nil
}

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

type InitGoogle2FALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInitGoogle2FALogic(ctx context.Context, svcCtx *svc.ServiceContext) *InitGoogle2FALogic {
	return &InitGoogle2FALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InitGoogle2FALogic) InitGoogle2FA() (resp *types.InitGoogle2FAResp, err error) {
	userId, err := utils.GetUidFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.UserCli.InitGoogle2FA(l.ctx, &user.InitGoogle2FAReq{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}
	return &types.InitGoogle2FAResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		QrCodeUrl: result.QrCodeUrl,
		Secret:    result.Secret,
	}, nil
}

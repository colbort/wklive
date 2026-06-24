// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"context"

	"chat-admin-api/internal/logicutil"
	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OptionsLogic {
	return &OptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OptionsLogic) Options() (*types.ChatAdminOptionsResp, error) {
	return &types.ChatAdminOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "ok"},
		Data: types.ChatAdminOptions{
			Options: logicutil.CoreOptions(),
		},
	}, nil
}

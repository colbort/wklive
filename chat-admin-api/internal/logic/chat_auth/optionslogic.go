// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_auth

import (
	"context"

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
			AgentStatuses: []types.OptionItem{
				{Key: "chat.agent.status.offline", Value: 1, TagType: "info"},
				{Key: "chat.agent.status.online", Value: 2, TagType: "success"},
				{Key: "chat.agent.status.busy", Value: 3, TagType: "warning"},
				{Key: "chat.agent.status.resting", Value: 4, TagType: "info"},
			},
		},
	}, nil
}

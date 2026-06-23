// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"wklive/proto/chat"
	"wklive/proto/common"

	"chat-api/internal/jwt"
	"chat-api/internal/logicutil"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatMessagesLogic {
	return &ListMyChatMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyChatMessagesLogic) ListMyChatMessages(req *types.ListChatMessagesReq) (resp *types.ListChatMessagesResp, err error) {
	claims, ok := jwt.ClaimsFromContext(l.ctx)
	if !ok || claims.SessionNo == "" {
		return &types.ListChatMessagesResp{
			RespBase: types.RespBase{Code: 200, Msg: "ok"},
			Data:     []types.ChatMessage{},
		}, nil
	}

	proxyReq := chat.ListMyChatMessagesReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		SessionNo: claims.SessionNo,
	}
	return logicutil.Proxy[types.ListChatMessagesResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.ListMyChatMessages)
}

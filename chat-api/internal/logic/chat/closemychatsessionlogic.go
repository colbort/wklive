// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"strings"
	"wklive/proto/chat"

	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseMyChatSessionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCloseMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseMyChatSessionLogic {
	return &CloseMyChatSessionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseMyChatSessionLogic) CloseMyChatSession(req *types.CloseMyChatSessionReq) (*types.RespBase, error) {
	sessionNo := strings.TrimSpace(req.SessionNo)
	if sessionNo == "" {
		return &types.RespBase{Code: 400, Msg: "sessionNo is required"}, nil
	}

	resp, err := l.svcCtx.ChatAppCli.CloseMyChatSession(l.ctx, &chat.CloseMyChatSessionReq{
		SessionNo:   sessionNo,
		CloseReason: strings.TrimSpace(req.CloseReason),
	})
	if err != nil {
		return nil, err
	}
	if resp.GetBase() == nil {
		return &types.RespBase{Code: 100001, Msg: "empty response"}, nil
	}
	return &types.RespBase{
		Code: resp.GetBase().GetCode(),
		Msg:  resp.GetBase().GetMsg(),
	}, nil
}

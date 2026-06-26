// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"
	"wklive/proto/chat"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAgentMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendAgentMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAgentMessageLogic {
	return &SendAgentMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendAgentMessageLogic) SendAgentMessage(req *types.SendAgentMessageReq) (resp *types.ChatMessageResp, err error) {
	rpcResp, err := l.svcCtx.ChatAdminCli.SendAgentMessage(l.ctx, &chat.SendAgentMessageReq{
		SessionNo:       req.SessionNo,
		ClientMessageId: req.ClientMsgNo,
		MessageType:     chat.ChatMessageType(req.MessageType),
		Content:         req.Content,
		Url:             req.Url,
		FileName:        req.FileName,
		MimeType:        req.MimeType,
		FileSize:        req.FileSize,
		Width:           int32(req.Width),
		Height:          int32(req.Height),
		Duration:        int32(req.Duration),
		Extra:           req.Extra,
	})
	if err != nil {
		return nil, err
	}
	return &types.ChatMessageResp{
		RespBase: respBaseToType(rpcResp.GetBase()),
		Data:     protoMessageToType(req.MerchantId, rpcResp.GetData()),
	}, nil
}

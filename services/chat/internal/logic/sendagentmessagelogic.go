package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendAgentMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendAgentMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAgentMessageLogic {
	return &SendAgentMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送客服消息
func (l *SendAgentMessageLogic) SendAgentMessage(in *chat.SendAgentMessageReq) (*chat.AdminChatMessageResp, error) {
	msg, err := ih.SendMessage(l.ctx, l.svcCtx, ih.SendMessageOptions{
		MerchantId:     in.GetMerchantId(),
		SessionNo:      in.GetSessionNo(),
		IsGuest:        in.GetIsGuest(),
		Sender:         in.GetSender(),
		Receiver:       in.GetReceiver(),
		MessageType:    in.GetMessageType(),
		Content:        in.GetContent(),
		Url:            in.GetUrl(),
		FileName:       in.GetFileName(),
		MimeType:       in.GetMimeType(),
		FileSize:       in.GetFileSize(),
		Duration:       in.GetDuration(),
		ReceiveChannel: chat.ChatAppEventChannel,
		ReceiptChannel: chat.ChatAdminEventChannel,
	})
	if err != nil {
		return &chat.AdminChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatMessageResp{Base: helper.OkResp(), Data: msg}, nil
}

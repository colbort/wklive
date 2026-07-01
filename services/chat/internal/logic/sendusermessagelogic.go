package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendUserMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendUserMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendUserMessageLogic {
	return &SendUserMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送用户消息
func (l *SendUserMessageLogic) SendUserMessage(in *chat.SendUserMessageReq) (*chat.AppChatMessageResp, error) {
	msg, base, err := ih.SendMessage(l.ctx, l.svcCtx, ih.SendMessageOptions{
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
		MessageChannel: chat.ChatAdminEventChannel,
		ReceiptChannel: chat.ChatAppEventChannel,
	})
	if base != nil {
		return &chat.AppChatMessageResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg}, nil
}

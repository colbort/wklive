package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatMessagesLogic {
	return &PageChatMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话消息
func (l *PageChatMessagesLogic) PageChatMessages(in *chat.PageChatMessagesReq) (*chat.PageChatMessagesResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatMessagesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatMessagesResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.PageChatMessagesResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.PageChatMessagesResp{Base: base}, nil
	}

	cursor, limit := pageInput(in.GetPage())
	model := l.svcCtx.ChatMessageFactory.New(merchantID)
	if model == nil {
		return &chat.PageChatMessagesResp{Base: badBase("invalid merchant_id")}, nil
	}
	list, err := model.FindPage(l.ctx, models.ChatMessagePageFilter{
		MerchantId: merchantID,
		SessionNo:  session.SessionNo,
		SenderType: int64(in.GetSenderType()),
		BeforeTime: cursor,
	}, limit)
	if err != nil {
		return &chat.PageChatMessagesResp{Base: errorBase(err)}, nil
	}
	nextCursor := messageNextCursor(list)
	base = helper.OkWithOthers(0, int64(len(list)) == limit && nextCursor > 0, cursor > 0, nextCursor, cursor)
	return &chat.PageChatMessagesResp{Base: base, Data: toProtoMessages(list)}, nil
}

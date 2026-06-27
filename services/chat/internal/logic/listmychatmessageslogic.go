package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/pageutil"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatMessagesLogic {
	return &ListMyChatMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询会话消息
func (l *ListMyChatMessagesLogic) ListMyChatMessages(in *chat.ListMyChatMessagesReq) (*chat.ListChatMessagesResp, error) {
	merchantID, userID, base, err := internal.ChatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.ListChatMessagesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	session, base, err := internal.GetSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.ListChatMessagesResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}

	cursor, limit := pageutil.Input(in.GetPage())
	model := l.svcCtx.ChatMessageFactory.New(merchantID)
	if model == nil {
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(400, "invalid merchant_id")}, nil
	}
	list, err := model.FindPage(l.ctx, models.ChatMessagePageFilter{
		MerchantId: merchantID,
		SessionNo:  session.SessionNo,
		BeforeTime: cursor,
	}, limit)
	if err != nil {
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	nextCursor := internal.MessageNextCursor(list)
	base = helper.OkWithOthers(0, int64(len(list)) == limit && nextCursor > 0, cursor > 0, nextCursor, cursor)
	return &chat.ListChatMessagesResp{Base: base, Data: internal.ToProtoMessages(list)}, nil
}

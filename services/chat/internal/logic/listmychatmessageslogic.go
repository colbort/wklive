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
func (l *ListMyChatMessagesLogic) ListMyChatMessages(in *chat.ListMyChatMessagesReq) (*chat.AppListChatMessagesResp, error) {
	if in.GetIsGuest() || in.GetMerchantId() > 0 {
		var cursor, limit int64
		if in.GetPage() != nil {
			cursor = in.GetPage().GetCursor()
			limit = in.GetPage().GetLimit()
		}
		list, hasNext, nextCursor, err := internal.ListTransientMessages(
			l.ctx,
			l.svcCtx.BusRedis,
			in.GetMerchantId(),
			in.GetSessionNo(),
			int64(in.GetSenderType()),
			cursor,
			limit,
		)
		if err != nil {
			return &chat.AppListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		return &chat.AppListChatMessagesResp{
			Base: helper.OkWithOthers(0, hasNext, cursor > 0, nextCursor, cursor),
			Data: list,
		}, nil
	}

	merchantID, userID, base, err := internal.ChatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.AppListChatMessagesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AppListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	session, base, err := internal.GetSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo(), false)
	if err != nil {
		return &chat.AppListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.AppListChatMessagesResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.AppListChatMessagesResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}

	cursor, limit := pageutil.Input(in.GetPage())
	model := l.svcCtx.ChatMessageFactory.New(merchantID)
	if model == nil {
		return &chat.AppListChatMessagesResp{Base: helper.ErrResp(400, "invalid merchant_id")}, nil
	}
	list, err := model.FindPage(l.ctx, models.ChatMessagePageFilter{
		MerchantId: merchantID,
		SessionNo:  session.SessionNo,
		BeforeTime: cursor,
	}, limit)
	if err != nil {
		return &chat.AppListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	nextCursor := internal.MessageNextCursor(list)
	base = helper.OkWithOthers(0, int64(len(list)) == limit && nextCursor > 0, cursor > 0, nextCursor, cursor)
	return &chat.AppListChatMessagesResp{Base: base, Data: internal.ToProtoMessages(list)}, nil
}

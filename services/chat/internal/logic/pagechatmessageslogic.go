package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/pageutil"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
	if in.GetIsGuest() || in.GetMerchantId() > 0 {
		var cursor, limit int64
		if in.GetPage() != nil {
			cursor = in.GetPage().GetCursor()
			limit = in.GetPage().GetLimit()
		}
		list, hasNext, nextCursor, err := ih.ListTransientMessages(
			l.ctx,
			l.svcCtx.BusRedis,
			in.GetMerchantId(),
			in.GetSessionNo(),
			int64(in.GetSenderType()),
			cursor,
			limit,
		)
		if err != nil {
			return &chat.PageChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		return &chat.PageChatMessagesResp{
			Base: helper.OkWithOthers(0, hasNext, cursor > 0, nextCursor, cursor),
			Data: list,
		}, nil
	}

	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.PageChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	session, err := ih.GetSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo(), false)
	if err != nil {
		return &chat.PageChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	cursor, limit := pageutil.Input(in.GetPage())
	model := l.svcCtx.ChatMessageFactory.New(merchantID)
	if model == nil {
		return &chat.PageChatMessagesResp{Base: helper.ErrResp(400, "invalid merchant_id")}, nil
	}
	list, err := model.FindPage(l.ctx, models.ChatMessagePageFilter{
		MerchantId: merchantID,
		SessionNo:  session.SessionNo,
		SenderType: int64(in.GetSenderType()),
		BeforeTime: cursor,
	}, limit)
	if err != nil {
		return &chat.PageChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	nextCursor := ih.MessageNextCursor(list)
	base := helper.OkWithOthers(0, int64(len(list)) == limit && nextCursor > 0, cursor > 0, nextCursor, cursor)
	return &chat.PageChatMessagesResp{Base: base, Data: ih.ToProtoMessages(list)}, nil
}

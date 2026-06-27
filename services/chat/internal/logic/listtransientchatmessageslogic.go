package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTransientChatMessagesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTransientChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTransientChatMessagesLogic {
	return &ListTransientChatMessagesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询游客临时消息
func (l *ListTransientChatMessagesLogic) ListTransientChatMessages(in *chat.ListTransientChatMessagesReq) (*chat.ListChatMessagesResp, error) {
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
		return &chat.ListChatMessagesResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.ListChatMessagesResp{
		Base: helper.OkWithOthers(0, hasNext, cursor > 0, nextCursor, cursor),
		Data: list,
	}, nil
}

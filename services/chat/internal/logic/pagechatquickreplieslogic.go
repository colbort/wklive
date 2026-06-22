package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatQuickRepliesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatQuickRepliesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatQuickRepliesLogic {
	return &PageChatQuickRepliesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询快捷回复
func (l *PageChatQuickRepliesLogic) PageChatQuickReplies(in *chat.PageChatQuickRepliesReq) (*chat.PageChatQuickRepliesResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatQuickRepliesResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatQuickRepliesResp{Base: errorBase(err)}, nil
	}
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatQuickReplyModel.FindPage(l.ctx, models.ChatQuickReplyPageFilter{
		MerchantId: merchantID,
		AgentId:    in.GetAgentId(),
		CategoryId: in.GetCategoryId(),
		Enabled:    int64(in.GetEnabled()),
		Keyword:    strings.TrimSpace(in.GetKeyword()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatQuickRepliesResp{Base: errorBase(err)}, nil
	}
	return &chat.PageChatQuickRepliesResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoChatQuickReplies(list),
	}, nil
}

package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatGroupsLogic {
	return &PageChatGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询客服分组
func (l *PageChatGroupsLogic) PageChatGroups(in *chat.PageChatGroupsReq) (*chat.PageChatGroupsResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatGroupsResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatGroupsResp{Base: errorBase(err)}, nil
	}
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatGroupModel.FindPage(l.ctx, models.ChatGroupPageFilter{
		MerchantId: merchantID,
		Keyword:    strings.TrimSpace(in.GetKeyword()),
		Enabled:    int64(in.GetEnabled()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatGroupsResp{Base: errorBase(err)}, nil
	}
	return &chat.PageChatGroupsResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoChatGroups(list),
	}, nil
}

package logic

import (
	"context"
	"strings"
	"wklive/common/helper"
	"wklive/common/pageutil"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.PageChatGroupsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	list, total, err := l.svcCtx.ChatGroupModel.FindPage(l.ctx, models.ChatGroupPageFilter{
		MerchantId: merchantID,
		Keyword:    strings.TrimSpace(in.GetKeyword()),
		Enabled:    int64(in.GetEnabled()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatGroupsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.PageChatGroupsResp{
		Base: ih.OffsetBase(cursor, limit, len(list), total),
		Data: ih.ToProtoChatGroups(list),
	}, nil
}

package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListMyChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatSessionsLogic {
	return &ListMyChatSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的会话列表
func (l *ListMyChatSessionsLogic) ListMyChatSessions(in *chat.ListMyChatSessionsReq) (*chat.ListChatSessionsResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.ListChatSessionsResp{Base: base}, nil
	}
	if err != nil {
		return &chat.ListChatSessionsResp{Base: errorBase(err)}, nil
	}
	if err := validateMerchantUser(merchantID, userID); err != nil {
		return &chat.ListChatSessionsResp{Base: badBase(err.Error())}, nil
	}
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatSessionModel.FindPage(l.ctx, models.ChatSessionPageFilter{
		MerchantId: merchantID,
		UserId:     userID,
		Status:     int64(in.GetStatus()),
	}, cursor, limit)
	if err != nil {
		return &chat.ListChatSessionsResp{Base: errorBase(err)}, nil
	}
	return &chat.ListChatSessionsResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoSessions(list),
	}, nil
}

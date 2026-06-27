package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppPageTransientChatSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppPageTransientChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPageTransientChatSessionsLogic {
	return &AppPageTransientChatSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询游客临时会话
func (l *AppPageTransientChatSessionsLogic) AppPageTransientChatSessions(in *chat.AppPageTransientChatSessionsReq) (*chat.AppPageTransientChatSessionsResp, error) {
	var cursor, limit int64
	if in.GetPage() != nil {
		cursor = in.GetPage().GetCursor()
		limit = in.GetPage().GetLimit()
	}
	list, hasNext, nextCursor, err := internal.PageTransientSessions(
		l.ctx,
		l.svcCtx.BusRedis,
		in.GetMerchantId(),
		in.GetUserId(),
		in.GetAgentId(),
		int64(in.GetStatus()),
		cursor,
		limit,
	)
	if err != nil {
		return &chat.AppPageTransientChatSessionsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppPageTransientChatSessionsResp{
		Base: helper.OkWithOthers(0, hasNext, cursor > 0, nextCursor, cursor),
		Data: list,
	}, nil
}

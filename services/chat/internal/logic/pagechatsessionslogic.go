package logic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatSessionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatSessionsLogic {
	return &PageChatSessionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询会话
func (l *PageChatSessionsLogic) PageChatSessions(in *chat.PageChatSessionsReq) (*chat.PageChatSessionsResp, error) {
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatSessionModel.FindPage(l.ctx, models.ChatSessionPageFilter{
		MerchantId: in.GetMerchantId(),
		UserId:     in.GetUserId(),
		AgentId:    in.GetAgentId(),
		Status:     int64(in.GetStatus()),
		Priority:   int64(in.GetPriority()),
		Category:   strings.TrimSpace(in.GetCategory()),
		StartTime:  pageutil.TimeRangeStart(in.GetTimeRange()),
		EndTime:    pageutil.TimeRangeEnd(in.GetTimeRange()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatSessionsResp{Base: errorBase(err)}, nil
	}
	return &chat.PageChatSessionsResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoSessions(list),
	}, nil
}

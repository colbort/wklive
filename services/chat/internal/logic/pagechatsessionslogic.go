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
	merchantID := in.GetMerchantId()
	if merchantID <= 0 {
		return &chat.PageChatSessionsResp{Base: helper.ErrResp(400, "merchant_id is required")}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	filter := models.ChatSessionPageFilter{
		MerchantId: merchantID,
		UserId:     in.GetUserId(),
		AgentId:    in.GetAgentId(),
		GroupId:    in.GetGroupId(),
		Status:     int64(in.GetStatus()),
		Priority:   int64(in.GetPriority()),
		Category:   strings.TrimSpace(in.GetCategory()),
		Keyword:    strings.TrimSpace(in.GetKeyword()),
		StartTime:  pageutil.TimeRangeStart(in.GetTimeRange()),
		EndTime:    pageutil.TimeRangeEnd(in.GetTimeRange()),
	}
	transient, err := l.listTransientSessions(merchantID, filter)
	if err != nil {
		return &chat.PageChatSessionsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	dbCursor := cursor - int64(len(transient))
	if dbCursor < 0 {
		dbCursor = 0
	}
	list, _, err := l.svcCtx.ChatSessionModel.FindPage(l.ctx, filter, dbCursor, limit)
	if err != nil {
		return &chat.PageChatSessionsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data := toProtoMergedSessions(list, transient)
	return &chat.PageChatSessionsResp{
		Base: ih.OffsetBase(cursor, limit, len(data), int64(len(data))),
		Data: data,
	}, nil
}

func (l *PageChatSessionsLogic) listTransientSessions(merchantID int64, filter models.ChatSessionPageFilter) ([]*models.TChatSession, error) {
	var cursor int64
	resp := make([]*models.TChatSession, 0)
	seenUserIDs := make(map[int64]struct{})
	for {
		sessions, hasNext, nextCursor, err := ih.PageTransientSessions(
			l.ctx,
			l.svcCtx.BusRedis,
			merchantID,
			filter.UserId,
			filter.AgentId,
			filter.Status,
			cursor,
			200,
		)
		if err != nil {
			return nil, err
		}
		for _, session := range sessions {
			if !transientSessionMatchesFilter(session, filter) {
				continue
			}
			if session.UserId > 0 {
				if _, ok := seenUserIDs[session.UserId]; ok {
					continue
				}
				seenUserIDs[session.UserId] = struct{}{}
			}
			resp = append(resp, session)
		}
		if !hasNext {
			return resp, nil
		}
		cursor = nextCursor
	}
}

func transientSessionMatchesFilter(session *models.TChatSession, filter models.ChatSessionPageFilter) bool {
	if session == nil {
		return false
	}
	if filter.GroupId > 0 && session.GroupId != filter.GroupId {
		return false
	}
	if filter.Priority > 0 && int64(session.Priority) != filter.Priority {
		return false
	}
	if filter.Category != "" && session.Category != filter.Category {
		return false
	}
	if filter.StartTime > 0 && session.CreateTimes < filter.StartTime {
		return false
	}
	if filter.EndTime > 0 && session.CreateTimes > filter.EndTime {
		return false
	}
	if filter.Keyword != "" &&
		!strings.Contains(session.SessionNo, filter.Keyword) &&
		!strings.Contains(session.Title, filter.Keyword) &&
		!strings.Contains(session.LastMessage, filter.Keyword) {
		return false
	}
	return true
}

func toProtoMergedSessions(list, transient []*models.TChatSession) []*chat.ChatSession {
	resp := make([]*chat.ChatSession, 0)
	for _, session := range transient {
		resp = append(resp, ih.ToProtoSession(session, true))
	}
	for _, session := range list {
		resp = append(resp, ih.ToProtoSession(session, false))
	}
	return resp
}

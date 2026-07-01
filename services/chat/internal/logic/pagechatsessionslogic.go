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
	list, total, err := l.svcCtx.ChatSessionModel.FindPage(l.ctx, filter, dbCursor, limit)
	if err != nil {
		return &chat.PageChatSessionsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data := mergeTransientAndStoredSessions(transient, list, cursor, limit)
	return &chat.PageChatSessionsResp{
		Base: ih.OffsetBase(cursor, limit, len(data), total+int64(len(transient))),
		Data: ih.ToProtoSessions(data),
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

func mergeTransientAndStoredSessions(transient, stored []*models.TChatSession, cursor, limit int64) []*models.TChatSession {
	if limit <= 0 {
		limit = pageutil.NormalizeLimit(limit)
	}
	resp := make([]*models.TChatSession, 0, limit)
	transientTotal := int64(len(transient))
	if cursor < transientTotal {
		end := cursor + limit
		if end > transientTotal {
			end = transientTotal
		}
		resp = append(resp, transient[cursor:end]...)
	}
	if int64(len(resp)) >= limit {
		return resp
	}
	remaining := limit - int64(len(resp))
	if remaining > int64(len(stored)) {
		remaining = int64(len(stored))
	}
	if remaining > 0 {
		resp = append(resp, stored[:remaining]...)
	}
	return resp
}

package helper

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

func ToProtoQueueInfo(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*chat.ChatQueuePayload, error) {
	if session == nil {
		return nil, nil
	}
	position, waitingCount, err := svcCtx.ChatSessionModel.CountWaitingPosition(ctx, session)
	if err != nil {
		return nil, err
	}
	transientPosition, transientWaitingCount, err := countTransientWaitingPosition(ctx, svcCtx, session)
	if err != nil {
		return nil, err
	}
	position += transientPosition
	waitingCount += transientWaitingCount
	if session.Id == 0 && session.AgentId == 0 && queuedSessionStatus(session.Status) {
		position++
	}
	return &chat.ChatQueuePayload{
		SessionNo:            session.SessionNo,
		UserId:               session.UserId,
		QueueAction:          chat.ChatQueueAction_CHAT_QUEUE_ACTION_UPDATE,
		QueuePosition:        position,
		WaitingCount:         waitingCount,
		EstimatedWaitSeconds: EstimateWaitSeconds(position),
		SessionStatus:        chat.ChatSessionStatus(session.Status),
		ActionTime:           utils.NowMillis(),
	}, nil
}

func EstimateWaitSeconds(position int64) int64 {
	if position <= 1 {
		return 0
	}
	return (position - 1) * 60
}

func countTransientWaitingPosition(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (int64, int64, error) {
	if svcCtx == nil || svcCtx.BusRedis == nil || session == nil || session.MerchantId <= 0 {
		return 0, 0, nil
	}
	var cursor int64
	var position int64
	var waitingCount int64
	for {
		list, hasNext, nextCursor, err := PageTransientSessions(
			ctx,
			svcCtx.BusRedis,
			session.MerchantId,
			0,
			0,
			0,
			cursor,
			200,
		)
		if err != nil {
			return 0, 0, err
		}
		for _, item := range list {
			if item.AgentId != 0 ||
				item.GroupId != session.GroupId ||
				!queuedSessionStatus(item.Status) {
				continue
			}
			waitingCount++
			if item.SessionNo == session.SessionNo {
				continue
			}
			if transientBeforeSession(item, session) {
				position++
			}
		}
		if !hasNext {
			return position, waitingCount, nil
		}
		cursor = nextCursor
	}
}

func transientBeforeSession(item *models.TChatSession, session *models.TChatSession) bool {
	if item == nil || session == nil {
		return false
	}
	itemPriority := item.Priority
	if itemPriority != session.Priority {
		return itemPriority > session.Priority
	}
	itemCreatedAt := item.CreateTimes
	if itemCreatedAt != session.CreateTimes {
		return itemCreatedAt < session.CreateTimes
	}
	return item.SessionNo <= session.SessionNo
}

func queuedSessionStatus(status int64) bool {
	return status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) ||
		status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
}

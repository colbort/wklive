package internal

import (
	"context"
	"strconv"

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
	return &chat.ChatQueuePayload{
		SessionNo:            session.SessionNo,
		UserId:               strconv.FormatInt(session.UserId, 10),
		QueueAction:          chat.ChatQueueAction_CHAT_QUEUE_ACTION_UPDATE,
		QueuePosition:        int32(position),
		WaitingCount:         int32(waitingCount),
		EstimatedWaitSeconds: int32(EstimateWaitSeconds(position)),
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

package internal

import (
	"context"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

func ToProtoQueueInfo(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*chat.ChatQueueInfo, error) {
	if session == nil {
		return nil, nil
	}
	position, waitingCount, err := svcCtx.ChatSessionModel.CountWaitingPosition(ctx, session)
	if err != nil {
		return nil, err
	}
	message := "正在排队，客服会尽快接入。"
	if position > 0 {
		if position == 1 {
			message = "您是当前队列第 1 位，客服即将接入。"
		} else {
			message = fmt.Sprintf("正在排队，您前面还有 %d 人。", position-1)
		}
	}
	if session.AgentId > 0 ||
		session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) ||
		session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) {
		message = "客服已接入。"
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		message = "本次会话已结束。"
	}
	return &chat.ChatQueueInfo{
		MerchantId:          session.MerchantId,
		SessionNo:           session.SessionNo,
		UserId:              session.UserId,
		GroupId:             session.GroupId,
		Position:            int32(position),
		WaitingCount:        int32(waitingCount),
		EstimateWaitSeconds: EstimateWaitSeconds(position),
		Message:             message,
		UpdateTimes:         utils.NowMillis(),
	}, nil
}

func EstimateWaitSeconds(position int64) int64 {
	if position <= 1 {
		return 0
	}
	return (position - 1) * 60
}

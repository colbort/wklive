package internal

import (
	"context"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

func GetSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantID int64, sessionNo string) (*models.TChatSession, *common.RespBase, error) {
	data, err := svcCtx.ChatSessionModel.FindOneBySessionNo(ctx, sessionNo)
	if err == models.ErrNotFound {
		return nil, helper.ErrResp(404, "chat session not found"), nil
	}
	if err != nil {
		return nil, nil, err
	}
	if data.MerchantId != merchantID {
		return nil, helper.ErrResp(404, "chat session not found"), nil
	}
	return data, nil, nil
}

func NormalizeAssignType(value chat.ChatAssignType) chat.ChatAssignType {
	if value == chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN {
		return chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL
	}
	return value
}

func ChangeAgentSessionCount(ctx context.Context, svcCtx *svc.ServiceContext, agentID int64, delta int64) error {
	if agentID <= 0 || delta == 0 {
		return nil
	}
	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, agentID)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	agent.CurrentSessionCount += delta
	if agent.CurrentSessionCount < 0 {
		agent.CurrentSessionCount = 0
	}
	agent.UpdateTimes = utils.NowMillis()
	return svcCtx.ChatAgentModel.Update(ctx, agent)
}

type AssignSessionOptions struct {
	SessionNo  string
	MerchantId int64
	ToAgentId  int64
	AssignType chat.ChatAssignType
	Reason     string
}

func AssignSession(ctx context.Context, svcCtx *svc.ServiceContext, in AssignSessionOptions) (*models.TChatSession, *common.RespBase, error) {
	session, base, err := GetSession(ctx, svcCtx, in.MerchantId, in.SessionNo)
	if base != nil || err != nil {
		return nil, base, err
	}
	if in.ToAgentId <= 0 {
		return nil, helper.ErrResp(400, "to_agent_id is required"), nil
	}
	assignType := NormalizeAssignType(in.AssignType)
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, helper.ErrResp(400, "chat session is closed"), nil
	}
	if base := ValidateAssignableSession(session, assignType); base != nil {
		return nil, base, nil
	}
	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, in.ToAgentId)
	if err == models.ErrNotFound || agent.MerchantId != in.MerchantId {
		return nil, helper.ErrResp(404, "chat agent not found"), nil
	}
	if err != nil {
		return nil, nil, err
	}

	fromAgentID := session.AgentId
	if fromAgentID != agent.Id {
		if err := ChangeAgentSessionCount(ctx, svcCtx, fromAgentID, -1); err != nil {
			return nil, nil, err
		}
		if err := ChangeAgentSessionCount(ctx, svcCtx, agent.Id, 1); err != nil {
			return nil, nil, err
		}
	}

	now := utils.NowMillis()
	session.AgentId = agent.Id
	session.GroupId = agent.GroupId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, nil, err
	}

	_, err = svcCtx.ChatAssignmentModel.Insert(ctx, &models.TChatAssignment{
		SessionNo:   session.SessionNo,
		MerchantId:  session.MerchantId,
		FromAgentId: fromAgentID,
		ToAgentId:   agent.Id,
		AssignType:  int64(assignType),
		Reason:      strings.TrimSpace(in.Reason),
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, nil, err
	}

	return session, nil, nil
}

func ValidateAssignableSession(session *models.TChatSession, assignType chat.ChatAssignType) *common.RespBase {
	switch assignType {
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_AUTO:
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return helper.ErrResp(400, "chat session cannot be assigned")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL:
		if session.AgentId != 0 {
			return helper.ErrResp(400, "chat session is already accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return helper.ErrResp(400, "chat session cannot be accepted")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_TRANSFER:
		if session.AgentId == 0 {
			return helper.ErrResp(400, "chat session is not accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return helper.ErrResp(400, "chat session cannot be transferred")
		}
	}
	return nil
}

func ReleaseSessionToPool(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) (*models.TChatSession, *common.RespBase, error) {
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, helper.ErrResp(400, "chat session is closed"), nil
	}

	fromAgentID := session.AgentId
	if fromAgentID > 0 {
		if err := ChangeAgentSessionCount(ctx, svcCtx, fromAgentID, -1); err != nil {
			return nil, nil, err
		}
	}

	now := utils.NowMillis()
	session.AgentId = 0
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return nil, nil, err
	}

	if fromAgentID > 0 {
		_, err := svcCtx.ChatAssignmentModel.Insert(ctx, &models.TChatAssignment{
			SessionNo:   session.SessionNo,
			MerchantId:  session.MerchantId,
			FromAgentId: fromAgentID,
			ToAgentId:   0,
			AssignType:  int64(chat.ChatAssignType_CHAT_ASSIGN_TYPE_TRANSFER),
			Reason:      strings.TrimSpace(reason),
			CreateTimes: now,
			UpdateTimes: now,
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return session, nil, nil
}

func RouteSessionToAvailableAgent(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) (*models.TChatSession, *common.RespBase, error) {
	agents, err := svcCtx.ChatAgentModel.FindAvailable(ctx, session.MerchantId, session.GroupId, 2)
	if err != nil {
		return nil, nil, err
	}
	if len(agents) == 1 {
		if session.AgentId == agents[0].Id {
			return session, nil, nil
		}
		return AssignSession(ctx, svcCtx, AssignSessionOptions{
			SessionNo:  session.SessionNo,
			MerchantId: session.MerchantId,
			ToAgentId:  agents[0].Id,
			AssignType: chat.ChatAssignType_CHAT_ASSIGN_TYPE_AUTO,
			Reason:     reason,
		})
	}
	if session.AgentId > 0 {
		return ReleaseSessionToPool(ctx, svcCtx, session, reason)
	}
	return session, nil, nil
}

func PrepareSessionForUserMessage(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*models.TChatSession, *common.RespBase, error) {
	if session.AgentId == 0 || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return RouteSessionToAvailableAgent(ctx, svcCtx, session, "auto assign")
	}

	agent, err := svcCtx.ChatAgentModel.FindOne(ctx, session.AgentId)
	if err == nil && agent.MerchantId == session.MerchantId && agent.Status == int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE) {
		return session, nil, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, nil, err
	}

	return RouteSessionToAvailableAgent(ctx, svcCtx, session, "current agent unavailable")
}

func AutoAssignSession(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) error {
	if session.AgentId != 0 || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil
	}
	_, _, err := RouteSessionToAvailableAgent(ctx, svcCtx, session, "auto assign")
	return err
}

func MarkRead(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, readerType chat.ChatSenderType, readerID int64) error {
	now := utils.NowMillis()
	lastNo := session.LastMessageNo
	cursor, err := svcCtx.ChatReadCursorModel.FindOneByMerchantIdSessionNoReaderTypeReaderIdDeviceId(ctx, session.MerchantId, session.SessionNo, int64(readerType), readerID, DefaultDeviceID)
	switch err {
	case models.ErrNotFound:
		_, err = svcCtx.ChatReadCursorModel.Insert(ctx, &models.TChatReadCursor{
			MerchantId:        session.MerchantId,
			SessionNo:         session.SessionNo,
			ReaderType:        int64(readerType),
			ReaderId:          readerID,
			DeviceId:          DefaultDeviceID,
			LastReadMessageNo: lastNo,
			LastReadTime:      now,
			CreateTimes:       now,
			UpdateTimes:       now,
		})
	case nil:
		cursor.LastReadMessageNo = lastNo
		cursor.LastReadTime = now
		cursor.UpdateTimes = now
		err = svcCtx.ChatReadCursorModel.Update(ctx, cursor)
	}
	if err != nil {
		return err
	}

	switch readerType {
	case chat.ChatSenderType_CHAT_SENDER_TYPE_USER:
		session.UserUnreadCount = 0
	case chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT:
		session.AgentUnreadCount = 0
	}
	session.UpdateTimes = now
	return svcCtx.ChatSessionModel.Update(ctx, session)
}

func CloseSession(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, reason string) error {
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil
	}
	now := utils.NowMillis()
	oldAgentID := session.AgentId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
	session.CloseTime = now
	session.CloseReason = strings.TrimSpace(reason)
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return err
	}
	return ChangeAgentSessionCount(ctx, svcCtx, oldAgentID, -1)
}

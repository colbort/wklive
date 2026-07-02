package helper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"
)

const (
	InternetErrorCloseDelay = 3 * time.Minute

	transientInternetErrorKey = "chat:session:internet_error_timeout"
)

func GetSession(ctx context.Context, svcCtx *svc.ServiceContext, merchantID int64, sessionNo string, isGuest bool) (*models.TChatSession, error) {
	if merchantID <= 0 || strings.TrimSpace(sessionNo) == "" {
		return nil, errors.New("params err, merchant id is invalid or session no is empty")
	}
	if isGuest {
		session, err := GetTransientSession(ctx, svcCtx.BusRedis, merchantID, sessionNo)
		if err != nil {
			return nil, errors.New("chat session not found")
		}
		if session.MerchantId != merchantID {
			return nil, errors.New("chat session not found")
		}
		return session, nil
	} else {
		data, err := svcCtx.ChatSessionModel.FindOneBySessionNo(ctx, sessionNo)
		if err == models.ErrNotFound {
			return nil, errors.New("chat session not found")
		}
		if err != nil {
			return nil, err
		}
		if data.MerchantId != merchantID {
			return nil, errors.New("chat session not found")
		}
		return data, nil
	}
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
	Agent      *models.TChatAgent
	AssignType chat.ChatAssignType
	Reason     string
	IsGuest    bool
}

func AcceptChatSession(ctx context.Context, svcCtx *svc.ServiceContext, in AssignSessionOptions) (*models.TChatSession, error) {
	session, err := GetSession(ctx, svcCtx, in.Agent.MerchantId, in.SessionNo, in.IsGuest)
	if err != nil {
		return nil, err
	}
	if in.Agent == nil || in.Agent.Id <= 0 {
		return nil, errors.New("invalid agent")
	}
	assignType := NormalizeAssignType(in.AssignType)
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil, errors.New("chat session is closed")
	}
	if err := ValidateAssignableSession(session, assignType); err != nil {
		return nil, err
	}

	fromAgentID := session.AgentId
	if fromAgentID != in.Agent.Id {
		if err := ChangeAgentSessionCount(ctx, svcCtx, fromAgentID, -1); err != nil {
			return nil, err
		}
		if err := ChangeAgentSessionCount(ctx, svcCtx, in.Agent.Id, 1); err != nil {
			return nil, err
		}
	}

	now := utils.NowMillis()
	session.AgentId = in.Agent.Id
	session.AgentUserId = in.Agent.UserId
	session.GroupId = in.Agent.GroupId
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING)
	session.UpdateTimes = now
	if in.IsGuest {
		session, err = UpsertTransientSession(ctx, svcCtx.BusRedis, session)
		if err != nil {
			return nil, err
		}
	} else {
		if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
			return nil, err
		}
	}

	_, err = svcCtx.ChatAssignmentModel.Insert(ctx, &models.TChatAssignment{
		SessionNo:   session.SessionNo,
		MerchantId:  session.MerchantId,
		FromAgentId: fromAgentID,
		ToAgentId:   in.Agent.Id,
		AssignType:  int64(assignType),
		Reason:      strings.TrimSpace(in.Reason),
		CreateTimes: now,
		UpdateTimes: now,
	})
	if err != nil {
		return nil, err
	}

	return session, nil
}

func ValidateAssignableSession(session *models.TChatSession, assignType chat.ChatAssignType) error {
	switch assignType {
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_AUTO:
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return errors.New("chat session cannot be assigned")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_MANUAL:
		if session.AgentId != 0 {
			return errors.New("chat session is already accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return errors.New("chat session cannot be accepted")
		}
	case chat.ChatAssignType_CHAT_ASSIGN_TYPE_TRANSFER:
		if session.AgentId == 0 {
			return errors.New("chat session is not accepted")
		}
		if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_USER) &&
			session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_PENDING_AGENT) {
			return errors.New("chat session cannot be transferred")
		}
	}
	return nil
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
	session.DisconnectTime = 0
	session.BeforeDisconnectStatus = 0
	session.UpdateTimes = now
	if err := svcCtx.ChatSessionModel.Update(ctx, session); err != nil {
		return err
	}
	return ChangeAgentSessionCount(ctx, svcCtx, oldAgentID, -1)
}

func MarkSessionInternetError(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) error {
	if session == nil || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return nil
	}
	now := utils.NowMillis()
	markInternetErrorSessionFields(session, now)
	return svcCtx.ChatSessionModel.Update(ctx, session)
}

func IsInternetErrorExpired(session *models.TChatSession, now int64) bool {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		return false
	}
	return session.DisconnectTime > 0 && now-session.DisconnectTime >= int64(InternetErrorCloseDelay/time.Millisecond)
}

func MarkTransientSessionInternetError(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*models.TChatSession, error) {
	if session == nil || session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return session, nil
	}
	now := utils.NowMillis()
	markInternetErrorSessionFields(session, now)
	next, err := UpsertTransientSession(ctx, svcCtx.BusRedis, session)
	if err != nil {
		return nil, err
	}
	if err := addTransientInternetErrorTimeout(ctx, svcCtx, next); err != nil {
		return nil, err
	}
	return next, nil
}

func RestoreSessionInternetError(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (bool, error) {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		return false, nil
	}
	now := utils.NowMillis()
	if IsInternetErrorExpired(session, now) {
		return false, CloseSession(ctx, svcCtx, session, "用户网络异常超时关闭")
	}
	restoreInternetErrorSessionFields(session, now)
	return true, svcCtx.ChatSessionModel.Update(ctx, session)
}

func RestoreTransientSessionInternetError(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) (*models.TChatSession, bool, error) {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		return session, false, nil
	}
	now := utils.NowMillis()
	if IsInternetErrorExpired(session, now) {
		session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
		session.CloseTime = now
		session.CloseReason = "用户网络异常超时关闭"
		session.DisconnectTime = 0
		session.BeforeDisconnectStatus = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN)
		session.UpdateTimes = now
		return upsertTransientSessionAndRemoveInternetErrorTimeout(ctx, svcCtx, session, false)
	}
	restoreInternetErrorSessionFields(session, now)
	return upsertTransientSessionAndRemoveInternetErrorTimeout(ctx, svcCtx, session, true)
}

func markInternetErrorSessionFields(session *models.TChatSession, now int64) {
	if session == nil {
		return
	}
	if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		session.BeforeDisconnectStatus = session.Status
	}
	session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR)
	session.DisconnectTime = now
	session.UpdateTimes = now
}

func restoreInternetErrorSessionFields(session *models.TChatSession, now int64) {
	if session == nil {
		return
	}
	status := session.BeforeDisconnectStatus
	if status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN) ||
		status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) ||
		status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
		status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
		if session.AgentId > 0 {
			status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_SERVING)
		}
	}
	session.Status = status
	session.DisconnectTime = 0
	session.BeforeDisconnectStatus = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN)
	session.UpdateTimes = now
}

func upsertTransientSessionAndRemoveInternetErrorTimeout(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession, restored bool) (*models.TChatSession, bool, error) {
	next, err := UpsertTransientSession(ctx, svcCtx.BusRedis, session)
	_ = RemoveTransientInternetErrorTimeout(ctx, svcCtx, session.MerchantId, session.SessionNo)
	return next, restored, err
}

func IsInternetErrorCloseReason(reasonType chat.ChatSessionCloseReason, reason string) bool {
	return reasonType == chat.ChatSessionCloseReason_CHAT_SESSION_CLOSE_REASON_INTERNET_ERROR
}

func addTransientInternetErrorTimeout(ctx context.Context, svcCtx *svc.ServiceContext, session *models.TChatSession) error {
	if svcCtx.BusRedis == nil || session == nil {
		return nil
	}
	score := session.DisconnectTime + int64(InternetErrorCloseDelay/time.Millisecond)
	member := transientInternetErrorMember(session.MerchantId, session.SessionNo)
	_, err := svcCtx.BusRedis.ZaddCtx(ctx, transientInternetErrorKey, score, member)
	return err
}

func RemoveTransientInternetErrorTimeout(ctx context.Context, svcCtx *svc.ServiceContext, merchantID int64, sessionNo string) error {
	if svcCtx.BusRedis == nil {
		return nil
	}
	_, err := svcCtx.BusRedis.ZremCtx(ctx, transientInternetErrorKey, transientInternetErrorMember(merchantID, sessionNo))
	return err
}

func transientInternetErrorMember(merchantID int64, sessionNo string) string {
	return fmt.Sprintf("%d:%s", merchantID, strings.TrimSpace(sessionNo))
}

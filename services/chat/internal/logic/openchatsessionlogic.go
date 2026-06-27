package logic

import (
	"context"
	"database/sql"
	"strings"
	"wklive/common/helper"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type OpenChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OpenChatSessionLogic {
	return &OpenChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建或获取当前会话;
// 用户进入客服页面 建立 WS ；创建/复用会话
// 游客创建/复用（本次会话未结束）临时会话，会话不保存；登录用户创建（首次打开）/复用（之后进来）会话，会话保存数据库
func (l *OpenChatSessionLogic) OpenChatSession(in *chat.OpenChatSessionReq) (*chat.AppChatSessionResp, error) {
	var session *models.TChatSession
	var respSession *chat.ChatSession
	if in.IsGuest {
		sessionNo := strings.TrimSpace(in.GetSessionNo())
		if sessionNo == "" {
			var err error
			sessionNo, err = l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				logx.Errorf("generate guest session no error: %v", err)
				return &chat.AppChatSessionResp{Base: helper.ErrResp(400, "generate session no error")}, nil
			}
		}

		respSession, _ = internal.GetTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), sessionNo)
		now := utils.NowMillis()
		if respSession == nil {
			respSession = &chat.ChatSession{
				SessionNo:       sessionNo,
				MerchantId:      in.GetMerchantId(),
				UserId:          in.GetUserId(),
				Source:          in.GetSource(),
				Status:          chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING,
				Priority:        chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL,
				LastMessageTime: now,
				CreateTimes:     now,
				UpdateTimes:     now,
				IsGuest:         true,
			}
		} else if respSession.GetStatus() == chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED {
			respSession.Status = chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING
			respSession.AgentId = 0
			respSession.CloseTime = 0
			respSession.CloseReason = ""
			respSession.UpdateTimes = now
		}
		respSession.MerchantId = in.GetMerchantId()
		respSession.UserId = in.GetUserId()
		respSession.IsGuest = true

		var err error
		respSession, err = internal.UpsertTransientSession(l.ctx, l.svcCtx.BusRedis, respSession, 0)
		if err != nil {
			return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		session = transientSessionToModel(respSession)
	} else {
		// 登录用户
		session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
		if err != nil && err != models.ErrNotFound {
			return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		if err == nil {
			now := utils.NowMillis()
			if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
				session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
				session.AgentId = 0
				session.CloseTime = 0
				session.CloseReason = ""
				session.UpdateTimes = now
				if err := l.svcCtx.ChatSessionModel.Update(l.ctx, session); err != nil {
					return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
				}
			}
		} else {
			sessionNo, err := l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				logx.Errorf("generate session no error: %v", err)
				return &chat.AppChatSessionResp{Base: helper.ErrResp(400, "generate message no error")}, nil
			}
			now := utils.NowMillis()
			session = &models.TChatSession{
				SessionNo:       sessionNo,
				MerchantId:      in.MerchantId,
				UserId:          in.UserId,
				Source:          int64(in.Source),
				Status:          int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING),
				Priority:        int64(chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL),
				Title:           "",
				Category:        "",
				LastMessageTime: now,
				ExtJson:         sql.NullString{String: "", Valid: true},
				CreateTimes:     now,
				UpdateTimes:     now,
			}
			result, err := l.svcCtx.ChatSessionModel.Insert(l.ctx, session)
			if err != nil {
				return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
			if id, err := result.LastInsertId(); err == nil {
				session.Id = id
			}
		}

	}
	if respSession == nil {
		respSession = internal.ToProtoSession(session, in.IsGuest)
	}

	// 向坐席 chat-admin-api 推送 用户上线通知
	l.publishUserJoinEvent(session, in.IsGuest)
	// 向用户同送排队信息
	l.publishQueueJoinEvent(session, in.IsGuest)
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: respSession}, nil
}

func transientSessionToModel(session *chat.ChatSession) *models.TChatSession {
	if session == nil {
		return nil
	}
	return &models.TChatSession{
		Id:               session.GetId(),
		SessionNo:        session.GetSessionNo(),
		MerchantId:       session.GetMerchantId(),
		UserId:           session.GetUserId(),
		Source:           int64(session.GetSource()),
		Status:           int64(session.GetStatus()),
		Priority:         int64(session.GetPriority()),
		AgentId:          session.GetAgentId(),
		GroupId:          session.GetGroupId(),
		Title:            session.GetTitle(),
		Category:         session.GetCategory(),
		LastMessageNo:    session.GetLastMessageNo(),
		LastMessage:      session.GetLastMessage(),
		LastSenderType:   int64(session.GetLastSenderType()),
		LastMessageTime:  session.GetLastMessageTime(),
		UserUnreadCount:  int64(session.GetUserUnreadCount()),
		AgentUnreadCount: int64(session.GetAgentUnreadCount()),
		CloseTime:        session.GetCloseTime(),
		CloseReason:      session.GetCloseReason(),
		ExtJson:          internal.StructToNullString(session.GetExtJson()),
		CreateTimes:      session.GetCreateTimes(),
		UpdateTimes:      session.GetUpdateTimes(),
	}
}

func (l *OpenChatSessionLogic) publishUserJoinEvent(session *models.TChatSession, isGuest bool) {
	if session == nil {
		return
	}
	internal.PublishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN, isGuest, session, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, "", "用户进入会话", chat.ChatAppMessageChannel)
}

func (l *OpenChatSessionLogic) publishQueueJoinEvent(session *models.TChatSession, isGuest bool) {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) {
		return
	}
	internal.PublishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_JOIN, isGuest, session, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, "", "正在排队，客服会尽快接入。", chat.ChatAppMessageChannel)
}

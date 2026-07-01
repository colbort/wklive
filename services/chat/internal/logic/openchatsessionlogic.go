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
// 用户进入客服页面建立 WS，创建/复用会话。
// 游客会话保存在 Redis：无会话时创建临时会话，未过期会话复用；已关闭但未过期的会话会重新打开。
// 登录用户会话保存在数据库：无会话时创建会话，已有会话按用户最新会话复用。
func (l *OpenChatSessionLogic) OpenChatSession(in *chat.OpenChatSessionReq) (*chat.OpenChatSessionResp, error) {
	var ms *models.TChatSession
	if in.IsGuest {
		sessionNo := strings.TrimSpace(in.GetSessionNo())
		var err error
		if sessionNo == "" {
			sessionNo, err = l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				logx.Errorf("generate guest session no error: %v", err)
				return &chat.OpenChatSessionResp{Base: helper.ErrResp(400, "generate session no error")}, nil
			}
		}

		ms, _ = internal.GetTransientSession(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), sessionNo)
		now := utils.NowMillis()
		if ms == nil {
			ms = &models.TChatSession{
				SessionNo:       sessionNo,
				MerchantId:      in.GetMerchantId(),
				UserId:          in.GetUserId(),
				Source:          int64(in.GetSource()),
				Status:          int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING),
				Priority:        int64(chat.ChatSessionPriority_CHAT_SESSION_PRIORITY_NORMAL),
				LastMessageTime: now,
				ExtJson:         sql.NullString{String: in.ExtJson, Valid: true},
				CreateTimes:     now,
				UpdateTimes:     now,
			}
		} else {
			if ms.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
				// 会话断线重连，恢复会话状态
				var restored bool
				var err error
				ms, restored, err = internal.RestoreTransientSessionInternetError(l.ctx, l.svcCtx, ms)
				if err != nil {
					return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
				}
				if restored {
					now = utils.NowMillis()
				}
			} else if ms.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
				// 游客会话已关闭，但是会话未过期，重新打开会话
				ms.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
				ms.CloseTime = 0
				ms.CloseReason = ""
				ms.DisconnectTime = 0
				ms.BeforeDisconnectStatus = 0
				ms.UpdateTimes = now
			}
		}
		ms.MerchantId = in.GetMerchantId()
		ms.UserId = in.GetUserId()

		ms, err = internal.UpsertTransientSession(l.ctx, l.svcCtx.BusRedis, ms)
		if err != nil {
			return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	} else {
		// 登录用户
		session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
		if err != nil && err != models.ErrNotFound {
			return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		if session == nil {
			sessionNo, err := l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				logx.Errorf("generate session no error: %v", err)
				return &chat.OpenChatSessionResp{Base: helper.ErrResp(400, "generate message no error")}, nil
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
				ExtJson:         sql.NullString{String: in.ExtJson, Valid: true},
				CreateTimes:     now,
				UpdateTimes:     now,
			}
			result, err := l.svcCtx.ChatSessionModel.Insert(l.ctx, session)
			if err != nil {
				return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
			if id, err := result.LastInsertId(); err == nil {
				session.Id = id
			}
		} else {
			restoredInternetError := false
			if err == nil && session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_INTERNET_ERROR) {
				var restored bool
				var err error
				restored, err = internal.RestoreSessionInternetError(l.ctx, l.svcCtx, session)
				if err != nil {
					return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
				}
				restoredInternetError = restored
			}
			now := utils.NowMillis()
			if !restoredInternetError {
				session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
				session.AgentId = 0
			}
			session.CloseTime = 0
			session.CloseReason = ""
			session.DisconnectTime = 0
			session.BeforeDisconnectStatus = 0
			session.UpdateTimes = now
			if err := l.svcCtx.ChatSessionModel.Update(l.ctx, session); err != nil {
				return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		ms = session
	}
	// 向坐席 chat-admin-api 推送 用户上线通知
	l.publishUserJoinEvent(ms, in.IsGuest)
	queue, err := internal.ToProtoQueueInfo(l.ctx, l.svcCtx, ms)
	if err != nil {
		return &chat.OpenChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if queue == nil {
		queue = &chat.ChatQueuePayload{
			SessionNo:  ms.SessionNo,
			UserId:     ms.UserId,
			ActionTime: utils.NowMillis(),
		}
	}
	queue.QueueAction = chat.ChatQueueAction_CHAT_QUEUE_ACTION_JOIN
	return &chat.OpenChatSessionResp{Base: helper.OkResp(), Data: queue}, nil
}

func (l *OpenChatSessionLogic) publishUserJoinEvent(session *models.TChatSession, isGuest bool) {
	if session == nil {
		l.Logger.Info("push event to admin err: session is nil")
		return
	}
	_ = internal.PublishMessageEvent(l.ctx, l.svcCtx, internal.PublishMessageEventReq{
		EventType:    chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN,
		Channel:      chat.ChatAdminEventChannel,
		IsGuest:      isGuest,
		Session:      session,
		AssignType:   chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN,
		EventMessage: "用户进入会话",
	})
}

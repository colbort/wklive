package logic

import (
	"context"
	"database/sql"
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
	if in.IsGuest {
		// 游客
	} else {
		// 登录用户
		session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
		if err != models.ErrNotFound {
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

	// 向坐席 chat-admin-api 推送 用户上线通知
	l.publishUserJoinEvent(session, in.IsGuest)
	// 向用户同送排队信息
	l.publishQueueJoinEvent(session, in.IsGuest)
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: internal.ToProtoSession(session, in.IsGuest)}, nil
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

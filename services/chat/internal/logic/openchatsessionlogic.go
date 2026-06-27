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
	"google.golang.org/protobuf/types/known/structpb"
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

// 创建或获取当前会话
func (l *OpenChatSessionLogic) OpenChatSession(in *chat.OpenChatSessionReq) (*chat.AppChatSessionResp, error) {
	session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
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
		l.publishUserJoinEvent(session)
		l.publishQueueJoinEvent(session)
		return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: internal.ToProtoSession(session)}, nil
	}
	if err != models.ErrNotFound {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

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
	l.publishUserJoinEvent(session)
	l.publishQueueJoinEvent(session)
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: internal.ToProtoSession(session)}, nil
}

func (l *OpenChatSessionLogic) publishUserJoinEvent(session *models.TChatSession) {
	if session == nil {
		return
	}
	internal.PublishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_USER_JOIN, session, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, "", "用户进入会话", chat.ChatAppMessageChannel)
}

func (l *OpenChatSessionLogic) publishQueueJoinEvent(session *models.TChatSession) {
	if session == nil || session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING) {
		return
	}
	internal.PublishSessionEvent(l.ctx, l.svcCtx, chat.ChatEventType_CHAT_EVENT_TYPE_QUEUE_JOIN, session, chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN, "", "正在排队，客服会尽快接入。", chat.ChatAppMessageChannel)
}

func userSnapshotExt(avatarUrl string) *structpb.Struct {
	avatarUrl = strings.TrimSpace(avatarUrl)
	if avatarUrl == "" {
		return nil
	}
	ext, err := structpb.NewStruct(map[string]interface{}{
		"userAvatarUrl": avatarUrl,
	})
	if err != nil {
		return nil
	}
	return ext
}

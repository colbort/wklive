package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseMyChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseMyChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseMyChatSessionLogic {
	return &CloseMyChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭我的会话
func (l *CloseMyChatSessionLogic) CloseMyChatSession(in *chat.CloseMyChatSessionReq) (*chat.AppChatSessionResp, error) {
	session, base, err := ih.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.GetSessionNo(), in.GetIsGuest())
	if err != nil {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if base != nil {
		return &chat.AppChatSessionResp{Base: base}, nil
	}
	if session.UserId != in.GetUserId() {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, true)}, nil
	}
	if ih.IsInternetErrorCloseReason(in.GetCloseReasonType(), in.GetCloseReason()) {
		if in.IsGuest {
			session, err = ih.MarkTransientSessionInternetError(l.ctx, l.svcCtx, session)
			if err != nil {
				return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		_ = ih.PublishMessageEvent(l.ctx, l.svcCtx, ih.PublishMessageEventReq{
			EventType:    chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
			Channel:      chat.ChatAdminEventChannel,
			IsGuest:      true,
			Session:      session,
			AssignType:   chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN,
			Reason:       in.GetCloseReason(),
			EventMessage: "用户网络异常断开",
		})
	}
	if in.GetIsGuest() {
		now := utils.NowMillis()
		session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED)
		session.CloseTime = now
		session.CloseReason = in.GetCloseReason()
		session.DisconnectTime = 0
		session.BeforeDisconnectStatus = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN)
		session.UpdateTimes = now
		_ = ih.RemoveTransientInternetErrorTimeout(l.ctx, l.svcCtx, session.MerchantId, session.SessionNo)
		session, err = ih.UpsertTransientSession(l.ctx, l.svcCtx.BusRedis, session)
		if err != nil {
			return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	} else {
		if err := ih.CloseSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
			return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	}
	_ = ih.PublishMessageEvent(l.ctx, l.svcCtx, ih.PublishMessageEventReq{
		EventType:    chat.ChatEventType_CHAT_EVENT_TYPE_USER_LEAVE,
		Channel:      chat.ChatAdminEventChannel,
		Session:      session,
		AssignType:   chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN,
		Reason:       in.GetCloseReason(),
		EventMessage: "本次会话已结束",
	})
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, in.IsGuest)}, nil
}

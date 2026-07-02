package logic

import (
	"context"
	"fmt"
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
	session, err := ih.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.GetSessionNo(), in.GetIsGuest())
	if err != nil {
		return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
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
		} else {
			if err := ih.MarkSessionInternetError(l.ctx, l.svcCtx, session); err != nil {
				return &chat.AppChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		_ = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAdminEventChannel, ih.PublishEventError, &chat.ChatWsResponse_Error{
			Error: &chat.ChatErrorPayload{
				MessageNo:    "",
				ErrorCode:    0,
				ErrorMessage: "internet error",
				Detail:       fmt.Sprintf("user %d internet error: %s", session.UserId, in.CloseReason),
				Retryable:    false,
			},
		})
		return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, in.IsGuest)}, nil
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
	sessionPayload := chat.ChatWsResponse_Session{Session: ih.ToProtoSession(session, in.IsGuest)}
	_ = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAppEventChannel, ih.PublishEventSessionClose, &sessionPayload)
	_ = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAdminEventChannel, ih.PublishEventUserLeave, &chat.ChatWsResponse_UserState{
		UserState: &chat.ChatUserStatePayload{
			SessionNo: in.SessionNo,
			UserId:    session.UserId,
			UserName:  "",
			Avatar:    "",
			Online:    true,
			Source:    chat.ChatSessionSource_CHAT_SESSION_SOURCE_APP,
		},
	})
	return &chat.AppChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, in.IsGuest)}, nil
}

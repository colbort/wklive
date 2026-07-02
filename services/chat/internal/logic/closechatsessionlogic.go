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

type CloseChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCloseChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseChatSessionLogic {
	return &CloseChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关闭会话
func (l *CloseChatSessionLogic) CloseChatSession(in *chat.CloseChatSessionReq) (*chat.AdminChatSessionResp, error) {
	session, err := ih.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.GetSessionNo(), in.GetIsGuest())
	if err != nil {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if session.UserId != in.GetUserId() {
		return &chat.AdminChatSessionResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}
	if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
		return &chat.AdminChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, true)}, nil
	}
	if ih.IsInternetErrorCloseReason(in.GetCloseReasonType(), in.GetCloseReason()) {
		// 网络错误
		if in.IsGuest {
			session, err = ih.MarkTransientSessionInternetError(l.ctx, l.svcCtx, session)
			if err != nil {
				return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		} else {
			if err := ih.MarkSessionInternetError(l.ctx, l.svcCtx, session); err != nil {
				return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		_ = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAppEventChannel, ih.PublishEventError, &chat.ChatWsResponse_Error{
			Error: &chat.ChatErrorPayload{
				MessageNo:    "",
				ErrorCode:    0,
				ErrorMessage: "internet error",
				Detail:       fmt.Sprintf("agent %d internet error: %s", session.AgentId, in.CloseReason),
				Retryable:    false,
			},
		})
		return &chat.AdminChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, in.IsGuest)}, nil
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
			return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	} else {
		if err := ih.CloseSession(l.ctx, l.svcCtx, session, in.GetCloseReason()); err != nil {
			return &chat.AdminChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	}
	_ = ih.PublishMessageEvent(l.ctx, l.svcCtx.BusRedis, chat.ChatAppEventChannel, ih.PublishEventSessionClose, &chat.ChatWsResponse_Session{Session: ih.ToProtoSession(session, false)})
	return &chat.AdminChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(session, false)}, nil
}

package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateChatSessionNoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateChatSessionNoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateChatSessionNoLogic {
	return &GenerateChatSessionNoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 生成会话编号
func (l *GenerateChatSessionNoLogic) GenerateChatSessionNo(in *chat.GenerateChatSessionNoReq) (*chat.GenerateChatSessionNoResp, error) {
	sessionNo := ""
	if in.IsGuest {
		session, err := l.findGuestTransientSession(in.GetMerchantId(), in.GetUserId())
		if err != nil {
			return &chat.GenerateChatSessionNoResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		if session != nil {
			sessionNo = session.GetSessionNo()
		} else {
			sn, err := l.svcCtx.GenerateNo(l.ctx, "GCS")
			if err != nil {
				return nil, err
			}
			sessionNo = sn
		}
	} else {
		session, err := l.svcCtx.ChatSessionModel.FindByUser(l.ctx, in.MerchantId, in.UserId)
		if err == nil {
			sessionNo = session.SessionNo
		} else if err != models.ErrNotFound {
			return &chat.GenerateChatSessionNoResp{Base: helper.ErrResp(500, err.Error())}, nil
		} else {
			sn, err := l.svcCtx.GenerateNo(l.ctx, "CS")
			if err != nil {
				return nil, err
			}
			sessionNo = sn
		}
	}
	if sessionNo == "" {
		return &chat.GenerateChatSessionNoResp{Base: helper.ErrResp(500, "session no is empty")}, nil
	} else {
		return &chat.GenerateChatSessionNoResp{Base: helper.OkResp(), SessionNo: sessionNo}, nil
	}
}

func (l *GenerateChatSessionNoLogic) findGuestTransientSession(merchantID, userID int64) (*chat.ChatSession, error) {
	if merchantID <= 0 || userID <= 0 {
		return nil, nil
	}
	var cursor int64
	for {
		list, hasNext, nextCursor, err := ih.PageTransientSessions(
			l.ctx,
			l.svcCtx.BusRedis,
			merchantID,
			userID,
			0,
			0,
			cursor,
			200,
		)
		if err != nil {
			return nil, err
		}
		for _, session := range list {
			if session.Status != int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
				return ih.ToProtoSession(session, true), nil
			}
		}
		if !hasNext {
			return nil, nil
		}
		cursor = nextCursor
	}
}

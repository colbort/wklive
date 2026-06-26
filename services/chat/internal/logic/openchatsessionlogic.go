package logic

import (
	"context"
	"database/sql"
	"strings"

	"wklive/common/helper"
	"wklive/proto/chat"
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
		now := nowMillis()
		changed := false
		if session.Status == int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_CLOSED) {
			session.Status = int64(chat.ChatSessionStatus_CHAT_SESSION_STATUS_WAITING)
			session.AgentId = 0
			session.CloseTime = 0
			session.CloseReason = ""
			changed = true
		}
		if changed {
			session.UpdateTimes = now
			if err := l.svcCtx.ChatSessionModel.Update(l.ctx, session); err != nil {
				return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
			}
		}
		return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
	} else if err != models.ErrNotFound {
		sessionNo := ""
		exists := true
		for attempt := 0; attempt < sessionNoInsertAttempts; attempt++ {
			sessionNo = nextNo("CS")
			_, err := l.svcCtx.ChatSessionModel.FindOneBySessionNo(l.ctx, sessionNo)
			if err == models.ErrNotFound {
				exists = false
				break
			}
		}
		if sessionNo == "" || exists {
			return &chat.AppChatSessionResp{Base: helper.FailResp()}, nil
		}
		now := nowMillis()
		data := models.TChatSession{
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
		result, err := l.svcCtx.ChatSessionModel.Insert(l.ctx, &data)
		if err == nil {
			if id, err := result.LastInsertId(); err == nil {
				data.Id = id
			}
			return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(session)}, nil
		} else {
			return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
		}

	} else {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
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

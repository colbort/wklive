package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitChatSatisfactionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitChatSatisfactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitChatSatisfactionLogic {
	return &SubmitChatSatisfactionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交会话评价
func (l *SubmitChatSatisfactionLogic) SubmitChatSatisfaction(in *chat.SubmitChatSatisfactionReq) (*chat.AppChatSatisfactionResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if err != nil {
		return &chat.AppChatSatisfactionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatSatisfactionResp{Base: base}, nil
	}
	if strings.TrimSpace(in.GetSessionNo()) == "" {
		return &chat.AppChatSatisfactionResp{Base: badBase("session_no is required")}, nil
	}
	if in.GetScore() < 1 || in.GetScore() > 5 {
		return &chat.AppChatSatisfactionResp{Base: badBase("score must be between 1 and 5")}, nil
	}

	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AppChatSatisfactionResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AppChatSatisfactionResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.AppChatSatisfactionResp{Base: badBase("permission denied")}, nil
	}

	now := nowMillis()
	satisfaction, err := l.svcCtx.ChatSatisfactionModel.FindOneBySessionNo(l.ctx, session.SessionNo)
	if err == models.ErrNotFound {
		satisfaction = &models.TChatSatisfaction{
			MerchantId:  session.MerchantId,
			SessionNo:   session.SessionNo,
			UserId:      session.UserId,
			AgentId:     session.AgentId,
			Score:       int64(in.GetScore()),
			Content:     strings.TrimSpace(in.GetContent()),
			Tags:        strings.TrimSpace(in.GetTags()),
			CreateTimes: now,
			UpdateTimes: now,
		}
		result, err := l.svcCtx.ChatSatisfactionModel.Insert(l.ctx, satisfaction)
		if err != nil {
			return &chat.AppChatSatisfactionResp{Base: errorBase(err)}, nil
		}
		if id, err := result.LastInsertId(); err == nil {
			satisfaction.Id = id
		}
	} else if err != nil {
		return &chat.AppChatSatisfactionResp{Base: errorBase(err)}, nil
	} else {
		satisfaction.Score = int64(in.GetScore())
		satisfaction.Content = strings.TrimSpace(in.GetContent())
		satisfaction.Tags = strings.TrimSpace(in.GetTags())
		satisfaction.AgentId = session.AgentId
		satisfaction.UpdateTimes = now
		if err := l.svcCtx.ChatSatisfactionModel.Update(l.ctx, satisfaction); err != nil {
			return &chat.AppChatSatisfactionResp{Base: errorBase(err)}, nil
		}
	}

	publishEvaluationEvent(l.ctx, l.svcCtx, session, satisfaction)
	return &chat.AppChatSatisfactionResp{Base: okBase(), Data: toProtoSatisfaction(satisfaction)}, nil
}

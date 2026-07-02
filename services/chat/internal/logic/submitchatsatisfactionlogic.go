package logic

import (
	"context"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
	if strings.TrimSpace(in.GetSessionNo()) == "" {
		return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(400, "session_no is required")}, nil
	}
	if in.GetScore() < 1 || in.GetScore() > 5 {
		return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(400, "score must be between 1 and 5")}, nil
	}

	session, err := ih.GetSession(l.ctx, l.svcCtx, in.MerchantId, in.GetSessionNo(), in.IsGuest)
	if err != nil {
		return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if session.UserId != in.UserId {
		return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(400, "permission denied")}, nil
	}

	now := utils.NowMillis()
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
			return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		if id, err := result.LastInsertId(); err == nil {
			satisfaction.Id = id
		}
	} else if err != nil {
		return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(500, err.Error())}, nil
	} else {
		satisfaction.Score = int64(in.GetScore())
		satisfaction.Content = strings.TrimSpace(in.GetContent())
		satisfaction.Tags = strings.TrimSpace(in.GetTags())
		satisfaction.AgentId = session.AgentId
		satisfaction.UpdateTimes = now
		if err := l.svcCtx.ChatSatisfactionModel.Update(l.ctx, satisfaction); err != nil {
			return &chat.AppChatSatisfactionResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	}

	_ = ih.PublishMessageEvent(ih.PublishMessageEventReq{
		Ctx:       l.ctx,
		BusRedis:  l.svcCtx.BusRedis,
		Channel:   chat.ChatAdminEventChannel,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_EVALUATION_SUBMIT,
		Payload: &chat.ChatMessageEvent_Evaluation{Evaluation: &chat.ChatEvaluationPayload{
			SessionNo:    session.SessionNo,
			UserId:       session.UserId,
			AgentId:      session.AgentId,
			EvaluationId: satisfaction.Id,
			Rating:       int32(in.Score),
			Tags:         []string{in.Tags},
			Comment:      in.Content,
			Submitted:    true,
			EvaluatedAt:  now,
		}},
	})
	return &chat.AppChatSatisfactionResp{Base: helper.OkResp(), Data: ih.ToProtoSatisfaction(satisfaction)}, nil
}

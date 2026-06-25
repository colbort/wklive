package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type MarkAgentMessagesReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAgentMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAgentMessagesReadLogic {
	return &MarkAgentMessagesReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 标记客服侧已读
func (l *MarkAgentMessagesReadLogic) MarkAgentMessagesRead(in *chat.MarkAgentMessagesReadReq) (*chat.AdminMarkMessagesReadResp, error) {
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminMarkMessagesReadResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.AdminMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.AdminMarkMessagesReadResp{Base: base}, nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || operatorID <= 0 {
		return &chat.AdminMarkMessagesReadResp{Base: badBase("operator_id is required")}, nil
	}
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, merchantID, operatorID)
	if err == models.ErrNotFound {
		return &chat.AdminMarkMessagesReadResp{Base: notFoundBase("chat agent not found")}, nil
	}
	if err != nil {
		return &chat.AdminMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	if err := markRead(l.ctx, l.svcCtx, session, chat.ChatSenderType_CHAT_SENDER_TYPE_AGENT, agent.Id); err != nil {
		return &chat.AdminMarkMessagesReadResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminMarkMessagesReadResp{Base: okBase()}, nil
}

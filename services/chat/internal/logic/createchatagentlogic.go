package logic

import (
	"context"
	"fmt"
	"strings"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatAgentLogic {
	return &CreateChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建坐席
func (l *CreateChatAgentLogic) CreateChatAgent(in *chat.CreateChatAgentReq) (*chat.AdminChatAgentResp, error) {
	if in.GetMerchantId() <= 0 || in.GetChatUserId() <= 0 {
		return &chat.AdminChatAgentResp{Base: badBase("merchant_id and chat_user_id are required")}, nil
	}
	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, in.GetChatUserId())
	if err == models.ErrNotFound || user.MerchantId != in.GetMerchantId() {
		return &chat.AdminChatAgentResp{Base: notFoundBase("chat user not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}

	maxCount := int64(in.GetMaxSessionCount())
	if maxCount <= 0 {
		maxCount = defaultAgentMaxSessionCount
	}
	agentNo := strings.TrimSpace(in.GetAgentNo())
	if agentNo == "" {
		agentNo = fmt.Sprintf("AG%d", in.GetChatUserId())
	}

	now := nowMillis()
	data := &models.TChatAgent{
		MerchantId:      in.GetMerchantId(),
		ChatUserId:      in.GetChatUserId(),
		GroupId:         in.GetGroupId(),
		AgentNo:         agentNo,
		WelcomeMessage:  strings.TrimSpace(in.GetWelcomeMessage()),
		Status:          int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_OFFLINE),
		MaxSessionCount: maxCount,
		Remark:          strings.TrimSpace(in.GetRemark()),
		CreateTimes:     now,
		UpdateTimes:     now,
	}
	result, err := l.svcCtx.ChatAgentModel.Insert(l.ctx, data)
	if err != nil {
		return &chat.AdminChatAgentResp{Base: errorBase(err)}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}
	return &chat.AdminChatAgentResp{Base: okBase(), Data: toProtoAgent(data)}, nil
}

package logic

import (
	"context"
	"strings"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatQuickReplyLogic {
	return &CreateChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建快捷回复
func (l *CreateChatQuickReplyLogic) CreateChatQuickReply(in *chat.CreateChatQuickReplyReq) (*chat.AdminChatQuickReplyResp, error) {
	title := strings.TrimSpace(in.GetTitle())
	content := strings.TrimSpace(in.GetContent())
	if title == "" || content == "" {
		return &chat.AdminChatQuickReplyResp{Base: badBase("title and content are required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatQuickReplyResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	now := nowMillis()
	data := &models.TChatQuickReply{
		MerchantId:  merchantID,
		AgentId:     in.GetAgentId(),
		CategoryId:  in.GetCategoryId(),
		Title:       title,
		Content:     nullString(content),
		Enabled:     enabled,
		Sort:        int64(in.GetSort()),
		Remark:      strings.TrimSpace(in.GetRemark()),
		CreateTimes: now,
		UpdateTimes: now,
	}
	result, err := l.svcCtx.ChatQuickReplyModel.Insert(l.ctx, data)
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}
	return &chat.AdminChatQuickReplyResp{Base: okBase(), Data: toProtoChatQuickReply(data)}, nil
}

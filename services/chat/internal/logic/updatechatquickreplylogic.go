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

type UpdateChatQuickReplyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatQuickReplyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatQuickReplyLogic {
	return &UpdateChatQuickReplyLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新快捷回复
func (l *UpdateChatQuickReplyLogic) UpdateChatQuickReply(in *chat.UpdateChatQuickReplyReq) (*chat.AdminChatQuickReplyResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatQuickReplyResp{Base: badBase("id is required")}, nil
	}
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
	data, err := l.svcCtx.ChatQuickReplyModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatQuickReplyResp{Base: notFoundBase("chat quick reply not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatQuickReplyResp{Base: notFoundBase("chat quick reply not found")}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	data.AgentId = in.GetAgentId()
	data.CategoryId = in.GetCategoryId()
	data.Title = title
	data.Content = nullString(content)
	data.Enabled = enabled
	data.Sort = int64(in.GetSort())
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = nowMillis()
	if err := l.svcCtx.ChatQuickReplyModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatQuickReplyResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatQuickReplyResp{Base: okBase(), Data: toProtoChatQuickReply(data)}, nil
}

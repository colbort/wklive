package logic

import (
	"context"
	"database/sql"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/proto/common"
	ih "wklive/services/chat/internal/helper"
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
		return &chat.AdminChatQuickReplyResp{Base: helper.ErrResp(400, "title and content are required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatQuickReplyResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	now := utils.NowMillis()
	data := &models.TChatQuickReply{
		MerchantId:  merchantID,
		AgentId:     in.GetAgentId(),
		CategoryId:  in.GetCategoryId(),
		Title:       title,
		Content:     sql.NullString{String: content, Valid: true},
		Enabled:     enabled,
		Sort:        int64(in.GetSort()),
		Remark:      strings.TrimSpace(in.GetRemark()),
		CreateTimes: now,
		UpdateTimes: now,
	}
	result, err := l.svcCtx.ChatQuickReplyModel.Insert(l.ctx, data)
	if err != nil {
		return &chat.AdminChatQuickReplyResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}
	return &chat.AdminChatQuickReplyResp{Base: helper.OkResp(), Data: ih.ToProtoChatQuickReply(data)}, nil
}

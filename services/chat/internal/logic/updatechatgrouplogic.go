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

type UpdateChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatGroupLogic {
	return &UpdateChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新客服分组
func (l *UpdateChatGroupLogic) UpdateChatGroup(in *chat.UpdateChatGroupReq) (*chat.AdminChatGroupResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatGroupResp{Base: badBase("id is required")}, nil
	}
	groupName := strings.TrimSpace(in.GetGroupName())
	if groupName == "" {
		return &chat.AdminChatGroupResp{Base: badBase("group_name is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatGroupResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatGroupModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatGroupResp{Base: notFoundBase("chat group not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	data.GroupName = groupName
	data.Description = strings.TrimSpace(in.GetDescription())
	data.Enabled = enabled
	data.Sort = int64(in.GetSort())
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = nowMillis()
	if err := l.svcCtx.ChatGroupModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatGroupResp{Base: okBase(), Data: toProtoChatGroup(data)}, nil
}

package logic

import (
	"context"
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
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(400, "id is required")}, nil
	}
	groupName := strings.TrimSpace(in.GetGroupName())
	if groupName == "" {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(400, "group_name is required")}, nil
	}
	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatGroupResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatGroupModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound || data.MerchantId != merchantID {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(404, "chat group not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
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
	data.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.ChatGroupModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatGroupResp{Base: helper.OkResp(), Data: ih.ToProtoChatGroup(data)}, nil
}

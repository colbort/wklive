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

type UpdateChatCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateChatCategoryLogic {
	return &UpdateChatCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新问题分类
func (l *UpdateChatCategoryLogic) UpdateChatCategory(in *chat.UpdateChatCategoryReq) (*chat.AdminChatCategoryResp, error) {
	if in.GetId() <= 0 {
		return &chat.AdminChatCategoryResp{Base: badBase("id is required")}, nil
	}
	categoryName := strings.TrimSpace(in.GetCategoryName())
	if categoryName == "" {
		return &chat.AdminChatCategoryResp{Base: badBase("category_name is required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatCategoryResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	data, err := l.svcCtx.ChatCategoryModel.FindOne(l.ctx, in.GetId())
	if err == models.ErrNotFound {
		return &chat.AdminChatCategoryResp{Base: notFoundBase("chat category not found")}, nil
	}
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	if data.MerchantId != merchantID {
		return &chat.AdminChatCategoryResp{Base: notFoundBase("chat category not found")}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	data.ParentId = in.GetParentId()
	data.CategoryName = categoryName
	data.Enabled = enabled
	data.Sort = int64(in.GetSort())
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = nowMillis()
	if err := l.svcCtx.ChatCategoryModel.Update(l.ctx, data); err != nil {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	return &chat.AdminChatCategoryResp{Base: okBase(), Data: toProtoChatCategory(data)}, nil
}

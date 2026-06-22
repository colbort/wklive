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

type CreateChatCategoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatCategoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatCategoryLogic {
	return &CreateChatCategoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建问题分类
func (l *CreateChatCategoryLogic) CreateChatCategory(in *chat.CreateChatCategoryReq) (*chat.AdminChatCategoryResp, error) {
	categoryCode := strings.TrimSpace(in.GetCategoryCode())
	categoryName := strings.TrimSpace(in.GetCategoryName())
	if categoryCode == "" || categoryName == "" {
		return &chat.AdminChatCategoryResp{Base: badBase("category_code and category_name are required")}, nil
	}
	merchantID, base, err := merchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatCategoryResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	if _, err := l.svcCtx.ChatCategoryModel.FindOneByMerchantIdCategoryCode(l.ctx, merchantID, categoryCode); err == nil {
		return &chat.AdminChatCategoryResp{Base: badBase("category_code already exists")}, nil
	} else if err != models.ErrNotFound {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	now := nowMillis()
	data := &models.TChatCategory{
		MerchantId:   merchantID,
		ParentId:     in.GetParentId(),
		CategoryCode: categoryCode,
		CategoryName: categoryName,
		Enabled:      enabled,
		Sort:         int64(in.GetSort()),
		Remark:       strings.TrimSpace(in.GetRemark()),
		CreateTimes:  now,
		UpdateTimes:  now,
	}
	result, err := l.svcCtx.ChatCategoryModel.Insert(l.ctx, data)
	if err != nil {
		return &chat.AdminChatCategoryResp{Base: errorBase(err)}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}
	return &chat.AdminChatCategoryResp{Base: okBase(), Data: toProtoChatCategory(data)}, nil
}

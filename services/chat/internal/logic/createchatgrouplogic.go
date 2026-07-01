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

type CreateChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatGroupLogic {
	return &CreateChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建客服分组
func (l *CreateChatGroupLogic) CreateChatGroup(in *chat.CreateChatGroupReq) (*chat.AdminChatGroupResp, error) {
	groupCode := strings.TrimSpace(in.GetGroupCode())
	groupName := strings.TrimSpace(in.GetGroupName())
	if groupCode == "" || groupName == "" {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(400, "group_code and group_name are required")}, nil
	}

	merchantID, base, err := ih.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.AdminChatGroupResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	if _, err := l.svcCtx.ChatGroupModel.FindOneByMerchantIdGroupCode(l.ctx, merchantID, groupCode); err == nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(400, "group_code already exists")}, nil
	} else if err != models.ErrNotFound {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}

	now := utils.NowMillis()
	data := &models.TChatGroup{
		MerchantId:  merchantID,
		GroupCode:   groupCode,
		GroupName:   groupName,
		Description: strings.TrimSpace(in.GetDescription()),
		Enabled:     enabled,
		Sort:        int64(in.GetSort()),
		Remark:      strings.TrimSpace(in.GetRemark()),
		CreateTimes: now,
		UpdateTimes: now,
	}
	result, err := l.svcCtx.ChatGroupModel.Insert(l.ctx, data)
	if err != nil {
		return &chat.AdminChatGroupResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}

	return &chat.AdminChatGroupResp{Base: helper.OkResp(), Data: ih.ToProtoChatGroup(data)}, nil
}

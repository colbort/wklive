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
		return &chat.AdminChatGroupResp{Base: badBase("group_code and group_name are required")}, nil
	}

	merchantID, base, err := currentMerchantID(l.ctx, l.svcCtx)
	if base != nil {
		return &chat.AdminChatGroupResp{Base: base}, nil
	}
	if err != nil {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}

	if _, err := l.svcCtx.ChatGroupModel.FindOneByMerchantIdGroupCode(l.ctx, merchantID, groupCode); err == nil {
		return &chat.AdminChatGroupResp{Base: badBase("group_code already exists")}, nil
	} else if err != models.ErrNotFound {
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}

	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}

	now := nowMillis()
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
		return &chat.AdminChatGroupResp{Base: errorBase(err)}, nil
	}
	if id, err := result.LastInsertId(); err == nil {
		data.Id = id
	}

	return &chat.AdminChatGroupResp{Base: okBase(), Data: toProtoChatGroup(data)}, nil
}

func toProtoChatGroup(data *models.TChatGroup) *chat.ChatGroup {
	if data == nil {
		return nil
	}
	return &chat.ChatGroup{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		GroupCode:   data.GroupCode,
		GroupName:   data.GroupName,
		Description: data.Description,
		Enabled:     common.Enable(data.Enabled),
		Sort:        int32(data.Sort),
		Remark:      data.Remark,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

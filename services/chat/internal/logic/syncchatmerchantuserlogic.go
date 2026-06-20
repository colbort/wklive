package logic

import (
	"context"
	"fmt"
	"strings"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncChatMerchantUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncChatMerchantUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncChatMerchantUserLogic {
	return &SyncChatMerchantUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步客服商户用户
func (l *SyncChatMerchantUserLogic) SyncChatMerchantUser(in *chat.SyncChatMerchantUserReq) (*chat.SyncChatMerchantUserResp, error) {
	if in.GetMerchantId() <= 0 {
		return &chat.SyncChatMerchantUserResp{Base: badBase("merchant_id is required")}, nil
	}
	username := strings.TrimSpace(in.GetMerchantCode())
	if username == "" {
		username = fmt.Sprintf("merchant_%d", in.GetMerchantId())
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}

	now := nowMillis()
	data, err := l.svcCtx.ChatUserModel.FindOneByMerchantIdUsername(l.ctx, in.GetMerchantId(), username)
	if err != nil && err != models.ErrNotFound {
		return &chat.SyncChatMerchantUserResp{Base: errorBase(err)}, nil
	}
	if in.GetAction() == chat.ChatSyncAction_CHAT_SYNC_ACTION_DELETE {
		if err == models.ErrNotFound {
			return &chat.SyncChatMerchantUserResp{Base: okBase()}, nil
		}
		data.Enabled = int64(common.Enable_ENABLE_DISABLED)
		data.UpdateTimes = now
		if err := l.svcCtx.ChatUserModel.Update(l.ctx, data); err != nil {
			return &chat.SyncChatMerchantUserResp{Base: errorBase(err)}, nil
		}
		return &chat.SyncChatMerchantUserResp{Base: okBase(), Data: toProtoUser(data)}, nil
	}

	if err == models.ErrNotFound {
		data = &models.TChatUser{
			MerchantId:  in.GetMerchantId(),
			UserType:    int64(chat.ChatUserType_CHAT_USER_TYPE_MERCHANT),
			IsOwner:     int64(common.YesNo_YES_NO_YES),
			Username:    username,
			Nickname:    strings.TrimSpace(in.GetMerchantName()),
			Mobile:      strings.TrimSpace(in.GetContactPhone()),
			Email:       strings.TrimSpace(in.GetContactEmail()),
			Enabled:     enabled,
			Remark:      strings.TrimSpace(in.GetRemark()),
			CreateTimes: now,
			UpdateTimes: now,
		}
		result, err := l.svcCtx.ChatUserModel.Insert(l.ctx, data)
		if err != nil {
			return &chat.SyncChatMerchantUserResp{Base: errorBase(err)}, nil
		}
		if id, err := result.LastInsertId(); err == nil {
			data.Id = id
		}
		return &chat.SyncChatMerchantUserResp{Base: okBase(), Data: toProtoUser(data)}, nil
	}

	data.Nickname = strings.TrimSpace(in.GetMerchantName())
	data.Mobile = strings.TrimSpace(in.GetContactPhone())
	data.Email = strings.TrimSpace(in.GetContactEmail())
	data.Enabled = enabled
	data.Remark = strings.TrimSpace(in.GetRemark())
	data.UpdateTimes = now
	if err := l.svcCtx.ChatUserModel.Update(l.ctx, data); err != nil {
		return &chat.SyncChatMerchantUserResp{Base: errorBase(err)}, nil
	}
	return &chat.SyncChatMerchantUserResp{Base: okBase(), Data: toProtoUser(data)}, nil
}

package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/proto/common"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
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
		return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(400, "merchant_id is required")}, nil
	}
	username := strings.TrimSpace(in.GetMerchantCode())
	if username == "" {
		username = fmt.Sprintf("merchant_%d", in.GetMerchantId())
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}

	now := utils.NowMillis()
	data, err := l.svcCtx.ChatUserModel.FindOneByMerchantIdUsername(l.ctx, in.GetMerchantId(), username)
	if err != nil && err != models.ErrNotFound {
		return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if in.GetAction() == chat.ChatSyncAction_CHAT_SYNC_ACTION_DELETE {
		if err != models.ErrNotFound {
			data.Enabled = int64(common.Enable_ENABLE_DISABLED)
			data.UpdateTimes = now
			if err := l.svcCtx.ChatUserModel.Update(l.ctx, data); err != nil {
				return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
			}
		}
		if err := l.disableMerchantInfo(in.GetMerchantId(), now); err != nil {
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		return &chat.SyncChatMerchantUserResp{Base: helper.OkResp(), Data: ih.ToProtoUser(data)}, nil
	}

	password := strings.TrimSpace(in.Password)
	if err == models.ErrNotFound {
		if password == "" {
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(400, "password is required")}, nil
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		data = &models.TChatUser{
			MerchantId:  in.GetMerchantId(),
			UserType:    int64(chat.ChatUserType_CHAT_USER_TYPE_MERCHANT),
			IsOwner:     int64(common.YesNo_YES_NO_YES),
			Username:    username,
			Password:    string(hashedPassword),
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
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		if id, err := result.LastInsertId(); err == nil {
			data.Id = id
		}
		if err := l.upsertMerchantInfo(in, enabled, now); err != nil {
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		return &chat.SyncChatMerchantUserResp{Base: helper.OkResp(), Data: ih.ToProtoUser(data)}, nil
	}

	data.Nickname = strings.TrimSpace(in.GetMerchantName())
	data.Mobile = strings.TrimSpace(in.GetContactPhone())
	data.Email = strings.TrimSpace(in.GetContactEmail())
	data.Enabled = enabled
	data.Remark = strings.TrimSpace(in.GetRemark())
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
		data.Password = string(hashedPassword)
	}
	data.UpdateTimes = now
	if err := l.svcCtx.ChatUserModel.Update(l.ctx, data); err != nil {
		return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if err := l.upsertMerchantInfo(in, enabled, now); err != nil {
		return &chat.SyncChatMerchantUserResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.SyncChatMerchantUserResp{Base: helper.OkResp(), Data: ih.ToProtoUser(data)}, nil
}

func (l *SyncChatMerchantUserLogic) upsertMerchantInfo(in *chat.SyncChatMerchantUserReq, enabled int64, now int64) error {
	info, err := l.svcCtx.ChatMerchantInfoModel.FindOneByMerchantId(l.ctx, in.GetMerchantId())
	if err != nil && err != models.ErrNotFound {
		return err
	}
	if err == models.ErrNotFound {
		apiKey, apiSecret, err := generateChatMerchantKeys()
		if err != nil {
			return err
		}
		info = &models.TChatMerchantInfo{
			MerchantId:  in.GetMerchantId(),
			ApiKey:      apiKey,
			ApiSecret:   apiSecret,
			Enabled:     enabled,
			ExpireTime:  in.GetExpireTime(),
			CreateTimes: now,
			UpdateTimes: now,
		}
		_, err = l.svcCtx.ChatMerchantInfoModel.Insert(l.ctx, info)
		return err
	}

	info.Enabled = enabled
	info.ExpireTime = in.GetExpireTime()
	if info.ApiKey == "" || info.ApiSecret == "" {
		apiKey, apiSecret, err := generateChatMerchantKeys()
		if err != nil {
			return err
		}
		if info.ApiKey == "" {
			info.ApiKey = apiKey
		}
		if info.ApiSecret == "" {
			info.ApiSecret = apiSecret
		}
	}
	info.UpdateTimes = now
	return l.svcCtx.ChatMerchantInfoModel.Update(l.ctx, info)
}

func (l *SyncChatMerchantUserLogic) disableMerchantInfo(merchantId, now int64) error {
	info, err := l.svcCtx.ChatMerchantInfoModel.FindOneByMerchantId(l.ctx, merchantId)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	info.Enabled = int64(common.Enable_ENABLE_DISABLED)
	info.UpdateTimes = now
	return l.svcCtx.ChatMerchantInfoModel.Update(l.ctx, info)
}

func generateChatMerchantKeys() (string, string, error) {
	apiKey, err := randomHex("ck_", 16)
	if err != nil {
		return "", "", err
	}
	apiSecret, err := randomHex("cs_", 32)
	if err != nil {
		return "", "", err
	}
	return apiKey, apiSecret, nil
}

func randomHex(prefix string, size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(buf), nil
}

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

type AuthChatMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAuthChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthChatMerchantLogic {
	return &AuthChatMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 商户接入鉴权
func (l *AuthChatMerchantLogic) AuthChatMerchant(in *chat.AuthChatMerchantReq) (*chat.AuthChatMerchantResp, error) {
	apiKey := strings.TrimSpace(in.GetApiKey())
	apiSecret := strings.TrimSpace(in.GetApiSecret())
	if apiKey == "" || apiSecret == "" {
		return &chat.AuthChatMerchantResp{Base: badBase("api_key and api_secret are required")}, nil
	}

	merchant, err := l.svcCtx.ChatMerchantInfoModel.FindOneByApiKey(l.ctx, apiKey)
	if err == models.ErrNotFound {
		return &chat.AuthChatMerchantResp{Base: notFoundBase("chat merchant not found")}, nil
	}
	if err != nil {
		return &chat.AuthChatMerchantResp{Base: errorBase(err)}, nil
	}
	if strings.TrimSpace(merchant.ApiSecret) != apiSecret {
		return &chat.AuthChatMerchantResp{Base: badBase("invalid api_secret")}, nil
	}
	if merchant.Enabled != int64(common.Enable_ENABLE_ENABLED) {
		return &chat.AuthChatMerchantResp{Base: badBase("chat merchant is disabled")}, nil
	}
	if merchant.ExpireTime > 0 && merchant.ExpireTime <= nowMillis() {
		return &chat.AuthChatMerchantResp{Base: badBase("chat merchant is expired")}, nil
	}

	return &chat.AuthChatMerchantResp{Base: okBase(), Data: toProtoMerchant(merchant)}, nil
}

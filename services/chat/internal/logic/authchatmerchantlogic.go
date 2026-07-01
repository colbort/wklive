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
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(400, "api_key and api_secret are required")}, nil
	}

	merchant, err := l.svcCtx.ChatMerchantInfoModel.FindOneByApiKey(l.ctx, apiKey)
	if err == models.ErrNotFound {
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(404, "chat merchant not found")}, nil
	}
	if err != nil {
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if strings.TrimSpace(merchant.ApiSecret) != apiSecret {
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(400, "invalid api_secret")}, nil
	}
	if merchant.Enabled != int64(common.Enable_ENABLE_ENABLED) {
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(400, "chat merchant is disabled")}, nil
	}
	if merchant.ExpireTime > 0 && merchant.ExpireTime <= utils.NowMillis() {
		return &chat.AuthChatMerchantResp{Base: helper.ErrResp(400, "chat merchant is expired")}, nil
	}

	return &chat.AuthChatMerchantResp{Base: helper.OkResp(), Data: &chat.AuthChatMerchantData{
		MerchantId: merchant.MerchantId,
		Merchant:   ih.ToProtoMerchant(merchant),
	}}, nil
}

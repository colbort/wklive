// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_token

import (
	"context"
	"fmt"
	"strings"
	"time"
	"wklive/proto/chat"

	"chat-api/internal/jwt"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultChatTokenTTLSeconds = int64(5 * 60)
	maxChatTokenTTLSeconds     = int64(30 * 60)
	guestSessionCacheTTL       = int64(24 * 60 * 60)
)

type CreateChatTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateChatTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatTokenLogic {
	return &CreateChatTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatTokenLogic) CreateChatToken(req *types.CreateChatTokenReq) (*types.CreateChatTokenResp, error) {
	apiKey := strings.TrimSpace(req.ApiKey)
	apiSecret := strings.TrimSpace(req.ApiSecret)
	if apiKey == "" || apiSecret == "" {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 400, Msg: "apiKey and apiSecret are required"}}, nil
	}
	if req.UserId <= 0 {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 400, Msg: "userId is required"}}, nil
	}

	authResp, err := l.svcCtx.ChatAppCli.AuthChatMerchant(l.ctx, &chat.AuthChatMerchantReq{
		ApiKey:    apiKey,
		ApiSecret: apiSecret,
	})
	if err != nil {
		return nil, err
	}
	if authResp.GetBase().GetCode() != 200 {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{
			Code: authResp.GetBase().GetCode(),
			Msg:  authResp.GetBase().GetMsg(),
		}}, nil
	}

	merchantId := authResp.GetData().GetMerchantId()
	nickname := firstNonEmpty(req.Nickname, fmt.Sprintf("user-%d", req.UserId))
	avatarUrl := strings.TrimSpace(req.AvatarUrl)
	ttl := req.TtlSeconds
	if ttl <= 0 {
		ttl = defaultChatTokenTTLSeconds
	}
	if ttl > maxChatTokenTTLSeconds {
		ttl = maxChatTokenTTLSeconds
	}

	resp, err := l.svcCtx.ChatAppCli.GenerateChatSessionNo(l.ctx, &chat.GenerateChatSessionNoReq{
		MerchantId: merchantId,
		UserId:     req.UserId,
		IsGuest:    req.IsGuest,
	})
	if err != nil {
		return nil, err
	}

	expireAt := time.Now().Add(time.Duration(ttl) * time.Second).UnixMilli()
	token, err := jwt.Sign(l.svcCtx.Config.Jwt.AccessSecret, jwt.Claims{
		MerchantId: merchantId,
		UserId:     req.UserId,
		SessionNo:  resp.SessionNo,
		Nickname:   nickname,
		AvatarUrl:  avatarUrl,
		IsGuest:    req.IsGuest,
		ExpireAt:   expireAt,
	})
	if err != nil {
		l.Errorf("sign chat token failed: %v", err)
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 100001, Msg: err.Error()}}, nil
	}

	return &types.CreateChatTokenResp{
		RespBase: types.RespBase{Code: 200, Msg: "ok"},
		Data: types.ChatToken{
			ChatToken: token,
			ExpireAt:  expireAt,
			SessionNo: resp.SessionNo,
		},
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

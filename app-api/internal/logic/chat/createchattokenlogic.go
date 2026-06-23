// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/system"
	"wklive/proto/user"

	"github.com/zeromicro/go-zero/core/logx"
)

const defaultChatTokenTTLSeconds = int64(30 * 60)

type CreateChatTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type chatTokenReq struct {
	ApiKey     string `json:"apiKey"`
	ApiSecret  string `json:"apiSecret"`
	UserId     int64  `json:"userId"`
	Nickname   string `json:"nickname,omitempty"`
	AvatarUrl  string `json:"avatarUrl,omitempty"`
	IsGuest    bool   `json:"isGuest,omitempty"`
	TtlSeconds int64  `json:"ttlSeconds,omitempty"`
}

type chatTokenResp struct {
	Code int32           `json:"code"`
	Msg  string          `json:"msg"`
	Data types.ChatToken `json:"data"`
}

type chatIdentity struct {
	UserId    int64
	Nickname  string
	AvatarUrl string
	IsGuest   bool
}

func NewCreateChatTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatTokenLogic {
	return &CreateChatTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateChatTokenLogic) CreateChatToken(r *http.Request, req *types.CreateChatTokenReq) (*types.CreateChatTokenResp, error) {
	chatConfig, err := l.chatConfig()
	if err != nil {
		return nil, err
	}
	if chatConfig.GetEnabled() == common.Enable_ENABLE_DISABLED {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 403, Msg: "客服未启用"}}, nil
	}
	if strings.TrimSpace(chatConfig.GetApi()) == "" ||
		strings.TrimSpace(chatConfig.GetApiKey()) == "" ||
		strings.TrimSpace(chatConfig.GetApiSecret()) == "" ||
		strings.TrimSpace(chatConfig.GetChatUiUrl()) == "" {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 400, Msg: "客服配置未完成"}}, nil
	}

	identity, err := l.resolveIdentity(r, req)
	if err != nil {
		return nil, err
	}
	if identity.UserId <= 0 {
		return &types.CreateChatTokenResp{RespBase: types.RespBase{Code: 400, Msg: "用户身份无效"}}, nil
	}

	token, err := l.requestChatToken(chatConfig, identity)
	if err != nil {
		return nil, err
	}
	return &types.CreateChatTokenResp{
		RespBase: types.RespBase{Code: 200, Msg: "ok"},
		Data: types.ChatToken{
			ChatToken: token.ChatToken,
			ExpireAt:  token.ExpireAt,
			SessionNo: token.SessionNo,
			ChatUiUrl: strings.TrimSpace(chatConfig.ChatUiUrl),
			ChatWsUrl: strings.TrimSpace(chatConfig.ChatWsUrl),
		},
	}, nil
}

func (l *CreateChatTokenLogic) chatConfig() (*system.ChatConfig, error) {
	tenantId := int64(0)
	key := system.SysConfigType_CHAT_CONFIG
	resp, err := l.svcCtx.SystemCli.SysConfigDetail(l.ctx, &system.SysConfigDetailReq{
		TenantId:  &tenantId,
		ConfigKey: &key,
	})
	if err != nil {
		return nil, err
	}
	if resp.GetBase().GetCode() != 200 {
		return nil, fmt.Errorf("%s", resp.GetBase().GetMsg())
	}
	var cfg system.ChatConfig
	if err := json.Unmarshal([]byte(resp.GetData().GetConfigValue()), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (l *CreateChatTokenLogic) resolveIdentity(r *http.Request, req *types.CreateChatTokenReq) (chatIdentity, error) {
	if claims, ok := l.claimsFromAuthorization(r); ok {
		return l.profileIdentity(claims.UserId, claims.Username)
	}
	return l.guestIdentity(req)
}

func (l *CreateChatTokenLogic) claimsFromAuthorization(r *http.Request) (*utils.Claims, bool) {
	auth := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return nil, false
	}
	claims, err := utils.ParseToken(l.svcCtx.Config.Jwt.AccessSecret, strings.TrimSpace(auth[7:]))
	if err != nil || claims.UserId <= 0 {
		return nil, false
	}
	return claims, true
}

func (l *CreateChatTokenLogic) profileIdentity(userId int64, username string) (chatIdentity, error) {
	ctx := context.WithValue(l.ctx, utils.CtxKeyUid, userId)
	ctx = context.WithValue(ctx, utils.CtxKeyUsername, username)
	resp, err := l.svcCtx.UserCli.GetProfile(ctx, &user.GetProfileReq{})
	if err != nil {
		return chatIdentity{}, err
	}
	if resp.GetBase().GetCode() != 200 {
		return chatIdentity{}, fmt.Errorf("%s", resp.GetBase().GetMsg())
	}
	profile := resp.GetData()
	if profile == nil || profile.GetUser() == nil {
		return chatIdentity{UserId: userId, Nickname: firstNonEmpty(username, fmt.Sprintf("user-%d", userId))}, nil
	}
	u := profile.GetUser()
	return chatIdentity{
		UserId:    u.GetId(),
		Nickname:  firstNonEmpty(u.GetNickname(), u.GetUsername(), username, fmt.Sprintf("user-%d", userId)),
		AvatarUrl: u.GetAvatar(),
	}, nil
}

func (l *CreateChatTokenLogic) guestIdentity(req *types.CreateChatTokenReq) (chatIdentity, error) {
	resp, err := l.svcCtx.UserCli.GuestLogin(l.ctx, &user.GuestLoginReq{
		DeviceId:    strings.TrimSpace(req.DeviceId),
		Fingerprint: strings.TrimSpace(req.Fingerprint),
	})
	if err != nil {
		return chatIdentity{}, err
	}
	if resp.GetBase().GetCode() != 200 {
		return chatIdentity{}, fmt.Errorf("%s", resp.GetBase().GetMsg())
	}
	userId, err := strconv.ParseInt(resp.GetData().GetUserId(), 10, 64)
	if err != nil {
		return chatIdentity{}, err
	}
	return chatIdentity{
		UserId:   userId,
		Nickname: firstNonEmpty(resp.GetData().GetUsername(), "游客"),
		IsGuest:  true,
	}, nil
}

func (l *CreateChatTokenLogic) requestChatToken(cfg *system.ChatConfig, identity chatIdentity) (types.ChatToken, error) {
	payload := chatTokenReq{
		ApiKey:     strings.TrimSpace(cfg.GetApiKey()),
		ApiSecret:  strings.TrimSpace(cfg.GetApiSecret()),
		UserId:     identity.UserId,
		Nickname:   identity.Nickname,
		AvatarUrl:  identity.AvatarUrl,
		IsGuest:    identity.IsGuest,
		TtlSeconds: defaultChatTokenTTLSeconds,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return types.ChatToken{}, err
	}

	httpReq, err := http.NewRequestWithContext(l.ctx, http.MethodPost, chatTokenURL(cfg.GetApi()), bytes.NewReader(body))
	if err != nil {
		return types.ChatToken{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return types.ChatToken{}, err
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return types.ChatToken{}, fmt.Errorf("chat-api HTTP %d", httpResp.StatusCode)
	}
	var resp chatTokenResp
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return types.ChatToken{}, err
	}
	if resp.Code != 200 {
		return types.ChatToken{}, errors.New(resp.Msg)
	}
	return resp.Data, nil
}

func chatTokenURL(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		return ""
	}
	u, err := url.Parse(base)
	if err != nil {
		return strings.TrimRight(base, "/") + "/chat/internal/tokens"
	}
	path := strings.TrimRight(u.Path, "/")
	if strings.HasSuffix(path, "/chat") || path == "/chat" {
		u.Path = path + "/internal/tokens"
	} else {
		u.Path = path + "/chat/internal/tokens"
	}
	return u.String()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if v := strings.TrimSpace(value); v != "" {
			return v
		}
	}
	return ""
}

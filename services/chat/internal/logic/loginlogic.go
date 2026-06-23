package logic

import (
	"context"
	"encoding/json"
	"strings"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 登录
func (l *LoginLogic) Login(in *chat.ChatAdminLoginReq) (*chat.ChatAdminLoginResp, error) {
	username := strings.TrimSpace(in.GetUsername())
	if username == "" || strings.TrimSpace(in.GetPassword()) == "" {
		return &chat.ChatAdminLoginResp{Base: badBase("username and password are required")}, nil
	}

	user, err := l.svcCtx.ChatUserModel.FindOneByUsername(l.ctx, username)
	if err == models.ErrNotFound {
		return &chat.ChatAdminLoginResp{Base: badBase("username or password is incorrect")}, nil
	}
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: errorBase(err)}, nil
	}
	if user.Enabled != int64(common.Enable_ENABLE_ENABLED) {
		return &chat.ChatAdminLoginResp{Base: badBase("chat user is disabled")}, nil
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.GetPassword())) != nil {
		return &chat.ChatAdminLoginResp{Base: badBase("username or password is incorrect")}, nil
	}

	agent, err := l.findAgent(user)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: errorBase(err)}, nil
	}

	expand, err := buildChatTokenExpand(user, agent)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: errorBase(err)}, nil
	}
	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		l.svcCtx.Config.Jwt.AccessExpire,
		user.Id,
		user.Username,
		expand,
	)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: errorBase(err)}, nil
	}

	now := nowMillis()
	user.LastLoginTime = now
	user.UpdateTimes = now
	_ = l.svcCtx.ChatUserModel.Update(l.ctx, user)
	if agent != nil && agent.AutoOnline == int64(common.YesNo_YES_NO_YES) {
		agent.Status = int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE)
		agent.LastActiveTime = now
		agent.UpdateTimes = now
		if err := l.svcCtx.ChatAgentModel.Update(l.ctx, agent); err == nil {
			publishAgentStatusEvent(l.ctx, l.svcCtx, agent)
		}
	}

	return &chat.ChatAdminLoginResp{
		Base: okBase(),
		Data: &chat.ChatAdminLoginData{
			Token: token,
			User:  toProtoUser(user),
			Agent: toProtoAgent(agent),
		},
	}, nil
}

func (l *LoginLogic) findAgent(user *models.TChatUser) (*models.TChatAgent, error) {
	if user == nil || user.UserType != int64(chat.ChatUserType_CHAT_USER_TYPE_AGENT) {
		return nil, nil
	}
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, user.MerchantId, user.Id)
	if err == models.ErrNotFound {
		return nil, nil
	}
	return agent, err
}

func buildChatTokenExpand(user *models.TChatUser, agent *models.TChatAgent) (string, error) {
	values := map[string]any{
		"merchantId": user.MerchantId,
		"userType":   user.UserType,
		"isOwner":    user.IsOwner,
	}
	if agent != nil {
		values["agentId"] = agent.Id
	}
	bs, err := json.Marshal(values)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

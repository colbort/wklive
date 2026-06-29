package logic

import (
	"context"
	"encoding/json"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/logic/internal"
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
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(400, "username and password are required")}, nil
	}

	user, err := l.svcCtx.ChatUserModel.FindOneByUsername(l.ctx, username)
	if err == models.ErrNotFound {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(400, "username or password is incorrect")}, nil
	}
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if user.Enabled != int64(common.Enable_ENABLE_ENABLED) {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(400, "chat user is disabled")}, nil
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.GetPassword())) != nil {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(400, "username or password is incorrect")}, nil
	}

	agent, err := l.findAgent(user)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	expand, err := buildChatTokenExpand(user, agent)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	token, err := buildTokenInfo(
		l.svcCtx.Config.Jwt.AccessSecret,
		l.svcCtx.Config.Jwt.AccessExpire,
		user.Id,
		user.Username,
		expand,
	)
	if err != nil {
		return &chat.ChatAdminLoginResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	now := utils.NowMillis()
	user.LastLoginTime = now
	user.UpdateTimes = now
	_ = l.svcCtx.ChatUserModel.Update(l.ctx, user)
	if agent != nil && agent.AutoOnline == int64(common.YesNo_YES_NO_YES) {
		agent.Status = int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_ONLINE)
		agent.LastActiveTime = now
		agent.UpdateTimes = now
		if err := l.svcCtx.ChatAgentModel.Update(l.ctx, agent); err == nil {
			_ = internal.PublishMessageEvent(l.ctx, l.svcCtx, internal.PublishMessageEventReq{
				EventType:    chat.ChatEventType_CHAT_EVENT_TYPE_SYSTEM_NOTICE,
				Channel:      chat.ChatAdminEventChannel,
				Agent:        agent,
				EventMessage: "坐席已上线",
			})
		}
	}

	return &chat.ChatAdminLoginResp{
		Base: helper.OkResp(),
		Data: &chat.ChatAdminLoginData{
			Token: token,
			User:  internal.ToProtoUser(user),
			Agent: internal.ToProtoAgent(agent),
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

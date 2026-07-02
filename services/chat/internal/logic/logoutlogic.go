package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type LogoutLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLogoutLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LogoutLogic {
	return &LogoutLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 退出登录
func (l *LogoutLogic) Logout(in *chat.ChatAdminLogoutReq) (*chat.AdminCommonResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || userID <= 0 {
		return &chat.AdminCommonResp{Base: helper.ErrResp(400, "invalid login session")}, nil
	}
	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, userID)
	if err != nil {
		if err == models.ErrNotFound {
			return &chat.AdminCommonResp{Base: helper.ErrResp(400, "invalid login session")}, nil
		}
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if user.UserType == int64(chat.ChatUserType_CHAT_USER_TYPE_AGENT) {
		if err := l.autoOfflineAgent(user); err != nil {
			return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
		}
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}

func (l *LogoutLogic) autoOfflineAgent(user *models.TChatUser) error {
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdUserId(l.ctx, user.MerchantId, user.Id)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if agent.AutoOnline != int64(common.YesNo_YES_NO_YES) {
		return nil
	}
	now := utils.NowMillis()
	agent.Status = int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_OFFLINE)
	agent.LastActiveTime = now
	agent.UpdateTimes = now
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, agent); err != nil {
		return err
	}
	_ = ih.PublishMessageEvent(ih.PublishMessageEventReq{
		Ctx:       l.ctx,
		BusRedis:  l.svcCtx.BusRedis,
		Channel:   chat.ChatAdminEventChannel,
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_AGENT_LEAVE,
		Payload: &chat.ChatWsResponse_Agent{Agent: &chat.ChatAgentPayload{
			SessionNo:     "",
			AgentId:       agent.Id,
			AgentName:     user.Nickname,
			AgentAvatar:   user.AvatarUrl,
			AgentStatus:   chat.ChatAgentStatus(agent.Status),
			AssignType:    chat.ChatAssignType_CHAT_ASSIGN_TYPE_UNKNOWN,
			SessionStatus: chat.ChatSessionStatus(chat.ChatSessionStatus_CHAT_SESSION_STATUS_UNKNOWN),
			Remark:        "auto offline",
			ActionTime:    now,
		}},
	})
	return nil
}

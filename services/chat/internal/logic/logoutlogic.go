package logic

import (
	"context"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
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
		return &chat.AdminCommonResp{Base: badBase("invalid login session")}, nil
	}
	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, userID)
	if err != nil {
		if err == models.ErrNotFound {
			return &chat.AdminCommonResp{Base: badBase("invalid login session")}, nil
		}
		return &chat.AdminCommonResp{Base: errorBase(err)}, nil
	}
	if user.UserType == int64(chat.ChatUserType_CHAT_USER_TYPE_AGENT) {
		if err := l.autoOfflineAgent(user); err != nil {
			return &chat.AdminCommonResp{Base: errorBase(err)}, nil
		}
	}
	return &chat.AdminCommonResp{Base: okBase()}, nil
}

func (l *LogoutLogic) autoOfflineAgent(user *models.TChatUser) error {
	agent, err := l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, user.MerchantId, user.Id)
	if err == models.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if agent.AutoOnline != int64(common.YesNo_YES_NO_YES) {
		return nil
	}
	now := nowMillis()
	agent.Status = int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_OFFLINE)
	agent.LastActiveTime = now
	agent.UpdateTimes = now
	if err := l.svcCtx.ChatAgentModel.Update(l.ctx, agent); err != nil {
		return err
	}
	publishAgentStatusEvent(l.ctx, l.svcCtx, agent)
	return nil
}

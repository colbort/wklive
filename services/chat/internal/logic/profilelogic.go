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

type ProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProfileLogic {
	return &ProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 当前登录用户资料
func (l *ProfileLogic) Profile(in *chat.ChatAdminProfileReq) (*chat.ChatAdminProfileResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil || userID <= 0 {
		return &chat.ChatAdminProfileResp{Base: badBase("invalid login session")}, nil
	}

	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, userID)
	if err == models.ErrNotFound {
		return &chat.ChatAdminProfileResp{Base: badBase("invalid login session")}, nil
	}
	if err != nil {
		return &chat.ChatAdminProfileResp{Base: errorBase(err)}, nil
	}
	if user.Enabled != int64(common.Enable_ENABLE_ENABLED) {
		return &chat.ChatAdminProfileResp{Base: badBase("chat user is disabled")}, nil
	}

	var agent *models.TChatAgent
	if user.UserType == int64(chat.ChatUserType_CHAT_USER_TYPE_AGENT) {
		agent, err = l.svcCtx.ChatAgentModel.FindOneByMerchantIdChatUserId(l.ctx, user.MerchantId, user.Id)
		if err == models.ErrNotFound {
			agent = nil
		} else if err != nil {
			return &chat.ChatAdminProfileResp{Base: errorBase(err)}, nil
		}
	}

	return &chat.ChatAdminProfileResp{
		Base:  okBase(),
		User:  toProtoUser(user),
		Agent: toProtoAgent(agent),
	}, nil
}

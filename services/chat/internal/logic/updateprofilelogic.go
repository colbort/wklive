package logic

import (
	"context"
	"strings"

	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/crypto/bcrypt"
)

type UpdateProfileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProfileLogic {
	return &UpdateProfileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新当前登录用户资料
func (l *UpdateProfileLogic) UpdateProfile(in *chat.UpdateChatAdminProfileReq) (*chat.ChatAdminProfileResp, error) {
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

	newPassword := strings.TrimSpace(in.GetNewPassword())
	if newPassword != "" {
		oldPassword := strings.TrimSpace(in.GetOldPassword())
		if oldPassword == "" {
			return &chat.ChatAdminProfileResp{Base: badBase("old_password is required")}, nil
		}
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)) != nil {
			return &chat.ChatAdminProfileResp{Base: badBase("old_password is incorrect")}, nil
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			return &chat.ChatAdminProfileResp{Base: errorBase(err)}, nil
		}
		user.Password = string(hashedPassword)
	}

	if avatarURL := strings.TrimSpace(in.GetAvatarUrl()); avatarURL != "" {
		user.AvatarUrl = avatarURL
	}

	user.UpdateTimes = nowMillis()
	if err := l.svcCtx.ChatUserModel.Update(l.ctx, user); err != nil {
		return &chat.ChatAdminProfileResp{Base: errorBase(err)}, nil
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

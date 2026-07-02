package logic

import (
	"context"
	"fmt"
	"strings"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	"wklive/proto/common"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type CreateChatAgentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateChatAgentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateChatAgentLogic {
	return &CreateChatAgentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建坐席
func (l *CreateChatAgentLogic) CreateChatAgent(in *chat.CreateChatAgentReq) (*chat.AdminChatAgentResp, error) {
	username := strings.TrimSpace(in.GetUsername())
	password := strings.TrimSpace(in.GetPassword())
	nickname := strings.TrimSpace(in.GetNickname())
	if username == "" || password == "" || nickname == "" {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(400, "username, password and nickname are required")}, nil
	}

	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if _, err := l.svcCtx.ChatUserModel.FindOneByMerchantIdUsername(l.ctx, merchantID, username); err == nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(400, "username already exists")}, nil
	} else if err != models.ErrNotFound {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}

	maxCount := int64(in.GetMaxSessionCount())
	if maxCount <= 0 {
		maxCount = ih.DefaultAgentMaxSessionCount
	}
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	autoOnline := int64(in.GetAutoOnline())
	if autoOnline == 0 {
		autoOnline = int64(common.YesNo_YES_NO_NO)
	}

	now := utils.NowMillis()
	user := &models.TChatUser{
		MerchantId:  merchantID,
		UserType:    int64(chat.ChatUserType_CHAT_USER_TYPE_AGENT),
		IsOwner:     int64(common.YesNo_YES_NO_NO),
		Username:    username,
		Password:    string(hashedPassword),
		Nickname:    nickname,
		Mobile:      strings.TrimSpace(in.GetMobile()),
		Email:       strings.TrimSpace(in.GetEmail()),
		Enabled:     enabled,
		Remark:      strings.TrimSpace(in.GetRemark()),
		CreateTimes: now,
		UpdateTimes: now,
	}
	agent := &models.TChatAgent{
		MerchantId:      merchantID,
		GroupId:         in.GetGroupId(),
		WelcomeMessage:  strings.TrimSpace(in.GetWelcomeMessage()),
		Status:          int64(chat.ChatAgentStatus_CHAT_AGENT_STATUS_OFFLINE),
		AutoOnline:      autoOnline,
		MaxSessionCount: maxCount,
		Remark:          strings.TrimSpace(in.GetRemark()),
		CreateTimes:     now,
		UpdateTimes:     now,
	}

	if err := l.createAgentWithUser(user, agent); err != nil {
		return &chat.AdminChatAgentResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminChatAgentResp{Base: helper.OkResp(), Data: ih.ToProtoAgent(agent)}, nil
}

func (l *CreateChatAgentLogic) createAgentWithUser(user *models.TChatUser, agent *models.TChatAgent) error {
	return l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		userResult, err := session.ExecCtx(
			ctx,
			"INSERT INTO t_chat_user (merchant_id,user_type,is_owner,username,password,nickname,avatar_url,mobile,email,enabled,last_login_ip,last_login_time,remark,create_times,update_times) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
			user.MerchantId,
			user.UserType,
			user.IsOwner,
			user.Username,
			user.Password,
			user.Nickname,
			user.AvatarUrl,
			user.Mobile,
			user.Email,
			user.Enabled,
			user.LastLoginIp,
			user.LastLoginTime,
			user.Remark,
			user.CreateTimes,
			user.UpdateTimes,
		)
		if err != nil {
			return err
		}
		userID, err := userResult.LastInsertId()
		if err != nil {
			return err
		}
		user.Id = userID

		if agent.AgentNo == "" {
			agent.AgentNo = fmt.Sprintf("AG%d", userID)
		}
		agent.UserId = userID
		agentResult, err := session.ExecCtx(
			ctx,
			"INSERT INTO t_chat_agent (merchant_id,user_id,agent_no,welcome_message,status,auto_online,max_session_count,current_session_count,last_active_time,group_id,remark,create_times,update_times) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)",
			agent.MerchantId,
			agent.UserId,
			agent.AgentNo,
			agent.WelcomeMessage,
			agent.Status,
			agent.AutoOnline,
			agent.MaxSessionCount,
			agent.CurrentSessionCount,
			agent.LastActiveTime,
			agent.GroupId,
			agent.Remark,
			agent.CreateTimes,
			agent.UpdateTimes,
		)
		if err != nil {
			return err
		}
		agentID, err := agentResult.LastInsertId()
		if err != nil {
			return err
		}
		agent.Id = agentID
		return nil
	})
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"
	"encoding/json"
	"strings"

	"chat-admin-api/internal/logicutil"
	"wklive/proto/chat"

	"chat-admin-api/internal/svc"
	"chat-admin-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPageChatSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatSessionsLogic {
	return &PageChatSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PageChatSessionsLogic) PageChatSessions(req *types.PageChatSessionsReq) (resp *types.PageChatSessionsResp, err error) {
	resp, err = logicutil.Proxy[types.PageChatSessionsResp](l.ctx, req, l.svcCtx.ChatAdminCli.PageChatSessions)
	if resp != nil {
		enrichSessions(resp.Data)
	}
	if err != nil || resp == nil || resp.Code != 200 || l.svcCtx.ChatAdminCli == nil {
		return resp, err
	}

	// transientResp, err := l.svcCtx.ChatAdminCli.AdminPageTransientChatSessions(l.ctx, &chat.AdminPageTransientChatSessionsReq{
	// 	MerchantId: req.MerchantId,
	// 	UserId:     req.UserId,
	// 	AgentId:    req.AgentId,
	// 	Status:     chat.ChatSessionStatus(req.Status),
	// 	Page:       &common.PageReq{Limit: req.Limit},
	// })
	// if err != nil || transientResp.GetBase().GetCode() != 200 {
	// 	return resp, nil
	// }
	// transient := transientResp.GetData()
	// if len(transient) == 0 {
	// 	return resp, nil
	// }

	// exists := make(map[string]struct{}, len(resp.Data))
	// for _, item := range resp.Data {
	// 	exists[item.SessionNo] = struct{}{}
	// }
	// merged := make([]types.ChatSession, 0, len(transient)+len(resp.Data))
	// for _, item := range transient {
	// 	if _, ok := exists[item.GetSessionNo()]; ok {
	// 		continue
	// 	}
	// 	merged = append(merged, protoSessionToType(item))
	// }
	// resp.Data = append(merged, resp.Data...)
	// resp.Total += int64(len(merged))
	return resp, nil
}

func protoSessionToType(item *chat.ChatSession) types.ChatSession {
	if item == nil {
		return types.ChatSession{}
	}
	return types.ChatSession{
		Id:               item.GetId(),
		SessionNo:        item.GetSessionNo(),
		MerchantId:       item.GetMerchantId(),
		UserId:           item.GetUserId(),
		UserNickname:     strings.TrimSpace(item.GetTitle()),
		UserAvatarUrl:    userAvatarFromProtoSession(item),
		Source:           int64(item.GetSource()),
		Status:           int64(item.GetStatus()),
		Priority:         int64(item.GetPriority()),
		AgentId:          item.GetAgentId(),
		Title:            item.GetTitle(),
		Category:         item.GetCategory(),
		LastMessage:      item.GetLastMessage(),
		LastSenderType:   int64(item.GetLastSenderType()),
		LastMessageTime:  item.GetLastMessageTime(),
		UserUnreadCount:  int64(item.GetUserUnreadCount()),
		AgentUnreadCount: int64(item.GetAgentUnreadCount()),
		CloseTime:        item.GetCloseTime(),
		CloseReason:      item.GetCloseReason(),
		ExtJson:          protoSessionExtJson(item),
		GroupId:          item.GetGroupId(),
		LastMessageNo:    item.GetLastMessageNo(),
		CreateTimes:      item.GetCreateTimes(),
		UpdateTimes:      item.GetUpdateTimes(),
	}
}

func enrichSessions(sessions []types.ChatSession) {
	for i := range sessions {
		enrichSession(&sessions[i])
	}
}

func enrichSession(session *types.ChatSession) {
	if session == nil {
		return
	}
	if strings.TrimSpace(session.UserNickname) == "" {
		session.UserNickname = strings.TrimSpace(session.Title)
	}
	if strings.TrimSpace(session.UserAvatarUrl) == "" {
		session.UserAvatarUrl = userAvatarFromExtJson(session.ExtJson)
	}
}

func userAvatarFromProtoSession(session *chat.ChatSession) string {
	if session == nil {
		return ""
	}
	if strings.TrimSpace(session.GetAvatarUrl()) != "" {
		return strings.TrimSpace(session.GetAvatarUrl())
	}
	if session.GetExtJson() == nil {
		return ""
	}
	if value := session.GetExtJson().GetFields()["userAvatarUrl"].GetStringValue(); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	if value := session.GetExtJson().GetFields()["avatar_url"].GetStringValue(); strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return ""
}

func protoSessionExtJson(session *chat.ChatSession) string {
	if session == nil || session.GetExtJson() == nil {
		return ""
	}
	bs, err := json.Marshal(session.GetExtJson().AsMap())
	if err != nil {
		return ""
	}
	return string(bs)
}

func userAvatarFromExtJson(extJson string) string {
	extJson = strings.TrimSpace(extJson)
	if extJson == "" {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(extJson), &payload); err != nil {
		return ""
	}
	value, _ := payload["userAvatarUrl"].(string)
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	value, _ = payload["avatar_url"].(string)
	return strings.TrimSpace(value)
}

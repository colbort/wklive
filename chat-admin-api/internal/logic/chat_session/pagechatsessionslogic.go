// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat_session

import (
	"context"
	"encoding/json"
	"strings"
	"wklive/proto/chat"
	"wklive/proto/common"

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
	rpcResp, err := l.svcCtx.ChatAdminCli.PageChatSessions(l.ctx, &chat.PageChatSessionsReq{
		MerchantId: req.MerchantId,
		UserId:     req.UserId,
		AgentId:    req.AgentId,
		Status:     chat.ChatSessionStatus(req.Status),
		Priority:   chat.ChatSessionPriority(req.Priority),
		Category:   req.Category,
		TimeRange: &common.TimeRange{
			StartTime: req.StartTime,
			EndTime:   req.EndTime,
		},
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
	})
	if err != nil {
		return nil, err
	}
	resp = &types.PageChatSessionsResp{
		RespBase: protoRespBaseToType(rpcResp.GetBase()),
		Data:     protoSessionsToTypes(rpcResp.GetData()),
	}
	return resp, nil
}

func protoSessionsToTypes(list []*chat.ChatSession) []types.ChatSession {
	resp := make([]types.ChatSession, 0, len(list))
	for _, item := range list {
		resp = append(resp, protoSessionToType(item))
	}
	return resp
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

func protoSessionExtJson(item *chat.ChatSession) string {
	if item == nil {
		return ""
	}
	payload := map[string]interface{}{}
	if item.GetExtJson() != nil {
		for key, value := range item.GetExtJson().AsMap() {
			payload[key] = value
		}
	}
	if strings.TrimSpace(item.GetTitle()) != "" {
		setDefaultString(payload, "nickname", item.GetTitle())
		setDefaultString(payload, "userNickname", item.GetTitle())
	}
	if strings.TrimSpace(item.GetAvatarUrl()) != "" {
		setDefaultString(payload, "avatarUrl", item.GetAvatarUrl())
		setDefaultString(payload, "userAvatarUrl", item.GetAvatarUrl())
	}
	if item.GetIsGuest() {
		payload["isGuest"] = true
	}
	if len(payload) == 0 {
		return ""
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func setDefaultString(payload map[string]interface{}, key string, value string) {
	if _, ok := payload[key]; ok {
		return
	}
	payload[key] = strings.TrimSpace(value)
}

func protoRespBaseToType(base *common.RespBase) types.RespBase {
	if base == nil {
		return types.RespBase{}
	}
	return types.RespBase{
		Code:       base.GetCode(),
		Msg:        base.GetMsg(),
		Total:      base.GetTotal(),
		HasNext:    base.GetHasNext(),
		HasPrev:    base.GetHasPrev(),
		NextCursor: base.GetNextCursor(),
		PrevCursor: base.GetPrevCursor(),
	}
}

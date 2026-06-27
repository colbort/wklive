package internal

import (
	"encoding/json"
	"strconv"

	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/models"
)

func ToProtoAgent(data *models.TChatAgent) *chat.ChatAgent {
	if data == nil {
		return nil
	}
	return &chat.ChatAgent{
		Id:                  data.Id,
		MerchantId:          data.MerchantId,
		ChatUserId:          data.ChatUserId,
		AgentNo:             data.AgentNo,
		WelcomeMessage:      data.WelcomeMessage,
		Status:              chat.ChatAgentStatus(data.Status),
		AutoOnline:          common.YesNo(data.AutoOnline),
		MaxSessionCount:     int32(data.MaxSessionCount),
		CurrentSessionCount: int32(data.CurrentSessionCount),
		LastActiveTime:      data.LastActiveTime,
		Remark:              data.Remark,
		CreateTimes:         data.CreateTimes,
		UpdateTimes:         data.UpdateTimes,
		GroupId:             data.GroupId,
	}
}

func ToProtoAgents(list []*models.TChatAgent) []*chat.ChatAgent {
	resp := make([]*chat.ChatAgent, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoAgent(item))
	}
	return resp
}

func ToProtoChatGroup(data *models.TChatGroup) *chat.ChatGroup {
	if data == nil {
		return nil
	}
	return &chat.ChatGroup{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		GroupCode:   data.GroupCode,
		GroupName:   data.GroupName,
		Description: data.Description,
		Enabled:     common.Enable(data.Enabled),
		Sort:        int32(data.Sort),
		Remark:      data.Remark,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func ToProtoChatGroups(list []*models.TChatGroup) []*chat.ChatGroup {
	resp := make([]*chat.ChatGroup, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoChatGroup(item))
	}
	return resp
}

func ToProtoChatCategory(data *models.TChatCategory) *chat.ChatCategory {
	if data == nil {
		return nil
	}
	return &chat.ChatCategory{
		Id:           data.Id,
		MerchantId:   data.MerchantId,
		ParentId:     data.ParentId,
		CategoryCode: data.CategoryCode,
		CategoryName: data.CategoryName,
		GroupId:      data.GroupId,
		Enabled:      common.Enable(data.Enabled),
		Sort:         int32(data.Sort),
		Remark:       data.Remark,
		CreateTimes:  data.CreateTimes,
		UpdateTimes:  data.UpdateTimes,
	}
}

func ToProtoChatCategories(list []*models.TChatCategory) []*chat.ChatCategory {
	resp := make([]*chat.ChatCategory, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoChatCategory(item))
	}
	return resp
}

func ToProtoChatQuickReply(data *models.TChatQuickReply) *chat.ChatQuickReply {
	if data == nil {
		return nil
	}
	return &chat.ChatQuickReply{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		AgentId:     data.AgentId,
		CategoryId:  data.CategoryId,
		Title:       data.Title,
		Content:     data.Content.String,
		Enabled:     common.Enable(data.Enabled),
		Sort:        int32(data.Sort),
		Remark:      data.Remark,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func ToProtoChatQuickReplies(list []*models.TChatQuickReply) []*chat.ChatQuickReply {
	resp := make([]*chat.ChatQuickReply, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoChatQuickReply(item))
	}
	return resp
}

func ToProtoChatWorkOrder(data *models.TChatWorkOrder) *chat.ChatWorkOrder {
	if data == nil {
		return nil
	}
	return &chat.ChatWorkOrder{
		Id:            data.Id,
		MerchantId:    data.MerchantId,
		WorkOrderNo:   data.WorkOrderNo,
		SessionNo:     data.SessionNo,
		UserId:        data.UserId,
		AgentId:       data.AgentId,
		GroupId:       data.GroupId,
		Title:         data.Title,
		Content:       data.Content.String,
		ContactName:   data.ContactName,
		ContactMobile: data.ContactMobile,
		ContactEmail:  data.ContactEmail,
		Priority:      chat.ChatSessionPriority(data.Priority),
		Status:        int32(data.Status),
		HandlerId:     data.HandlerId,
		HandleResult:  data.HandleResult,
		FinishTime:    data.FinishTime,
		Remark:        data.Remark,
		CreateTimes:   data.CreateTimes,
		UpdateTimes:   data.UpdateTimes,
	}
}

func ToProtoChatWorkOrders(list []*models.TChatWorkOrder) []*chat.ChatWorkOrder {
	resp := make([]*chat.ChatWorkOrder, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoChatWorkOrder(item))
	}
	return resp
}

func ToProtoSatisfaction(data *models.TChatSatisfaction) *chat.ChatSatisfaction {
	if data == nil {
		return nil
	}
	return &chat.ChatSatisfaction{
		Id:          data.Id,
		MerchantId:  data.MerchantId,
		SessionNo:   data.SessionNo,
		UserId:      data.UserId,
		AgentId:     data.AgentId,
		Score:       int32(data.Score),
		Content:     data.Content,
		Tags:        data.Tags,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func ToProtoMerchant(data *models.TChatMerchantInfo) *chat.ChatMerchant {
	if data == nil {
		return nil
	}
	return &chat.ChatMerchant{
		MerchantId:  data.MerchantId,
		ApiKey:      data.ApiKey,
		Enabled:     common.Enable(data.Enabled),
		ExpireTime:  data.ExpireTime,
		CreateTimes: data.CreateTimes,
		UpdateTimes: data.UpdateTimes,
	}
}

func ToProtoSession(data *models.TChatSession) *chat.ChatSession {
	if data == nil {
		return nil
	}
	return &chat.ChatSession{
		Id:               data.Id,
		SessionNo:        data.SessionNo,
		MerchantId:       data.MerchantId,
		UserId:           data.UserId,
		Source:           chat.ChatSessionSource(data.Source),
		Status:           chat.ChatSessionStatus(data.Status),
		Priority:         chat.ChatSessionPriority(data.Priority),
		AgentId:          data.AgentId,
		Title:            data.Title,
		Category:         data.Category,
		LastMessage:      data.LastMessage,
		LastSenderType:   chat.ChatSenderType(data.LastSenderType),
		LastMessageTime:  data.LastMessageTime,
		UserUnreadCount:  int32(data.UserUnreadCount),
		AgentUnreadCount: int32(data.AgentUnreadCount),
		CloseTime:        data.CloseTime,
		CloseReason:      data.CloseReason,
		ExtJson:          NullStringToStruct(data.ExtJson),
		CreateTimes:      data.CreateTimes,
		UpdateTimes:      data.UpdateTimes,
		GroupId:          data.GroupId,
		LastMessageNo:    data.LastMessageNo,
	}
}

func ToProtoSessions(list []*models.TChatSession) []*chat.ChatSession {
	resp := make([]*chat.ChatSession, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoSession(item))
	}
	return resp
}

func ToProtoMessage(data *models.ChatMessage) *chat.ChatMessage {
	if data == nil {
		return nil
	}
	return &chat.ChatMessage{
		MessageNo:   data.MessageNo,
		SessionNo:   data.SessionNo,
		EventType:   chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE,
		MessageType: chat.ChatMessageType(data.MessageType),
		Sender:      ToProtoMessageSender(data),
		Receiver:    ToProtoMessageUser(data.Receiver),
		Content:     data.Content,
		Url:         data.Url,
		FileName:    data.FileName,
		FileSize:    data.FileSize,
		MimeType:    data.MimeType,
		Width:       data.Width,
		Height:      data.Height,
		Duration:    int32(data.Duration),
		Status:      chat.ChatMessageStatus(data.Status),
		AgentId:     strconv.FormatInt(MessageAgentID(data), 10),
		Extra:       MessagePayloadJSON(data.Payload),
		CreateTime:  data.CreateTimes,
		UpdateTime:  data.UpdateTimes,
	}
}

func ToProtoMessages(list []*models.ChatMessage) []*chat.ChatMessage {
	resp := make([]*chat.ChatMessage, 0, len(list))
	for _, item := range list {
		resp = append(resp, ToProtoMessage(item))
	}
	return resp
}

func ToProtoMessageSender(data *models.ChatMessage) *chat.ChatMessageUser {
	if data == nil {
		return nil
	}
	return ToProtoMessageUser(data.Sender)
}

func ToProtoMessageUser(data *models.ChatMessageUser) *chat.ChatMessageUser {
	if data == nil {
		return nil
	}
	return &chat.ChatMessageUser{
		Id:        data.Id,
		Type:      chat.ChatSenderType(data.Type),
		Nickname:  data.Nickname,
		AvatarUrl: data.AvatarUrl,
	}
}

func MessagePayloadJSON(payload map[string]interface{}) string {
	if len(payload) == 0 {
		return ""
	}
	bs, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	return string(bs)
}

func ToProtoUser(data *models.TChatUser) *chat.ChatUser {
	if data == nil {
		return nil
	}
	return &chat.ChatUser{
		Id:            data.Id,
		MerchantId:    data.MerchantId,
		UserType:      chat.ChatUserType(data.UserType),
		IsOwner:       common.YesNo(data.IsOwner),
		Username:      data.Username,
		Nickname:      data.Nickname,
		AvatarUrl:     data.AvatarUrl,
		Mobile:        data.Mobile,
		Email:         data.Email,
		Enabled:       common.Enable(data.Enabled),
		LastLoginTime: data.LastLoginTime,
		Remark:        data.Remark,
		CreateTimes:   data.CreateTimes,
		UpdateTimes:   data.UpdateTimes,
		Password:      data.Password,
		LastLoginIp:   data.LastLoginIp,
	}
}

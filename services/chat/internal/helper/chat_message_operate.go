package helper

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
)

func OperateMessage(ctx context.Context, svcCtx *svc.ServiceContext, in *chat.ChatMessageOperatePayload, operatorType chat.ChatSenderType, isGuest bool) *common.RespBase {
	if svcCtx == nil || in == nil {
		return helper.ErrResp(400, "message operate payload is required")
	}
	if strings.TrimSpace(in.GetSessionNo()) == "" || strings.TrimSpace(in.GetMessageNo()) == "" {
		return helper.ErrResp(400, "session_no and message_no are required")
	}

	merchantId, err := utils.GetMerchantIdFromMd(ctx)
	if err != nil || merchantId <= 0 {
		merchantId, _ = utils.GetMerchantIdFromCtx(ctx)
	}
	if merchantId <= 0 {
		return helper.ErrResp(400, "merchant_id is required")
	}

	now := utils.NowMillis()
	in.SessionNo = strings.TrimSpace(in.GetSessionNo())
	in.MessageNo = strings.TrimSpace(in.GetMessageNo())
	in.OperatorType = operatorType
	if in.OperatedAt == 0 {
		in.OperatedAt = now
	}
	if in.OperateType == chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_UNKNOWN {
		return helper.ErrResp(400, "operate_type is required")
	}
	if in.OperateType == chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_DELETE &&
		in.DeleteScope == chat.ChatMessageDeleteScope_CHAT_MESSAGE_DELETE_SCOPE_UNKNOWN {
		in.DeleteScope = chat.ChatMessageDeleteScope_CHAT_MESSAGE_DELETE_SCOPE_BOTH
	}

	status, eventType, eventPayload, err := messageOperateEvent(in)
	if err != nil {
		return helper.ErrResp(400, err.Error())
	}
	if shouldPersistMessageOperate(in) {
		if err := updateMessageOperateStatus(ctx, svcCtx, in, merchantId, status, isGuest); err != nil {
			return helper.ErrResp(500, err.Error())
		}
	}
	if err := publishMessageOperate(ctx, svcCtx, in, eventType, eventPayload); err != nil {
		return helper.ErrResp(500, err.Error())
	}
	return helper.OkResp()
}

func updateMessageOperateStatus(ctx context.Context, svcCtx *svc.ServiceContext, in *chat.ChatMessageOperatePayload, merchantId int64, status chat.ChatMessageStatus, isGuest bool) error {
	if isGuest {
		return UpdateTransientMessageStatus(ctx, svcCtx.BusRedis, merchantId, in.GetSessionNo(), in.GetMessageNo(), status, in.GetOperatedAt())
	}
	model := svcCtx.ChatMessageFactory.New(merchantId)
	if model == nil {
		return fmt.Errorf("chat message model is not configured")
	}
	return model.UpdateStatus(ctx, merchantId, in.GetMessageNo(), int64(status), in.GetOperatedAt())
}

func messageOperateEvent(in *chat.ChatMessageOperatePayload) (chat.ChatMessageStatus, chat.ChatEventType, *chat.ChatWsResponse_MessageOperate, error) {
	payload := &chat.ChatWsResponse_MessageOperate{MessageOperate: in}
	switch in.GetOperateType() {
	case chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_RECALL:
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_RECALLED, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL, payload, nil
	case chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_DELETE:
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_DELETED, chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE, payload, nil
	default:
		return chat.ChatMessageStatus_CHAT_MESSAGE_STATUS_UNKNOWN, chat.ChatEventType_CHAT_EVENT_TYPE_UNSPECIFIED, nil, fmt.Errorf("unsupported operate_type")
	}
}

func shouldPersistMessageOperate(in *chat.ChatMessageOperatePayload) bool {
	if in.GetOperateType() == chat.ChatMessageOperateType_CHAT_MESSAGE_OPERATE_TYPE_RECALL {
		return true
	}
	return in.GetDeleteScope() == chat.ChatMessageDeleteScope_CHAT_MESSAGE_DELETE_SCOPE_BOTH
}

func publishMessageOperate(ctx context.Context, svcCtx *svc.ServiceContext, in *chat.ChatMessageOperatePayload, eventType chat.ChatEventType, payload *chat.ChatWsResponse_MessageOperate) error {
	if svcCtx.BusRedis == nil {
		return nil
	}
	switch eventType {
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_RECALL:
		if err := PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAppEventChannel, PublishEventMessageRecall, payload); err != nil {
			return err
		}
		return PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAdminEventChannel, PublishEventMessageRecall, payload)
	case chat.ChatEventType_CHAT_EVENT_TYPE_MESSAGE_DELETE:
		if in.GetDeleteScope() == chat.ChatMessageDeleteScope_CHAT_MESSAGE_DELETE_SCOPE_SELF {
			if in.GetOperatorType() == chat.ChatSenderType_CHAT_SENDER_TYPE_USER {
				return PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAppEventChannel, PublishEventMessageDelete, payload)
			}
			return PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAdminEventChannel, PublishEventMessageDelete, payload)
		}
		if err := PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAppEventChannel, PublishEventMessageDelete, payload); err != nil {
			return err
		}
		return PublishMessageEvent(ctx, svcCtx.BusRedis, chat.ChatAdminEventChannel, PublishEventMessageDelete, payload)
	default:
		return nil
	}
}

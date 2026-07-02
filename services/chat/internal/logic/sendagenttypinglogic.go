package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/encoding/protojson"
)

type SendAgentTypingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSendAgentTypingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendAgentTypingLogic {
	return &SendAgentTypingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 发送用户输入状态
func (l *SendAgentTypingLogic) SendAgentTyping(in *chat.SendAgentTypingReq) (*chat.AdminCommonResp, error) {
	event := &chat.ChatWsResponse{
		EventType: chat.ChatEventType_CHAT_EVENT_TYPE_TYPING,
		CreatedAt: utils.NowMillis(),
		Payload:   &chat.ChatWsResponse_Typing{Typing: in.Typing},
	}
	payload, err := protojson.MarshalOptions{UseProtoNames: false}.Marshal(event)
	if err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	if _, err := l.svcCtx.BusRedis.PublishCtx(l.ctx, chat.ChatAppEventChannel, string(payload)); err != nil {
		return &chat.AdminCommonResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AdminCommonResp{Base: helper.OkResp()}, nil
}

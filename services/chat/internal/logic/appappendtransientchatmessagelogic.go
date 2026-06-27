package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type AppAppendTransientChatMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppAppendTransientChatMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppAppendTransientChatMessageLogic {
	return &AppAppendTransientChatMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 追加游客临时消息并更新会话摘要
func (l *AppAppendTransientChatMessageLogic) AppAppendTransientChatMessage(in *chat.AppAppendTransientChatMessageReq) (*chat.AppChatMessageResp, error) {
	msg, err := internal.AppendTransientMessage(l.ctx, l.svcCtx.BusRedis, in.GetMerchantId(), in.GetMessage(), in.GetSession(), in.GetTtlSeconds())
	if err != nil {
		return &chat.AppChatMessageResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.AppChatMessageResp{Base: helper.OkResp(), Data: msg}, nil
}

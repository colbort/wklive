package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOpenChatSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOpenChatSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOpenChatSessionLogic {
	return &GetOpenChatSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户未关闭会话
func (l *GetOpenChatSessionLogic) GetOpenChatSession(in *chat.GetOpenChatSessionReq) (*chat.InternalChatSessionResp, error) {
	data, err := l.svcCtx.ChatSessionModel.FindOpenByUser(l.ctx, in.GetMerchantId(), in.GetUserId())
	if err == models.ErrNotFound {
		return &chat.InternalChatSessionResp{Base: notFoundBase("chat session not found")}, nil
	}
	if err != nil {
		return &chat.InternalChatSessionResp{Base: errorBase(err)}, nil
	}
	return &chat.InternalChatSessionResp{Base: okBase(), Data: toProtoSession(data)}, nil
}

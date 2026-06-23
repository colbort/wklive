package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatSessionByUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatSessionByUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatSessionByUserLogic {
	return &GetChatSessionByUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按商户和用户查询会话
func (l *GetChatSessionByUserLogic) GetChatSessionByUser(in *chat.GetChatSessionByUserReq) (*chat.AppChatSessionResp, error) {
	if err := validateMerchantUser(in.GetMerchantId(), in.GetUserId()); err != nil {
		return &chat.AppChatSessionResp{Base: badBase(err.Error())}, nil
	}
	data, err := l.svcCtx.ChatSessionModel.FindLatestByUser(l.ctx, in.GetMerchantId(), in.GetUserId())
	if err == models.ErrNotFound {
		return &chat.AppChatSessionResp{Base: notFoundBase("chat session not found")}, nil
	}
	if err != nil {
		return &chat.AppChatSessionResp{Base: errorBase(err)}, nil
	}
	return &chat.AppChatSessionResp{Base: okBase(), Data: toProtoSession(data)}, nil
}

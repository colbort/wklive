package applogic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
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
func (l *GetChatSessionByUserLogic) GetChatSessionByUser(in *chat.GetChatSessionByUserReq) (*chat.UserChatSessionResp, error) {
	data, err := l.svcCtx.ChatSessionModel.FindLatestByUser(l.ctx, in.GetMerchantId(), in.GetUserId())
	if err == models.ErrNotFound {
		return &chat.UserChatSessionResp{Base: helper.ErrResp(404, "chat session not found")}, nil
	}
	if err != nil {
		return &chat.UserChatSessionResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.UserChatSessionResp{Base: helper.OkResp(), Data: ih.ToProtoSession(data, false)}, nil
}

// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package chat

import (
	"context"
	"wklive/common/utils"

	"chat-api/internal/logicutil"
	"chat-api/internal/svc"
	"chat-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListMyChatMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListMyChatMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListMyChatMessagesLogic {
	return &ListMyChatMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListMyChatMessagesLogic) ListMyChatMessages(req *types.ListChatMessagesReq) (resp *types.ListChatMessagesResp, err error) {
	userId, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}
	merchantId, err := utils.GetMerchantIdFromCtx(l.ctx)
	if err != nil {
		return nil, err
	}

	proxyReq := struct {
		*types.ListChatMessagesReq
		MerchantId int64
		UserId     int64
	}{
		ListChatMessagesReq: req,
		MerchantId:          merchantId,
		UserId:              userId,
	}
	return logicutil.Proxy[types.ListChatMessagesResp](l.ctx, proxyReq, l.svcCtx.ChatAppCli.ListMyChatMessages)
}

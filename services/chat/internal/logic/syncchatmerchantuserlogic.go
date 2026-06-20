package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncChatMerchantUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSyncChatMerchantUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncChatMerchantUserLogic {
	return &SyncChatMerchantUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 同步客服商户用户
func (l *SyncChatMerchantUserLogic) SyncChatMerchantUser(in *chat.SyncChatMerchantUserReq) (*chat.SyncChatMerchantUserResp, error) {
	// todo: add your logic here and delete this line

	return &chat.SyncChatMerchantUserResp{}, nil
}

package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitChatSatisfactionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitChatSatisfactionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitChatSatisfactionLogic {
	return &SubmitChatSatisfactionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交会话评价
func (l *SubmitChatSatisfactionLogic) SubmitChatSatisfaction(in *chat.SubmitChatSatisfactionReq) (*chat.AppChatSatisfactionResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AppChatSatisfactionResp{}, nil
}

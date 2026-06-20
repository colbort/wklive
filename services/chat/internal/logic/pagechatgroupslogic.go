package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatGroupsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatGroupsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatGroupsLogic {
	return &PageChatGroupsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询客服分组
func (l *PageChatGroupsLogic) PageChatGroups(in *chat.PageChatGroupsReq) (*chat.PageChatGroupsResp, error) {
	// todo: add your logic here and delete this line

	return &chat.PageChatGroupsResp{}, nil
}

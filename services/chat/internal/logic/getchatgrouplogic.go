package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatGroupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatGroupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatGroupLogic {
	return &GetChatGroupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询客服分组详情
func (l *GetChatGroupLogic) GetChatGroup(in *chat.GetChatGroupReq) (*chat.AdminChatGroupResp, error) {
	// todo: add your logic here and delete this line

	return &chat.AdminChatGroupResp{}, nil
}

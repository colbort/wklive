package logic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatUserByIdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatUserByIdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatUserByIdLogic {
	return &GetChatUserByIdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取用户
func (l *GetChatUserByIdLogic) GetChatUserById(in *chat.GetChatUserByIdReq) (*chat.GetChatUserByIdResp, error) {
	user, err := l.svcCtx.ChatUserModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return &chat.GetChatUserByIdResp{Base: helper.FailResp()}, nil
	} else {
		return &chat.GetChatUserByIdResp{
			Base: helper.OkResp(),
			Data: ih.ToProtoUser(user),
		}, nil
	}
}

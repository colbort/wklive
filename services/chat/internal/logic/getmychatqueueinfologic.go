package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyChatQueueInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyChatQueueInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyChatQueueInfoLogic {
	return &GetMyChatQueueInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询我的排队信息
func (l *GetMyChatQueueInfoLogic) GetMyChatQueueInfo(in *chat.GetMyChatQueueInfoReq) (*chat.ChatQueueInfoResp, error) {
	merchantID, userID, base, err := chatAppIdentityFromMetadata(l.ctx)
	if base != nil {
		return &chat.ChatQueueInfoResp{Base: base}, nil
	}
	if err != nil {
		return &chat.ChatQueueInfoResp{Base: errorBase(err)}, nil
	}
	session, base, err := getSession(l.ctx, l.svcCtx, merchantID, in.GetSessionNo())
	if err != nil {
		return &chat.ChatQueueInfoResp{Base: errorBase(err)}, nil
	}
	if base != nil {
		return &chat.ChatQueueInfoResp{Base: base}, nil
	}
	if session.UserId != userID {
		return &chat.ChatQueueInfoResp{Base: notFoundBase("chat session not found")}, nil
	}
	queue, err := toProtoQueueInfo(l.ctx, l.svcCtx, session)
	if err != nil {
		return &chat.ChatQueueInfoResp{Base: errorBase(err)}, nil
	}
	return &chat.ChatQueueInfoResp{Base: okBase(), Data: queue}, nil
}

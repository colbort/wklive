package logic

import (
	"context"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatAgentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatAgentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatAgentsLogic {
	return &PageChatAgentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询坐席
func (l *PageChatAgentsLogic) PageChatAgents(in *chat.PageChatAgentsReq) (*chat.PageChatAgentsResp, error) {
	cursor, limit := pageInput(in.GetPage())
	list, total, err := l.svcCtx.ChatAgentModel.FindPage(l.ctx, models.ChatAgentPageFilter{
		MerchantId: in.GetMerchantId(),
		ChatUserId: in.GetChatUserId(),
		GroupId:    in.GetGroupId(),
		Status:     int64(in.GetStatus()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatAgentsResp{Base: errorBase(err)}, nil
	}
	return &chat.PageChatAgentsResp{
		Base: offsetBase(cursor, limit, len(list), total),
		Data: toProtoAgents(list),
	}, nil
}

package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/pageutil"

	"wklive/proto/chat"
	"wklive/services/chat/internal/logic/internal"
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
	merchantID, base, err := internal.MerchantIDFromMetadata(l.ctx)
	if base != nil {
		return &chat.PageChatAgentsResp{Base: base}, nil
	}
	if err != nil {
		return &chat.PageChatAgentsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	list, total, err := l.svcCtx.ChatAgentModel.FindPage(l.ctx, models.ChatAgentPageFilter{
		MerchantId: merchantID,
		ChatUserId: in.GetChatUserId(),
		GroupId:    in.GetGroupId(),
		Status:     int64(in.GetStatus()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatAgentsResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.PageChatAgentsResp{
		Base: internal.OffsetBase(cursor, limit, len(list), total),
		Data: internal.ToProtoAgents(list),
	}, nil
}

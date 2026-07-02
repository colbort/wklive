package logic

import (
	"context"
	"strings"
	"wklive/common/helper"

	"wklive/common/pageutil"
	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageChatWorkOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageChatWorkOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageChatWorkOrdersLogic {
	return &PageChatWorkOrdersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询工单
func (l *PageChatWorkOrdersLogic) PageChatWorkOrders(in *chat.PageChatWorkOrdersReq) (*chat.PageChatWorkOrdersResp, error) {
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.PageChatWorkOrdersResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	list, total, err := l.svcCtx.ChatWorkOrderModel.FindPage(l.ctx, models.ChatWorkOrderPageFilter{
		MerchantId: merchantID,
		SessionNo:  strings.TrimSpace(in.GetSessionNo()),
		UserId:     in.GetUserId(),
		AgentId:    in.GetAgentId(),
		GroupId:    in.GetGroupId(),
		Priority:   int64(in.GetPriority()),
		Status:     int64(in.GetStatus()),
		HandlerId:  in.GetHandlerId(),
		Keyword:    strings.TrimSpace(in.GetKeyword()),
		StartTime:  pageutil.TimeRangeStart(in.GetTimeRange()),
		EndTime:    pageutil.TimeRangeEnd(in.GetTimeRange()),
	}, cursor, limit)
	if err != nil {
		return &chat.PageChatWorkOrdersResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.PageChatWorkOrdersResp{
		Base: ih.OffsetBase(cursor, limit, len(list), total),
		Data: ih.ToProtoChatWorkOrders(list),
	}, nil
}

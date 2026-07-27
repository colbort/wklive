package platformlogic

import (
	"context"
	"strings"
	"wklive/services/chat/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPlatformChatMerchantsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPlatformChatMerchantsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPlatformChatMerchantsLogic {
	return &ListPlatformChatMerchantsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListPlatformChatMerchantsLogic) ListPlatformChatMerchants(in *chat.PlatformChatMerchantListReq) (*chat.PlatformChatMerchantListResp, error) {
	if base, err := helpers.PlatformScope(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &chat.PlatformChatMerchantListResp{Base: base}, nil
	}
	offset, limit := pageutil.Input(in.GetPage())
	rows, total, err := l.svcCtx.ChatMerchantModel.FindPage(l.ctx, models.ChatMerchantPageFilter{
		Keyword: strings.TrimSpace(in.GetKeyword()), Enabled: int64(in.GetEnabled()),
		MerchantCode: strings.TrimSpace(in.GetMerchantCode()),
		MerchantName: strings.TrimSpace(in.GetMerchantName()),
		ContactName:  strings.TrimSpace(in.GetContactName()),
		ContactPhone: strings.TrimSpace(in.GetContactPhone()),
		ContactEmail: strings.TrimSpace(in.GetContactEmail()),
	}, offset, limit)
	if err != nil {
		return nil, err
	}
	data := make([]*chat.PlatformChatMerchant, 0, len(rows))
	for _, row := range rows {
		data = append(data, helpers.MerchantProto(row))
	}
	return &chat.PlatformChatMerchantListResp{
		Base: helpers.OffsetBase(offset, limit, len(rows), total),
		Data: data,
	}, nil
}

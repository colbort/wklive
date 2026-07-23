package adminlogic

import (
	"context"
	"wklive/common/helper"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChatConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetChatConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChatConfigLogic {
	return &GetChatConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询chat-ui配置
func (l *GetChatConfigLogic) GetChatConfig(in *chat.GetChatConfigReq) (*chat.ChatConfigResp, error) {
	merchantID, err := ih.MerchantIDFromMetadata(l.ctx)
	if err != nil {
		return &chat.ChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	data, err := l.svcCtx.ChatMerchantInfoModel.FindOneByMerchantId(l.ctx, merchantID)
	if err == models.ErrNotFound {
		return &chat.ChatConfigResp{Base: helper.ErrResp(404, "chat merchant config not found")}, nil
	}
	if err != nil {
		return &chat.ChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	return &chat.ChatConfigResp{Base: helper.OkResp(), Data: ih.ToProtoMerchant(data)}, nil
}

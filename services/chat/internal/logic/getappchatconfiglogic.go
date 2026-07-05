package logic

import (
	"context"
	"wklive/common/helper"
	"wklive/common/utils"

	"wklive/proto/chat"
	ih "wklive/services/chat/internal/helper"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAppChatConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAppChatConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAppChatConfigLogic {
	return &GetAppChatConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取chat-ui配置
func (l *GetAppChatConfigLogic) GetAppChatConfig(in *chat.GetAppChatConfigReq) (*chat.AppChatConfigResp, error) {
	var (
		data *models.TChatMerchantInfo
		err  error
	)
	merchantId, err := utils.GetMerchantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	data, err = l.svcCtx.ChatMerchantInfoModel.FindOneByMerchantId(l.ctx, merchantId)

	if err == models.ErrNotFound {
		return &chat.AppChatConfigResp{Base: helper.ErrResp(404, "chat merchant config not found")}, nil
	}
	if err != nil {
		return &chat.AppChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	merchant := ih.ToProtoMerchant(data)
	return &chat.AppChatConfigResp{Base: helper.OkResp(), Data: &chat.ChatAppConfig{
		Title:         merchant.GetTitle(),
		UiConfig:      merchant.GetUiConfig(),
		FeatureConfig: merchant.GetFeatureConfig(),
	}}, nil
}

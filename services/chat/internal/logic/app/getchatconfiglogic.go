package applogic

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

// 获取chat-ui配置
func (l *GetChatConfigLogic) GetChatConfig(in *chat.GetAppChatConfigReq) (*chat.UserChatConfigResp, error) {
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
		return &chat.UserChatConfigResp{Base: helper.ErrResp(404, "chat merchant config not found")}, nil
	}
	if err != nil {
		return &chat.UserChatConfigResp{Base: helper.ErrResp(500, err.Error())}, nil
	}
	merchant := ih.ToProtoMerchant(data)
	return &chat.UserChatConfigResp{Base: helper.OkResp(), Data: &chat.ChatAppConfig{
		Title:         merchant.GetTitle(),
		UiConfig:      merchant.GetUiConfig(),
		FeatureConfig: merchant.GetFeatureConfig(),
	}}, nil
}

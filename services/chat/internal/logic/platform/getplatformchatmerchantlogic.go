package platformlogic

import (
	"context"
	"errors"
	"strings"
	"wklive/services/chat/internal/logic/helpers"

	"wklive/proto/chat"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformChatMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPlatformChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformChatMerchantLogic {
	return &GetPlatformChatMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPlatformChatMerchantLogic) GetPlatformChatMerchant(in *chat.PlatformChatMerchantDetailReq) (*chat.PlatformChatMerchantDetailResp, error) {
	if base, err := helpers.PlatformScope(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &chat.PlatformChatMerchantDetailResp{Base: base}, nil
	}
	var row *models.TChatMerchant
	var err error
	if in.GetId() > 0 {
		row, err = l.svcCtx.ChatMerchantModel.FindOne(l.ctx, in.GetId())
	} else {
		row, err = l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, strings.TrimSpace(in.GetMerchantCode()))
	}
	if errors.Is(err, models.ErrNotFound) {
		return &chat.PlatformChatMerchantDetailResp{Base: helpers.ParamError(l.ctx)}, nil
	}
	if err != nil {
		return nil, err
	}
	return &chat.PlatformChatMerchantDetailResp{Base: helpers.OkResp(), Data: helpers.MerchantProto(row)}, nil
}

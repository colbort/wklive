package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/chat"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysChatMerchantDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysChatMerchantDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantDeleteLogic {
	return &SysChatMerchantDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除客服商户
func (l *SysChatMerchantDeleteLogic) SysChatMerchantDelete(in *system.SysChatMerchantDeleteReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	merchant, err := l.svcCtx.ChatMerchantModel.FindOne(l.ctx, in.Id)
	if errors.Is(err, models.ErrNotFound) || merchant == nil {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.ChatMerchantModel.Delete(l.ctx, in.Id); err != nil {
		return nil, err
	}
	if err := syncChatMerchantUser(l.ctx, l.svcCtx, chat.ChatSyncAction_CHAT_SYNC_ACTION_DELETE, merchant, ""); err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/system"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SysChatMerchantUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysChatMerchantUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantUpdateLogic {
	return &SysChatMerchantUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新客服商户
func (l *SysChatMerchantUpdateLogic) SysChatMerchantUpdate(in *system.SysChatMerchantUpdateReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	if in.Id <= 0 {
		return &system.RespBase{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
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

	if in.MerchantCode != "" && in.MerchantCode != merchant.MerchantCode {
		exists, err := l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, in.MerchantCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil && exists.Id != in.Id {
			return &system.RespBase{
				Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
			}, nil
		}
		merchant.MerchantCode = in.MerchantCode
	}
	if in.MerchantName != "" {
		merchant.MerchantName = in.MerchantName
	}
	if in.Enabled != 0 {
		merchant.Enabled = commonStatusToModel(in.Enabled)
	}
	if in.ExpireTime != 0 {
		merchant.ExpireTime = in.ExpireTime
	}
	if in.ContactName != "" {
		merchant.ContactName = sql.NullString{String: in.ContactName, Valid: true}
	}
	if in.ContactPhone != "" {
		merchant.ContactPhone = sql.NullString{String: in.ContactPhone, Valid: true}
	}
	if in.ContactEmail != "" {
		merchant.ContactEmail = sql.NullString{String: in.ContactEmail, Valid: true}
	}
	if in.Remark != "" {
		merchant.Remark = sql.NullString{String: in.Remark, Valid: true}
	}
	operator, err := utils.GetUsernameFromMd(l.ctx)
	if err != nil {
		operator = ""
	}
	merchant.UpdateBy = sql.NullString{String: operator, Valid: operator != ""}
	merchant.UpdateTimes = utils.NowMillis()

	if err := l.svcCtx.ChatMerchantModel.Update(l.ctx, merchant); err != nil {
		return nil, err
	}
	action := chat.ChatSyncAction_CHAT_SYNC_ACTION_UPSERT
	if merchant.Enabled == 2 {
		action = chat.ChatSyncAction_CHAT_SYNC_ACTION_DELETE
	}
	if err := syncChatMerchantUser(l.ctx, l.svcCtx, action, merchant, in.Password); err != nil {
		return nil, err
	}
	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

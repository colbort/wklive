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

type SysChatMerchantCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSysChatMerchantCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SysChatMerchantCreateLogic {
	return &SysChatMerchantCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建客服商户
func (l *SysChatMerchantCreateLogic) SysChatMerchantCreate(in *system.SysChatMerchantCreateReq) (*system.RespBase, error) {
	if base, err := systemAdminWriteScopeResp(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &system.RespBase{
			Base: base,
		}, nil
	}

	if in.MerchantCode == "" || in.MerchantName == "" {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}

	exists, err := l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, in.MerchantCode)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if exists != nil {
		return &system.RespBase{
			Base: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}

	operator, err := utils.GetUsernameFromMd(l.ctx)
	if err != nil {
		operator = ""
	}
	now := utils.NowMillis()
	merchant := &models.SysChatMerchant{
		MerchantCode: in.MerchantCode,
		MerchantName: in.MerchantName,
		Enabled:      commonStatusToModel(in.Enabled),
		ExpireTime:   in.ExpireTime,
		ContactName:  sql.NullString{String: in.ContactName, Valid: in.ContactName != ""},
		ContactPhone: sql.NullString{String: in.ContactPhone, Valid: in.ContactPhone != ""},
		ContactEmail: sql.NullString{String: in.ContactEmail, Valid: in.ContactEmail != ""},
		Remark:       sql.NullString{String: in.Remark, Valid: in.Remark != ""},
		CreateBy:     sql.NullString{String: operator, Valid: operator != ""},
		CreateTimes:  now,
		UpdateBy:     sql.NullString{String: operator, Valid: operator != ""},
		UpdateTimes:  now,
	}
	result, err := l.svcCtx.ChatMerchantModel.Insert(l.ctx, merchant)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	merchant.Id = id
	if err := syncChatMerchantUser(l.ctx, l.svcCtx, chat.ChatSyncAction_CHAT_SYNC_ACTION_UPSERT, merchant); err != nil {
		return nil, err
	}

	return &system.RespBase{
		Base: helper.OkResp(),
	}, nil
}

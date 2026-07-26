package platformlogic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type DeletePlatformChatMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePlatformChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePlatformChatMerchantLogic {
	return &DeletePlatformChatMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePlatformChatMerchantLogic) DeletePlatformChatMerchant(in *chat.PlatformChatMerchantDeleteReq) (*chat.CommonResp, error) {
	if base, err := platformScope(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &chat.CommonResp{Base: base}, nil
	}
	merchant, err := l.svcCtx.ChatMerchantModel.FindOne(l.ctx, in.GetId())
	if errors.Is(err, models.ErrNotFound) {
		return &chat.CommonResp{Base: paramError(l.ctx)}, nil
	}
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		merchantModel := models.NewTChatMerchantModel(conn, l.svcCtx.Config.CacheRedis)
		userModel := models.NewTChatUserModel(conn, l.svcCtx.Config.CacheRedis)
		infoModel := models.NewTChatMerchantInfoModel(conn, l.svcCtx.Config.CacheRedis)
		user, err := userModel.FindOneByMerchantIdUsername(ctx, merchant.Id, merchant.MerchantCode)
		if err != nil {
			return err
		}
		user.Enabled = int64(common.Enable_ENABLE_DISABLED)
		user.UpdateTimes = now
		if err = userModel.Update(ctx, user); err != nil {
			return err
		}
		info, err := infoModel.FindOneByMerchantId(ctx, merchant.Id)
		if err != nil {
			return err
		}
		info.Enabled = int64(common.Enable_ENABLE_DISABLED)
		info.UpdateTimes = now
		if err = infoModel.Update(ctx, info); err != nil {
			return err
		}
		return merchantModel.Delete(ctx, merchant.Id)
	})
	if err != nil {
		return nil, err
	}
	return &chat.CommonResp{Base: helper.OkResp()}, nil
}

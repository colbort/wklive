package platformlogic

import (
	"context"
	"errors"
	"strings"
	"wklive/services/chat/internal/logic/helpers"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/internal/svc"
	"wklive/services/chat/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type UpdatePlatformChatMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePlatformChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePlatformChatMerchantLogic {
	return &UpdatePlatformChatMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePlatformChatMerchantLogic) UpdatePlatformChatMerchant(in *chat.PlatformChatMerchantUpdateReq) (*chat.CommonResp, error) {
	if base, err := helpers.PlatformScope(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &chat.CommonResp{Base: base}, nil
	}
	merchant, err := l.svcCtx.ChatMerchantModel.FindOne(l.ctx, in.GetId())
	if errors.Is(err, models.ErrNotFound) {
		return &chat.CommonResp{Base: helpers.ParamError(l.ctx)}, nil
	}
	if err != nil {
		return nil, err
	}
	oldCode := merchant.MerchantCode
	if code := strings.TrimSpace(in.GetMerchantCode()); code != "" {
		if code != oldCode {
			existing, findErr := l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, code)
			if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
				return nil, findErr
			}
			if existing != nil && existing.Id != merchant.Id {
				return &chat.CommonResp{Base: helpers.ParamError(l.ctx)}, nil
			}
		}
		merchant.MerchantCode = code
	}
	if name := strings.TrimSpace(in.GetMerchantName()); name != "" {
		merchant.MerchantName = name
	}
	if in.GetEnabled() != common.Enable_ENABLE_UNKNOWN {
		merchant.Enabled = int64(in.GetEnabled())
	}
	if in.GetExpireTime() != 0 {
		merchant.ExpireTime = in.GetExpireTime()
	}
	if in.GetContactName() != "" {
		merchant.ContactName = strings.TrimSpace(in.GetContactName())
	}
	if in.GetContactPhone() != "" {
		merchant.ContactPhone = strings.TrimSpace(in.GetContactPhone())
	}
	if in.GetContactEmail() != "" {
		merchant.ContactEmail = strings.TrimSpace(in.GetContactEmail())
	}
	if in.GetRemark() != "" {
		merchant.Remark = strings.TrimSpace(in.GetRemark())
	}
	operator, _ := utils.GetUsernameFromMd(l.ctx)
	merchant.UpdateBy = operator
	merchant.UpdateTimes = utils.NowMillis()
	var passwordHash []byte
	if password := strings.TrimSpace(in.GetPassword()); password != "" {
		passwordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		merchantModel := models.NewTChatMerchantModel(conn, l.svcCtx.Config.CacheRedis)
		userModel := models.NewTChatUserModel(conn, l.svcCtx.Config.CacheRedis)
		infoModel := models.NewTChatMerchantInfoModel(conn, l.svcCtx.Config.CacheRedis)
		if err := merchantModel.UpdateWithUniqueCache(ctx, merchant); err != nil {
			return err
		}
		owner, err := userModel.FindOneByMerchantIdUsername(ctx, merchant.Id, oldCode)
		if err != nil {
			return err
		}
		owner.Username = merchant.MerchantCode
		owner.Nickname = merchant.MerchantName
		owner.Mobile = merchant.ContactPhone
		owner.Email = merchant.ContactEmail
		owner.Enabled = merchant.Enabled
		owner.Remark = merchant.Remark
		owner.UpdateTimes = merchant.UpdateTimes
		if len(passwordHash) > 0 {
			owner.Password = string(passwordHash)
		}
		if err = userModel.UpdateWithUniqueCache(ctx, owner); err != nil {
			return err
		}
		info, err := infoModel.FindOneByMerchantId(ctx, merchant.Id)
		if err != nil {
			return err
		}
		info.Title = merchant.MerchantName
		info.Enabled = merchant.Enabled
		info.ExpireTime = merchant.ExpireTime
		info.UpdateTimes = merchant.UpdateTimes
		return infoModel.Update(ctx, info)
	})
	if err != nil {
		return nil, err
	}
	return &chat.CommonResp{Base: helper.OkResp()}, nil
}

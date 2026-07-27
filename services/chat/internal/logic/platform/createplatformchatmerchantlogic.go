package platformlogic

import (
	"context"
	"database/sql"
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
	"google.golang.org/protobuf/encoding/protojson"
)

type CreatePlatformChatMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePlatformChatMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePlatformChatMerchantLogic {
	return &CreatePlatformChatMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePlatformChatMerchantLogic) CreatePlatformChatMerchant(in *chat.PlatformChatMerchantCreateReq) (*chat.CommonResp, error) {
	if base, err := helpers.PlatformScope(l.ctx); err != nil {
		return nil, err
	} else if base != nil {
		return &chat.CommonResp{Base: base}, nil
	}
	code := strings.TrimSpace(in.GetMerchantCode())
	name := strings.TrimSpace(in.GetMerchantName())
	password := strings.TrimSpace(in.GetPassword())
	if code == "" || name == "" || password == "" {
		return &chat.CommonResp{Base: helpers.ParamError(l.ctx)}, nil
	}
	if _, err := l.svcCtx.ChatMerchantModel.FindOneByMerchantCode(l.ctx, code); err == nil {
		return &chat.CommonResp{Base: helpers.ParamError(l.ctx)}, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	apiKey, apiSecret, err := helpers.NewMerchantKeys()
	if err != nil {
		return nil, err
	}
	operator, _ := utils.GetUsernameFromMd(l.ctx)
	now := utils.NowMillis()
	enabled := int64(in.GetEnabled())
	if enabled == 0 {
		enabled = int64(common.Enable_ENABLE_ENABLED)
	}
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		merchantModel := models.NewTChatMerchantModel(conn, l.svcCtx.Config.CacheRedis)
		userModel := models.NewTChatUserModel(conn, l.svcCtx.Config.CacheRedis)
		infoModel := models.NewTChatMerchantInfoModel(conn, l.svcCtx.Config.CacheRedis)
		merchant := &models.TChatMerchant{
			MerchantCode: code, MerchantName: name, Enabled: enabled,
			ExpireTime: in.GetExpireTime(), ContactName: strings.TrimSpace(in.GetContactName()),
			ContactPhone: strings.TrimSpace(in.GetContactPhone()),
			ContactEmail: strings.TrimSpace(in.GetContactEmail()),
			Remark:       strings.TrimSpace(in.GetRemark()), CreateBy: operator,
			CreateTimes: now, UpdateBy: operator, UpdateTimes: now,
		}
		result, err := merchantModel.Insert(ctx, merchant)
		if err != nil {
			return err
		}
		merchantID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		if _, err = userModel.Insert(ctx, &models.TChatUser{
			MerchantId: merchantID, UserType: int64(chat.ChatUserType_CHAT_USER_TYPE_MERCHANT),
			IsOwner: int64(common.YesNo_YES_NO_YES), Username: code,
			Password: string(passwordHash), Nickname: name, Mobile: merchant.ContactPhone,
			Email: merchant.ContactEmail, Enabled: enabled, Remark: merchant.Remark,
			CreateTimes: now, UpdateTimes: now,
		}); err != nil {
			return err
		}
		theme, _ := protojson.Marshal(helpers.DefaultPlatformChatTheme())
		features, _ := protojson.Marshal(helpers.DefaultPlatformChatFeatures())
		_, err = infoModel.Insert(ctx, &models.TChatMerchantInfo{
			MerchantId: merchantID, Title: name, ApiKey: apiKey, ApiSecret: apiSecret,
			UiConfig:      sql.NullString{String: string(theme), Valid: true},
			FeatureConfig: sql.NullString{String: string(features), Valid: true},
			Enabled:       enabled, ExpireTime: in.GetExpireTime(),
			CreateTimes: now, UpdateTimes: now,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return &chat.CommonResp{Base: helper.OkResp()}, nil
}

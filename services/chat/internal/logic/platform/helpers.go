package platformlogic

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/chat"
	"wklive/proto/common"
	"wklive/services/chat/models"
)

func platformScope(ctx context.Context) (*common.RespBase, error) {
	subjectDomain, err := utils.GetSubjectDomainFromMd(ctx)
	if err != nil || subjectDomain != utils.SubjectDomainSystem {
		return permissionDenied(ctx), nil
	}
	userType, err := utils.GetUserTypeFromMd(ctx)
	if err != nil || userType != utils.SysUserTypeSystemAdmin {
		return permissionDenied(ctx), nil
	}
	return nil, nil
}

func permissionDenied(ctx context.Context) *common.RespBase {
	return helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx))
}

func paramError(ctx context.Context) *common.RespBase {
	return helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx))
}

func okResp() *common.RespBase {
	return helper.OkResp()
}

func merchantProto(v *models.TChatMerchant) *chat.PlatformChatMerchant {
	return &chat.PlatformChatMerchant{
		Id: v.Id, MerchantCode: v.MerchantCode, MerchantName: v.MerchantName,
		Enabled: common.Enable(v.Enabled), ExpireTime: v.ExpireTime,
		ContactName: v.ContactName, ContactPhone: v.ContactPhone,
		ContactEmail: v.ContactEmail, Remark: v.Remark, CreateBy: v.CreateBy,
		CreateTimes: v.CreateTimes, UpdateBy: v.UpdateBy, UpdateTimes: v.UpdateTimes,
	}
}

func offsetBase(offset, limit int64, size int, total int64) *common.RespBase {
	base := helper.OkResp()
	base.Total = total
	base.HasPrev = offset > 0
	base.HasNext = offset+int64(size) < total
	base.PrevCursor = max(0, offset-limit)
	base.NextCursor = offset + int64(size)
	return base
}

func newMerchantKeys() (string, string, error) {
	key := make([]byte, 16)
	secret := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}
	if _, err := rand.Read(secret); err != nil {
		return "", "", err
	}
	return "ck_" + hex.EncodeToString(key), "cs_" + hex.EncodeToString(secret), nil
}

func defaultPlatformChatTheme() *chat.ChatThemeConfig {
	return &chat.ChatThemeConfig{
		BackgroundColor: "#F6F8FB", PrimaryColor: "#128577",
		NoticeBarColor: "#102A43", NoticeTextColor: "#E6FFFA",
		AgentBubbleColor: "#FFFFFF", UserBubbleColor: "#128577",
	}
}

func defaultPlatformChatFeatures() *chat.ChatFeatureConfig {
	return &chat.ChatFeatureConfig{EnableCopy: true, EnableRevoke: true, EnableDelete: true}
}

package logic

import (
	"context"
	"fmt"

	"wklive/proto/chat"
	"wklive/services/system/internal/svc"
	"wklive/services/system/models"
)

func syncChatMerchantUser(ctx context.Context, svcCtx *svc.ServiceContext, action chat.ChatSyncAction, merchant *models.SysChatMerchant) error {
	if merchant == nil {
		return nil
	}
	resp, err := svcCtx.ChatInternal.SyncChatMerchantUser(ctx, &chat.SyncChatMerchantUserReq{
		Action:       action,
		MerchantId:   merchant.Id,
		MerchantCode: merchant.MerchantCode,
		MerchantName: merchant.MerchantName,
		Enabled:      commonStatusToProto(merchant.Enabled),
		ExpireTime:   merchant.ExpireTime,
		ContactName:  merchant.ContactName.String,
		ContactPhone: merchant.ContactPhone.String,
		ContactEmail: merchant.ContactEmail.String,
		Remark:       merchant.Remark.String,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Base == nil {
		return fmt.Errorf("chat merchant sync returned empty response")
	}
	if resp.Base.Code != 200 {
		return fmt.Errorf("chat merchant sync failed: code=%d msg=%s", resp.Base.Code, resp.Base.Msg)
	}
	return nil
}

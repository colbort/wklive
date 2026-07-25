package adminlogic

import (
	"context"
	"fmt"

	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/liquidity/internal/svc"
)

func validateInternalTradingUser(ctx context.Context, svcCtx *svc.ServiceContext, userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("trade_user_id is required for internal provider")
	}
	if svcCtx.UserClient == nil {
		return fmt.Errorf("user internal client is not configured")
	}
	resp, err := svcCtx.UserClient.GetInternalTradingUser(ctx, &user.GetInternalTradingUserReq{UserId: userID})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 || resp.GetData() == nil {
		return fmt.Errorf("invalid internal market-maker user: %s", resp.GetBase().GetMsg())
	}
	if resp.Data.AccountType != common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER {
		return fmt.Errorf("trade user is not an internal market-maker account")
	}
	return nil
}

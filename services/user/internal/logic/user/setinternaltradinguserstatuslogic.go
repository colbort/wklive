package userlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetInternalTradingUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetInternalTradingUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetInternalTradingUserStatusLogic {
	return &SetInternalTradingUserStatusLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *SetInternalTradingUserStatusLogic) SetInternalTradingUserStatus(in *user.SetInternalTradingUserStatusReq) (*user.CommonResp, error) {
	if in == nil || in.UserId <= 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	if in.Status != user.UserStatus_USER_STATUS_NORMAL && in.Status != user.UserStatus_USER_STATUS_DISABLED {
		return nil, fmt.Errorf("internal trading user status must be NORMAL or DISABLED")
	}
	row, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if row.AccountType != int64(common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER) {
		return nil, fmt.Errorf("user is not an internal market-maker account")
	}
	row.Status = int64(in.Status)
	if remark := strings.TrimSpace(in.Remark); remark != "" {
		row.Remark.String, row.Remark.Valid = remark, true
	}
	row.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.UserModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &user.CommonResp{Base: helper.OkResp()}, nil
}

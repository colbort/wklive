package userlogic

import (
	"context"
	"fmt"

	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetInternalTradingUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetInternalTradingUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetInternalTradingUserLogic {
	return &GetInternalTradingUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetInternalTradingUserLogic) GetInternalTradingUser(in *user.GetInternalTradingUserReq) (*user.InternalTradingUserResp, error) {
	if in == nil || in.UserId <= 0 {
		return nil, fmt.Errorf("user_id is required")
	}
	row, err := l.svcCtx.UserModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	if row.AccountType != int64(common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER) {
		return nil, fmt.Errorf("user is not an internal market-maker account")
	}
	return internalTradingUserResp(row), nil
}

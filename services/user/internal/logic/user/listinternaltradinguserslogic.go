package userlogic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/common"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"
	"wklive/services/user/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListInternalTradingUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListInternalTradingUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListInternalTradingUsersLogic {
	return &ListInternalTradingUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListInternalTradingUsersLogic) ListInternalTradingUsers(in *user.ListInternalTradingUsersReq) (*user.ListInternalTradingUsersResp, error) {
	if in == nil {
		in = &user.ListInternalTradingUsersReq{}
	}
	cursor, limit := pageutil.Input(in.Page)
	rows, total, err := l.svcCtx.UserModel.FindPage(l.ctx, models.UserPageFilter{
		TenantId:    in.TenantId,
		Keyword:     in.Keyword,
		AccountType: int64(common.UserAccountType_USER_ACCOUNT_TYPE_INTERNAL_MARKET_MAKER),
	}, cursor, limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	lastID := int64(0)
	resp := &user.ListInternalTradingUsersResp{}
	for _, row := range rows {
		lastID = row.Id
		resp.Data = append(resp.Data, internalTradingUserResp(row).Data)
	}
	resp.Base = pageutil.Base(cursor, limit, len(rows), total, lastID)
	return resp, nil
}

package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListTenantPayAccountsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayAccountsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayAccountsLogic {
	return &ListTenantPayAccountsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付账号列表
func (l *ListTenantPayAccountsLogic) ListTenantPayAccounts(in *payment.ListTenantPayAccountsReq) (*payment.ListTenantPayAccountsResp, error) {
	items, total, err := l.svcCtx.TenantPayAccountModel.FindPage(l.ctx, in.TenantId, in.PlatformId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*payment.TenantPayAccount, 0, len(items))
	for _, acc := range items {
		data = append(data, toTenantPayAccountProto(acc))
	}

	return &payment.ListTenantPayAccountsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}

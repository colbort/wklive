package adminlogic

import (
	"context"
	"errors"
	"wklive/services/payment/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/common/utils"
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
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	items, total, err := l.svcCtx.TenantPayAccountModel.FindPage(l.ctx, models.TenantPayAccountPageFilter{
		TenantId:   in.TenantId,
		PlatformId: in.PlatformId,
		Keyword:    in.Keyword,
		Enabled:    int64(in.Enabled),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*payment.TenantPayAccount, 0, len(items))
	platformIDs := make([]int64, 0, len(items))
	seenPlatformIDs := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if _, exists := seenPlatformIDs[item.PlatformId]; !exists {
			seenPlatformIDs[item.PlatformId] = struct{}{}
			platformIDs = append(platformIDs, item.PlatformId)
		}
	}
	platforms, err := l.svcCtx.PayPlatformModel.FindByIDs(l.ctx, platformIDs)
	if err != nil {
		return nil, err
	}
	platformNames := make(map[int64]string, len(platforms))
	for _, platform := range platforms {
		platformNames[platform.Id] = platform.PlatformName
	}
	for _, acc := range items {
		item := helpers.ToTenantPayAccountProto(acc)
		item.PlatformName = platformNames[acc.PlatformId]
		data = append(data, item)
	}

	return &payment.ListTenantPayAccountsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}

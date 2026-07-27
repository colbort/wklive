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

type ListTenantPayChannelsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListTenantPayChannelsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListTenantPayChannelsLogic {
	return &ListTenantPayChannelsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 租户支付通道列表
func (l *ListTenantPayChannelsLogic) ListTenantPayChannels(in *payment.ListTenantPayChannelsReq) (*payment.ListTenantPayChannelsResp, error) {
	if in.TenantId <= 0 {
		if tenantId, err := utils.GetTenantIdFromMd(l.ctx); err == nil {
			in.TenantId = tenantId
		}
	}
	channels, total, err := l.svcCtx.TenantPayChannelModel.FindPage(l.ctx, models.TenantPayChannelPageFilter{
		TenantId:   in.TenantId,
		PlatformId: in.PlatformId,
		ProductId:  in.ProductId,
		AccountId:  in.AccountId,
		Keyword:    in.Keyword,
		Enabled:    int64(in.Enabled),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(channels) > 0 {
		lastID = channels[len(channels)-1].Id
	}

	data := make([]*payment.TenantPayChannel, 0, len(channels))
	platformIDs := make([]int64, 0, len(channels))
	productIDs := make([]int64, 0, len(channels))
	accountIDs := make([]int64, 0, len(channels))
	seenPlatforms := make(map[int64]struct{}, len(channels))
	seenProducts := make(map[int64]struct{}, len(channels))
	seenAccounts := make(map[int64]struct{}, len(channels))
	for _, channel := range channels {
		if _, exists := seenPlatforms[channel.PlatformId]; !exists {
			seenPlatforms[channel.PlatformId] = struct{}{}
			platformIDs = append(platformIDs, channel.PlatformId)
		}
		if _, exists := seenProducts[channel.ProductId]; !exists {
			seenProducts[channel.ProductId] = struct{}{}
			productIDs = append(productIDs, channel.ProductId)
		}
		if _, exists := seenAccounts[channel.AccountId]; !exists {
			seenAccounts[channel.AccountId] = struct{}{}
			accountIDs = append(accountIDs, channel.AccountId)
		}
	}
	platforms, err := l.svcCtx.PayPlatformModel.FindByIDs(l.ctx, platformIDs)
	if err != nil {
		return nil, err
	}
	products, err := l.svcCtx.PayProductModel.FindByIDs(l.ctx, productIDs)
	if err != nil {
		return nil, err
	}
	accounts, err := l.svcCtx.TenantPayAccountModel.FindByIDs(l.ctx, accountIDs)
	if err != nil {
		return nil, err
	}
	platformNames := make(map[int64]string, len(platforms))
	productNames := make(map[int64]string, len(products))
	accountNames := make(map[int64]string, len(accounts))
	for _, platform := range platforms {
		platformNames[platform.Id] = platform.PlatformName
	}
	for _, product := range products {
		productNames[product.Id] = product.ProductName
	}
	for _, account := range accounts {
		accountNames[account.Id] = account.AccountName
	}
	for _, c := range channels {
		item := helpers.ToTenantPayChannelProto(c)
		item.PlatformName = platformNames[c.PlatformId]
		item.ProductName = productNames[c.ProductId]
		item.AccountName = accountNames[c.AccountId]
		data = append(data, item)
	}

	return &payment.ListTenantPayChannelsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(channels), total, lastID),
		Data: data,
	}, nil
}

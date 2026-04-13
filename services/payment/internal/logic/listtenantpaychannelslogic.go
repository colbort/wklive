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
	channels, total, err := l.svcCtx.TenantPayChannelModel.FindPage(l.ctx, in.TenantId, in.PlatformId, in.ProductId, in.AccountId, in.Keyword, int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(channels) > 0 {
		lastID = channels[len(channels)-1].Id
	}

	data := make([]*payment.TenantPayChannel, 0, len(channels))
	for _, c := range channels {
		data = append(data, toTenantPayChannelProto(c))
	}

	return &payment.ListTenantPayChannelsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(channels), total, lastID),
		Data: data,
	}, nil
}

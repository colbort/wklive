package applogic

import (
	"context"

	"wklive/common/userevent"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func publishOptionOrderChanged(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TOptionOrder) {
	if order == nil {
		return
	}
	event := userevent.NewOrderChanged(userevent.DomainOption, order.TenantId, order.UserId, order.Id, order.OrderNo)
	if err := userevent.Publish(ctx, svcCtx.UserEventPublisher, event); err != nil {
		logx.WithContext(ctx).Errorf("publish option order changed failed, orderNo=%s err=%v", order.OrderNo, err)
	}
}

// PublishOptionOrderChanged exposes the post-commit notification to
// administrative recovery flows.
func PublishOptionOrderChanged(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TOptionOrder) {
	publishOptionOrderChanged(ctx, svcCtx, order)
}

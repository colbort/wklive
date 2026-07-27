package tasklogic

import (
	"context"

	"wklive/common/userevent"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

func publishStakingOrderChanged(ctx context.Context, svcCtx *svc.ServiceContext, order *models.TStakeOrder) {
	if order == nil {
		return
	}
	event := userevent.NewOrderChanged(userevent.DomainStaking, order.TenantId, order.UserId, order.Id, order.OrderNo)
	if err := userevent.Publish(ctx, svcCtx.UserEventPublisher, event); err != nil {
		logx.WithContext(ctx).Errorf("publish staking order changed failed, orderNo=%s err=%v", order.OrderNo, err)
	}
}

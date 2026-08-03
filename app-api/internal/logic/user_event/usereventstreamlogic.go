// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package user_event

import (
	"context"
	"fmt"
	"time"

	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"
	"wklive/common/userevent"
	"wklive/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
)

type UserEventStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserEventStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserEventStreamLogic {
	return &UserEventStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserEventStreamLogic) UserEventStream(
	_ *types.UserEventStreamReq,
	client chan<- *types.UserEventStreamResp,
) error {
	tenantID, err := tenantIDFromJwtContext(l.ctx)
	if err != nil {
		return fmt.Errorf("get tenant from access token: %w", err)
	}
	if tenantID <= 0 {
		return fmt.Errorf("get tenant from access token: invalid tenant id")
	}
	userID, err := utils.GetUserIdFromCtx(l.ctx)
	if err != nil {
		return fmt.Errorf("get user from access token: %w", err)
	}
	if userID <= 0 {
		return fmt.Errorf("get user from access token: invalid user id")
	}

	events, unsubscribe := l.svcCtx.UserEventHub.Subscribe(tenantID, userID)
	defer unsubscribe()

	if !send(l.ctx, client, &types.UserEventStreamResp{
		EventType: "connected",
		ServerTs:  time.Now().UnixMilli(),
	}) {
		return nil
	}

	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return nil
		case <-heartbeat.C:
			if !send(l.ctx, client, &types.UserEventStreamResp{
				EventType: "heartbeat",
				ServerTs:  time.Now().UnixMilli(),
			}) {
				return nil
			}
		case event, ok := <-events:
			if !ok {
				return nil
			}
			if !send(l.ctx, client, toStreamResponse(event)) {
				return nil
			}
		}
	}
}

func tenantIDFromJwtContext(ctx context.Context) (int64, error) {
	return utils.GetTrustedTenantIdFromCtx(ctx)
}

func send(
	ctx context.Context,
	client chan<- *types.UserEventStreamResp,
	event *types.UserEventStreamResp,
) bool {
	select {
	case client <- event:
		return true
	case <-ctx.Done():
		return false
	}
}

func toStreamResponse(event userevent.Event) *types.UserEventStreamResp {
	return &types.UserEventStreamResp{
		EventType: event.Type,
		ServerTs:  time.Now().UnixMilli(),
		Data: &types.UserEventStreamData{
			Version:     event.Version,
			Id:          event.ID,
			EventType:   event.Type,
			Domain:      event.Domain,
			TenantId:    event.TenantID,
			UserId:      event.UserID,
			BizId:       event.BizID,
			BizNo:       event.BizNo,
			SymbolId:    event.SymbolID,
			ProductType: event.ProductType,
			OccurredAt:  event.OccurredAt,
		},
	}
}

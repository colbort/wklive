package adminlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReleaseUserKillSwitchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReleaseUserKillSwitchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReleaseUserKillSwitchLogic {
	return &ReleaseUserKillSwitchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 人工复核后解除用户 kill switch
func (l *ReleaseUserKillSwitchLogic) ReleaseUserKillSwitch(in *option.ReleaseUserKillSwitchReq) (*option.CommonResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(in.Reason)
	if forbidden || !allowed {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	if in.TenantId <= 0 || in.UserId <= 0 || operatorID <= 0 || reason == "" || len(reason) > 255 {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	releaseBlocked := false
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		controlModel := models.NewTOptionUserTradingControlModel(conn, l.svcCtx.Config.CacheRedis)
		item, err := controlModel.FindForUpdate(ctx, in.TenantId, in.UserId)
		if err != nil {
			return err
		}
		if item.KillSwitch == int64(common.YesNo_YES_NO_NO) {
			return nil
		}
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		unsafeOrders, err := orderModel.HasUnsafeKillSwitchReleaseOrders(
			ctx, in.TenantId, in.UserId,
		)
		if err != nil {
			return err
		}
		if unsafeOrders {
			releaseBlocked = true
			return nil
		}
		item.KillSwitch = int64(common.YesNo_YES_NO_NO)
		item.Reason = reason
		item.ReleasedAt = now
		item.ReleasedBy = operatorID
		item.UpdateTimes = now
		if err := controlModel.Update(ctx, item); err != nil {
			return err
		}
		eventModel := models.NewTOptionTradingControlEventModel(conn, l.svcCtx.Config.CacheRedis)
		if _, err := eventModel.Insert(ctx, &models.TOptionTradingControlEvent{
			TenantId: in.TenantId, UserId: in.UserId,
			EventType: "KILL_SWITCH_RELEASED", Reason: "USER_KILL_SWITCH",
			Detail: reason, OperatorId: operatorID, CreateTimes: now,
		}); err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, models.ErrNotFound) {
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.UserNotFound, i18n.Translate(i18n.UserNotFound, l.ctx)),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	if releaseBlocked {
		l.Errorf(
			"option kill switch release blocked by non-terminal orders, tenantId=%d userId=%d operatorId=%d",
			in.TenantId, in.UserId, operatorID,
		)
		return &option.CommonResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, i18n.Translate(i18n.OperationNotAllowed, l.ctx)),
		}, nil
	}
	l.Infof(
		"option trading control metric event=%s reason=%s tenantId=%d userId=%d operatorId=%d",
		"KILL_SWITCH_RELEASED", "USER_KILL_SWITCH", in.TenantId, in.UserId, operatorID,
	)
	return &option.CommonResp{Base: helper.OkResp()}, nil
}

package applogic

import (
	"context"
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

type ActivateKillSwitchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewActivateKillSwitchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ActivateKillSwitchLogic {
	return &ActivateKillSwitchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 激活用户 kill switch；立即阻止新单并撤销全部活动订单
func (l *ActivateKillSwitchLogic) ActivateKillSwitch(in *option.ActivateKillSwitchReq) (*option.GetUserTradingControlResp, error) {
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "USER_REQUEST"
	}
	if len(reason) > 255 {
		return &option.GetUserTradingControlResp{
			Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx)),
		}, nil
	}
	now := time.Now().Unix()
	var control *models.TOptionUserTradingControl
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		controlModel := models.NewTOptionUserTradingControlModel(conn, l.svcCtx.Config.CacheRedis)
		item, err := controlModel.EnsureForUpdate(ctx, tenantID, userID, now)
		if err != nil {
			return err
		}
		if item.KillSwitch != int64(common.YesNo_YES_NO_YES) {
			item.KillSwitch = int64(common.YesNo_YES_NO_YES)
			item.Reason = reason
			item.ActivatedAt = now
			item.ActivatedBy = userID
			item.ReleasedAt = 0
			item.ReleasedBy = 0
			item.UpdateTimes = now
			if err := controlModel.Update(ctx, item); err != nil {
				return err
			}
			if err := insertTradingControlEvent(ctx, l.svcCtx, conn, &models.TOptionTradingControlEvent{
				TenantId: tenantID, UserId: userID,
				EventType: controlEventKillActivated, Reason: controlReasonKillSwitch,
				Detail: reason, OperatorId: userID, CreateTimes: now,
			}); err != nil {
				return err
			}
		}
		control = item
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := cancelAllUserOrders(l.ctx, l.svcCtx, tenantID, userID, controlReasonKillSwitch); err != nil {
		// The switch remains active. A retry is safe and resumes cancellation.
		return nil, err
	}
	return &option.GetUserTradingControlResp{
		Base: helper.OkResp(), Data: optionUserTradingControl(control),
	}, nil
}

func cancelAllUserOrders(
	ctx context.Context, svcCtx *svc.ServiceContext, tenantID, userID int64, reason string,
) error {
	cursor := int64(0)
	for {
		orders, _, err := svcCtx.OptionOrderModel.FindPage(ctx, models.OptionOrderPageFilter{
			TenantId: tenantID, UserId: userID,
			Statuses: []int64{
				int64(option.OrderStatus_ORDER_STATUS_FUNDING),
				int64(option.OrderStatus_ORDER_STATUS_PENDING),
				int64(option.OrderStatus_ORDER_STATUS_PART_FILLED),
			},
		}, cursor, 100)
		if err != nil {
			return err
		}
		for _, order := range orders {
			cursor = order.Id
			if _, err := CancelOrderByControl(ctx, svcCtx, order.Id, reason); err != nil {
				return err
			}
		}
		if len(orders) < 100 {
			return nil
		}
	}
}

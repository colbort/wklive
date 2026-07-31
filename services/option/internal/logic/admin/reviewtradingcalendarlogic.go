package adminlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ReviewTradingCalendarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewTradingCalendarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewTradingCalendarLogic {
	return &ReviewTradingCalendarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 由独立管理员批准或拒绝日历版本
func (l *ReviewTradingCalendarLogic) ReviewTradingCalendar(in *option.ReviewTradingCalendarReq) (*option.GetTradingCalendarResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return calendarPermissionDenied(l.ctx), nil
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.CalendarId <= 0 || operatorID <= 0 || reason == "" || len(reason) > 500 {
		return calendarParamError(l.ctx), nil
	}
	now := time.Now().Unix()
	var reviewed *models.TOptionTradingCalendar
	var sessions []*models.TOptionTradingCalendarSession
	var exceptions []*models.TOptionTradingCalendarException
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, l.svcCtx.Config.CacheRedis)
		sessionModel := models.NewTOptionTradingCalendarSessionModel(conn, l.svcCtx.Config.CacheRedis)
		exceptionModel := models.NewTOptionTradingCalendarExceptionModel(conn, l.svcCtx.Config.CacheRedis)
		item, findErr := calendarModel.FindOneForUpdate(ctx, in.CalendarId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId ||
			item.Status != int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_DRAFT) ||
			item.CreatedBy == operatorID {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		latest, findErr := calendarModel.FindLatestForUpdate(ctx, item.TenantId, item.CalendarCode)
		if findErr != nil {
			return findErr
		}
		if latest.Id != item.Id {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		sessions, findErr = sessionModel.FindByCalendar(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		exceptions, findErr = exceptionModel.FindByCalendar(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		if validationErr := helpers.ValidateTradingCalendarDefinition(
			item.Timezone, item.EffectiveFrom, item.EffectiveUntil, sessions, exceptions,
		); validationErr != nil {
			return errInvalidTradingCalendar
		}
		if in.Approve {
			if item.EffectiveFrom < now || strings.TrimSpace(item.EvidenceRef) == "" {
				return errInvalidTradingCalendar
			}
			previous, previousErr := calendarModel.FindOpenEndedApprovedForUpdate(
				ctx, item.TenantId, item.CalendarCode,
			)
			if previousErr != nil && !errors.Is(previousErr, models.ErrNotFound) {
				return previousErr
			}
			if previous != nil {
				if item.EffectiveFrom <= previous.EffectiveFrom || item.Version <= previous.Version {
					return errInvalidTradingCalendar
				}
				previous.Status = int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_SUPERSEDED)
				previous.EffectiveUntil = item.EffectiveFrom
				previous.UpdateTimes = now
				if updateErr := calendarModel.Update(ctx, previous); updateErr != nil {
					return updateErr
				}
				item.SupersedesId = previous.Id
			}
			item.Status = int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_APPROVED)
		} else {
			item.Status = int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_REJECTED)
		}
		item.ReviewedBy = operatorID
		item.ReviewReason = reason
		item.ReviewedAt = now
		item.UpdateTimes = now
		if updateErr := calendarModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		reviewed = item
		return nil
	})
	if errors.Is(err, errInvalidTradingCalendar) {
		return calendarParamError(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetTradingCalendarResp{
		Base: helper.OkResp(), Data: helpers.ToTradingCalendarProto(reviewed, sessions, exceptions),
	}, nil
}

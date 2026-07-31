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
	logichelpers "wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var errInvalidTradingCalendar = errors.New("invalid option trading calendar")

type CreateTradingCalendarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateTradingCalendarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateTradingCalendarLogic {
	return &CreateTradingCalendarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建不可覆盖的交易日历版本草案
func (l *CreateTradingCalendarLogic) CreateTradingCalendar(in *option.CreateTradingCalendarReq) (*option.GetTradingCalendarResp, error) {
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
	code, validCode := logichelpers.NormalizeTradingCalendarCode(in.CalendarCode)
	changeReason := strings.TrimSpace(in.ChangeReason)
	evidenceRef := strings.TrimSpace(in.EvidenceRef)
	effectiveFrom := in.EffectiveFrom
	now := time.Now().Unix()
	if effectiveFrom <= 0 {
		effectiveFrom = now + 300
	}
	if in.TenantId <= 0 || operatorID <= 0 || !validCode || changeReason == "" || evidenceRef == "" ||
		len(changeReason) > 500 || len(evidenceRef) > 500 || effectiveFrom < now {
		return calendarParamError(l.ctx), nil
	}

	var created *models.TOptionTradingCalendar
	var createdSessions []*models.TOptionTradingCalendarSession
	var createdExceptions []*models.TOptionTradingCalendarException
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		calendarModel := models.NewTOptionTradingCalendarModel(conn, l.svcCtx.Config.CacheRedis)
		sessionModel := models.NewTOptionTradingCalendarSessionModel(conn, l.svcCtx.Config.CacheRedis)
		exceptionModel := models.NewTOptionTradingCalendarExceptionModel(conn, l.svcCtx.Config.CacheRedis)
		latest, findErr := calendarModel.FindLatestForUpdate(ctx, in.TenantId, code)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if latest != nil && latest.Status == int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_DRAFT) {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		version := int64(1)
		if latest != nil {
			version = latest.Version + 1
		}

		timezone := strings.TrimSpace(in.Timezone)
		sessions := tradingCalendarSessionsFromInput(in.Sessions)
		exceptions := tradingCalendarExceptionsFromInput(in.Exceptions)
		if in.SourceCalendarId > 0 {
			if len(in.Sessions) > 0 {
				return errInvalidTradingCalendar
			}
			source, sourceErr := calendarModel.FindOneForUpdate(ctx, in.SourceCalendarId)
			if sourceErr != nil {
				return sourceErr
			}
			if source.TenantId != in.TenantId || source.CalendarCode != code ||
				(source.Status != int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_APPROVED) &&
					source.Status != int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_SUPERSEDED)) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			timezone = source.Timezone
			sessions, sourceErr = sessionModel.FindByCalendar(ctx, source.TenantId, source.Id)
			if sourceErr != nil {
				return sourceErr
			}
			if len(exceptions) == 0 {
				sourceExceptions, sourceErr := exceptionModel.FindByCalendar(ctx, source.TenantId, source.Id)
				if sourceErr != nil {
					return sourceErr
				}
				for _, exception := range sourceExceptions {
					if exception.StartTime >= effectiveFrom {
						exceptions = append(exceptions, exception)
					}
				}
			}
		}
		if validationErr := logichelpers.ValidateTradingCalendarDefinition(
			timezone, effectiveFrom, 0, sessions, exceptions,
		); validationErr != nil {
			return errInvalidTradingCalendar
		}
		created = &models.TOptionTradingCalendar{
			TenantId: in.TenantId, CalendarCode: code, Version: version,
			Status:   int64(option.TradingCalendarStatus_TRADING_CALENDAR_STATUS_DRAFT),
			Timezone: timezone, EffectiveFrom: effectiveFrom, ChangeReason: changeReason,
			EvidenceRef: evidenceRef, CreatedBy: operatorID, CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := calendarModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}
		createdSessions = make([]*models.TOptionTradingCalendarSession, 0, len(sessions))
		for _, source := range sessions {
			item := &models.TOptionTradingCalendarSession{
				TenantId: in.TenantId, CalendarId: created.Id, Weekday: source.Weekday,
				OpenSecond: source.OpenSecond, CloseSecond: source.CloseSecond, CreateTimes: now,
			}
			result, insertErr = sessionModel.Insert(ctx, item)
			if insertErr != nil {
				return insertErr
			}
			item.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			createdSessions = append(createdSessions, item)
		}
		createdExceptions = make([]*models.TOptionTradingCalendarException, 0, len(exceptions))
		for _, source := range exceptions {
			item := &models.TOptionTradingCalendarException{
				TenantId: in.TenantId, CalendarId: created.Id, ExceptionType: source.ExceptionType,
				StartTime: source.StartTime, EndTime: source.EndTime, Reason: source.Reason,
				AnnouncementRef: source.AnnouncementRef, CreateTimes: now,
			}
			result, insertErr = exceptionModel.Insert(ctx, item)
			if insertErr != nil {
				return insertErr
			}
			item.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			createdExceptions = append(createdExceptions, item)
		}
		return nil
	})
	if errors.Is(err, errInvalidTradingCalendar) {
		return calendarParamError(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetTradingCalendarResp{
		Base: helper.OkResp(),
		Data: logichelpers.ToTradingCalendarProto(created, createdSessions, createdExceptions),
	}, nil
}

func tradingCalendarSessionsFromInput(
	input []*option.TradingCalendarSessionInput,
) []*models.TOptionTradingCalendarSession {
	result := make([]*models.TOptionTradingCalendarSession, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		result = append(result, &models.TOptionTradingCalendarSession{
			Weekday: int64(item.Weekday), OpenSecond: int64(item.OpenSecond), CloseSecond: int64(item.CloseSecond),
		})
	}
	return result
}

func tradingCalendarExceptionsFromInput(
	input []*option.TradingCalendarExceptionInput,
) []*models.TOptionTradingCalendarException {
	result := make([]*models.TOptionTradingCalendarException, 0, len(input))
	for _, item := range input {
		if item == nil {
			continue
		}
		result = append(result, &models.TOptionTradingCalendarException{
			ExceptionType: int64(item.ExceptionType), StartTime: item.StartTime, EndTime: item.EndTime,
			Reason: strings.TrimSpace(item.Reason), AnnouncementRef: strings.TrimSpace(item.AnnouncementRef),
		})
	}
	return result
}

func calendarParamError(ctx context.Context) *option.GetTradingCalendarResp {
	return &option.GetTradingCalendarResp{
		Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, ctx)),
	}
}

func calendarPermissionDenied(ctx context.Context) *option.GetTradingCalendarResp {
	return &option.GetTradingCalendarResp{
		Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)),
	}
}

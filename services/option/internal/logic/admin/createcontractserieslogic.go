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
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreateContractSeriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContractSeriesLogic {
	return &CreateContractSeriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建不可覆盖的到期/行权价系列版本草案
func (l *CreateContractSeriesLogic) CreateContractSeries(in *option.CreateContractSeriesReq) (*option.GetContractSeriesResp, error) {
	if in == nil {
		return contractSeriesError(l.ctx, i18n.ParamError), nil
	}
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return contractSeriesError(l.ctx, i18n.PermissionDenied), nil
	}
	prepared, err := prepareContractSeries(in)
	if err != nil || operatorID <= 0 {
		return contractSeriesError(l.ctx, i18n.ParamError), nil
	}
	if existing, findErr := l.svcCtx.OptionContractSeriesModel.FindOneByTenantIdRequestKey(
		l.ctx, in.TenantId, strings.TrimSpace(in.RequestKey),
	); findErr == nil {
		if existing.PayloadHash != prepared.payloadHash {
			return contractSeriesError(l.ctx, i18n.OperationNotAllowed), nil
		}
		return l.getContractSeries(existing)
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	now := time.Now().Unix()
	for _, expiry := range prepared.expiries {
		if expiry.ExpireTime <= now {
			return contractSeriesError(l.ctx, i18n.ParamError), nil
		}
	}
	var created *models.TOptionContractSeries
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		seriesModel := models.NewTOptionContractSeriesModel(conn, l.svcCtx.Config.CacheRedis)
		expiryModel := models.NewTOptionContractSeriesExpiryModel(conn, l.svcCtx.Config.CacheRedis)
		bandModel := models.NewTOptionContractSeriesStrikeBandModel(conn, l.svcCtx.Config.CacheRedis)
		version, supersedesID := int64(1), int64(0)
		latest, findErr := seriesModel.FindLatestForUpdate(ctx, in.TenantId, prepared.seriesCode)
		if findErr == nil {
			if latest.Status == int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_PENDING_REVIEW) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			version, supersedesID = latest.Version+1, latest.Id
		} else if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		created = &models.TOptionContractSeries{
			TenantId: in.TenantId, RequestKey: prepared.requestKey,
			SeriesCode: prepared.seriesCode, Version: version, SupersedesId: supersedesID,
			Status:             int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_PENDING_REVIEW),
			TemplateContractId: 0, TemplateSnapshot: prepared.templateSnapshot,
			UnderlyingSymbol: prepared.template.UnderlyingSymbol,
			ReferencePrice:   prepared.referencePrice, ReferenceSource: prepared.referenceSource,
			ReferenceTime: in.ReferenceTime, EvidenceRef: prepared.evidenceRef,
			ChangeReason: prepared.changeReason, PayloadHash: prepared.payloadHash,
			ExpectedContractCount: prepared.expectedCount, CreatedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := seriesModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		if insertErr != nil {
			return insertErr
		}
		for _, expiry := range prepared.expiries {
			expiry.SeriesId, expiry.CreateTimes = created.Id, now
			result, insertErr = expiryModel.Insert(ctx, expiry)
			if insertErr != nil {
				return insertErr
			}
			expiry.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
		}
		for _, band := range prepared.bands {
			band.SeriesId, band.CreateTimes = created.Id, now
			result, insertErr = bandModel.Insert(ctx, band)
			if insertErr != nil {
				return insertErr
			}
			band.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
		}
		return nil
	})
	if err != nil {
		// A concurrent retry can win the unique request key. Return that exact
		// version only when the normalized payload is identical.
		existing, findErr := l.svcCtx.OptionContractSeriesModel.FindOneByTenantIdRequestKey(
			l.ctx, in.TenantId, prepared.requestKey,
		)
		if findErr == nil && existing.PayloadHash == prepared.payloadHash {
			return l.getContractSeries(existing)
		}
		return nil, err
	}
	return &option.GetContractSeriesResp{
		Base: helper.OkResp(), Data: toContractSeriesProto(created, prepared.expiries, prepared.bands),
	}, nil
}

func (l *CreateContractSeriesLogic) getContractSeries(
	item *models.TOptionContractSeries,
) (*option.GetContractSeriesResp, error) {
	expiries, err := l.svcCtx.OptionContractSeriesExpiryModel.FindBySeries(l.ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	bands, err := l.svcCtx.OptionContractSeriesStrikeBandModel.FindBySeries(l.ctx, item.TenantId, item.Id)
	if err != nil {
		return nil, err
	}
	return &option.GetContractSeriesResp{
		Base: helper.OkResp(), Data: toContractSeriesProto(item, expiries, bands),
	}, nil
}

func contractSeriesError(ctx context.Context, key int32) *option.GetContractSeriesResp {
	return &option.GetContractSeriesResp{
		Base: helper.ErrResp(key, i18n.Translate(key, ctx)),
	}
}

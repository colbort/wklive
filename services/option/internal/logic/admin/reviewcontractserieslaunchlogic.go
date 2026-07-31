package adminlogic

import (
	"context"
	"errors"
	"fmt"
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

type ReviewContractSeriesLaunchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewContractSeriesLaunchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewContractSeriesLaunchLogic {
	return &ReviewContractSeriesLaunchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 对已生成系列执行独立上市复核；批准前生命周期必须保持 PENDING
func (l *ReviewContractSeriesLaunchLogic) ReviewContractSeriesLaunch(in *option.ReviewContractSeriesLaunchReq) (*option.GetContractSeriesResp, error) {
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
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.SeriesId <= 0 || operatorID <= 0 ||
		reason == "" || len(reason) > 500 {
		return contractSeriesError(l.ctx, i18n.ParamError), nil
	}
	now := time.Now().Unix()
	var reviewed *models.TOptionContractSeries
	var expiries []*models.TOptionContractSeriesExpiry
	var bands []*models.TOptionContractSeriesStrikeBand
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		seriesModel := models.NewTOptionContractSeriesModel(conn, l.svcCtx.Config.CacheRedis)
		expiryModel := models.NewTOptionContractSeriesExpiryModel(conn, l.svcCtx.Config.CacheRedis)
		bandModel := models.NewTOptionContractSeriesStrikeBandModel(conn, l.svcCtx.Config.CacheRedis)
		detailModel := models.NewTOptionContractSeriesDetailModel(conn, l.svcCtx.Config.CacheRedis)
		contractModel := models.NewTOptionContractModel(conn, l.svcCtx.Config.CacheRedis)

		item, findErr := seriesModel.FindOneForUpdate(ctx, in.SeriesId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId {
			return i18n.StatusError(ctx, i18n.PermissionDenied)
		}
		target := int64(option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_REJECTED)
		if in.Approve {
			target = int64(option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_APPROVED)
		}
		if item.LaunchStatus == target {
			reviewed = item
			return nil
		}
		if item.Status != int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED) ||
			item.LaunchStatus != int64(
				option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_PENDING_REVIEW,
			) ||
			item.CreatedBy == operatorID {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		latest, findErr := seriesModel.FindLatestForUpdate(ctx, item.TenantId, item.SeriesCode)
		if findErr != nil {
			return findErr
		}
		if latest.Id != item.Id {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		expiries, findErr = expiryModel.FindBySeries(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		bands, findErr = bandModel.FindBySeries(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		details, findErr := detailModel.FindBySeries(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		if int64(len(details)) != item.ExpectedContractCount ||
			item.GeneratedContractCount != item.ExpectedContractCount {
			return fmt.Errorf("contract series generation lineage count mismatch")
		}
		if in.Approve {
			template, decodeErr := decodeContractSeriesTemplate(item.TemplateSnapshot)
			if decodeErr != nil {
				return decodeErr
			}
			expiryByID := make(map[int64]*models.TOptionContractSeriesExpiry, len(expiries))
			for _, expiry := range expiries {
				expiryByID[expiry.Id] = expiry
			}
			for _, detail := range details {
				expiry := expiryByID[detail.ExpiryId]
				if expiry == nil {
					return fmt.Errorf("contract series detail references missing expiry")
				}
				contract, contractErr := contractModel.FindOne(ctx, detail.ContractId)
				if contractErr != nil {
					return contractErr
				}
				expected, buildErr := cloneSeriesContract(
					template, item, expiry, detail.StrikePrice, 0,
					option.OptionType(detail.OptionType), contract.CreateTimes,
				)
				if buildErr != nil {
					return buildErr
				}
				expected.ContractCode = detail.ContractCode
				if contract.TenantId != item.TenantId ||
					contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_PENDING) ||
					contract.IsDeleted != 2 ||
					!economicContractFieldsEqual(expected, contract) ||
					!expected.MaxUserLongQty.Equal(contract.MaxUserLongQty) ||
					!expected.MaxUserShortQty.Equal(contract.MaxUserShortQty) ||
					!expected.MaxOpenInterest.Equal(contract.MaxOpenInterest) ||
					!expected.OrderPriceBandRatio.Equal(contract.OrderPriceBandRatio) ||
					!expected.CircuitBreakerRatio.Equal(contract.CircuitBreakerRatio) {
					return fmt.Errorf("generated contract no longer matches approved series")
				}
			}
		}
		item.LaunchStatus = target
		item.LaunchReviewedBy = operatorID
		item.LaunchReviewReason = reason
		item.LaunchReviewedAt = now
		item.UpdateTimes = now
		if updateErr := seriesModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		reviewed = item
		return nil
	})
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return contractSeriesError(l.ctx, i18n.BusinessDataNotFound), nil
		}
		return nil, err
	}
	if expiries == nil {
		expiries, err = l.svcCtx.OptionContractSeriesExpiryModel.FindBySeries(
			l.ctx, reviewed.TenantId, reviewed.Id,
		)
		if err != nil {
			return nil, err
		}
	}
	if bands == nil {
		bands, err = l.svcCtx.OptionContractSeriesStrikeBandModel.FindBySeries(
			l.ctx, reviewed.TenantId, reviewed.Id,
		)
		if err != nil {
			return nil, err
		}
	}
	return &option.GetContractSeriesResp{
		Base: helper.OkResp(), Data: toContractSeriesProto(reviewed, expiries, bands),
	}, nil
}

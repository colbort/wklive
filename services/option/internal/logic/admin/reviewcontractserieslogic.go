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

type ReviewContractSeriesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewContractSeriesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewContractSeriesLogic {
	return &ReviewContractSeriesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 独立管理员复核；批准时原子生成全部 PENDING 合约
func (l *ReviewContractSeriesLogic) ReviewContractSeries(in *option.ReviewContractSeriesReq) (*option.GetContractSeriesResp, error) {
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
	generatedContracts := make([]*models.TOptionContract, 0)
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
		if item.Status == int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED) && in.Approve {
			reviewed = item
			return nil
		}
		if item.Status == int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_REJECTED) && !in.Approve {
			reviewed = item
			return nil
		}
		if item.Status != int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_PENDING_REVIEW) ||
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
		item.ReviewedBy, item.ReviewReason, item.ReviewedAt = operatorID, reason, now
		item.UpdateTimes = now
		if !in.Approve {
			item.Status = int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_REJECTED)
			if updateErr := seriesModel.Update(ctx, item); updateErr != nil {
				return updateErr
			}
			reviewed = item
			return nil
		}
		expiries, findErr = expiryModel.FindBySeries(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		bands, findErr = bandModel.FindBySeries(ctx, item.TenantId, item.Id)
		if findErr != nil {
			return findErr
		}
		if len(expiries) == 0 || len(bands) == 0 {
			return fmt.Errorf("contract series has incomplete specifications")
		}
		for _, expiry := range expiries {
			if expiry.ExpireTime <= now {
				return fmt.Errorf("contract series expiry is no longer in the future")
			}
		}
		bandInputs := make([]*option.ContractSeriesStrikeBandInput, 0, len(bands))
		for _, band := range bands {
			bandInputs = append(bandInputs, &option.ContractSeriesStrikeBandInput{
				SequenceNo: band.SequenceNo, LowerStrike: band.LowerStrike.String(),
				UpperStrike: band.UpperStrike.String(), StrikeStep: band.StrikeStep.String(),
			})
		}
		_, strikes, normalizeErr := normalizeContractSeriesBands(item.TenantId, bandInputs)
		if normalizeErr != nil {
			return normalizeErr
		}
		if int64(len(expiries)*len(strikes)*2) != item.ExpectedContractCount ||
			item.ExpectedContractCount > maxContractSeriesContracts {
			return fmt.Errorf("contract series expected count mismatch")
		}
		template, decodeErr := decodeContractSeriesTemplate(item.TemplateSnapshot)
		if decodeErr != nil {
			return decodeErr
		}
		for _, expiry := range expiries {
			for strikeIndex, strike := range strikes {
				for _, optionType := range []option.OptionType{
					option.OptionType_OPTION_TYPE_CALL, option.OptionType_OPTION_TYPE_PUT,
				} {
					contract, buildErr := cloneSeriesContract(
						template, item, expiry, strike, strikeIndex, optionType, now,
					)
					if buildErr != nil {
						return buildErr
					}
					result, insertErr := contractModel.Insert(ctx, contract)
					if insertErr != nil {
						return insertErr
					}
					contract.Id, insertErr = result.LastInsertId()
					if insertErr != nil {
						return insertErr
					}
					detail := &models.TOptionContractSeriesDetail{
						TenantId: item.TenantId, SeriesId: item.Id, ExpiryId: expiry.Id,
						OptionType: int64(optionType), StrikePrice: strike,
						ContractCode: contract.ContractCode, ContractId: contract.Id, CreateTimes: now,
					}
					if _, insertErr = detailModel.Insert(ctx, detail); insertErr != nil {
						return insertErr
					}
					generatedContracts = append(generatedContracts, contract)
				}
			}
		}
		item.Status = int64(option.ContractSeriesStatus_CONTRACT_SERIES_STATUS_GENERATED)
		item.GeneratedContractCount = int64(len(generatedContracts))
		item.GeneratedAt = now
		item.LaunchStatus = int64(
			option.ContractSeriesLaunchStatus_CONTRACT_SERIES_LAUNCH_STATUS_PENDING_REVIEW,
		)
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
	for _, contract := range generatedContracts {
		for _, enqueueErr := range enqueueContractSchedules(l.svcCtx, contract) {
			l.Errorf("enqueue series contract schedule failed, contractId=%d err=%v", contract.Id, enqueueErr)
		}
	}
	if expiries == nil {
		expiries, err = l.svcCtx.OptionContractSeriesExpiryModel.FindBySeries(l.ctx, reviewed.TenantId, reviewed.Id)
		if err != nil {
			return nil, err
		}
	}
	if bands == nil {
		bands, err = l.svcCtx.OptionContractSeriesStrikeBandModel.FindBySeries(l.ctx, reviewed.TenantId, reviewed.Id)
		if err != nil {
			return nil, err
		}
	}
	return &option.GetContractSeriesResp{
		Base: helper.OkResp(), Data: toContractSeriesProto(reviewed, expiries, bands),
	}, nil
}

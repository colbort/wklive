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

type ReviewPortfolioRiskConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewPortfolioRiskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewPortfolioRiskConfigLogic {
	return &ReviewPortfolioRiskConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 由独立管理员批准或拒绝组合保证金参数版本
func (l *ReviewPortfolioRiskConfigLogic) ReviewPortfolioRiskConfig(in *option.ReviewPortfolioRiskConfigReq) (*option.GetPortfolioRiskConfigResp, error) {
	operatorID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, in.TenantId)
	if err != nil {
		return nil, err
	}
	if forbidden || !allowed {
		return &option.GetPortfolioRiskConfigResp{
			Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx)),
		}, nil
	}
	reason := strings.TrimSpace(in.Reason)
	if in.TenantId <= 0 || in.ConfigId <= 0 || operatorID <= 0 ||
		reason == "" || len(reason) > 500 {
		return portfolioConfigParamError(l.ctx), nil
	}
	now := time.Now().Unix()
	var reviewed *models.TOptionPortfolioRiskConfig
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		configModel := models.NewTOptionPortfolioRiskConfigModel(conn, l.svcCtx.Config.CacheRedis)
		item, findErr := configModel.FindOneForUpdate(ctx, in.ConfigId)
		if findErr != nil {
			return findErr
		}
		if item.TenantId != in.TenantId ||
			item.Status != int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING) ||
			item.CreatedBy == operatorID {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		latest, findErr := configModel.FindLatestForUpdate(ctx, item.TenantId, item.SettleCoin)
		if findErr != nil {
			return findErr
		}
		if validationErr := validatePortfolioRiskConfigReview(item, latest, operatorID); validationErr != nil {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}

		if in.Approve {
			if strings.TrimSpace(item.EvidenceRef) == "" {
				return errInvalidPortfolioRiskConfig
			}
			previous, previousErr := configModel.FindOpenEndedForUpdate(
				ctx, item.TenantId, item.SettleCoin,
			)
			if previousErr != nil && !errors.Is(previousErr, models.ErrNotFound) {
				return previousErr
			}
			if previous != nil {
				if item.EffectiveFrom <= previous.EffectiveFrom || item.Version <= previous.Version {
					return errInvalidPortfolioRiskConfig
				}
				previous.Status = int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED)
				previous.EffectiveUntil = item.EffectiveFrom
				previous.UpdateTimes = now
				if updateErr := configModel.Update(ctx, previous); updateErr != nil {
					return updateErr
				}
				item.SupersedesId = previous.Id
			}
			item.Status = int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED)
		} else {
			item.Status = int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_REJECTED)
		}
		item.ReviewedBy = operatorID
		item.ReviewReason = reason
		item.ReviewedAt = now
		item.UpdateTimes = now
		if updateErr := configModel.Update(ctx, item); updateErr != nil {
			return updateErr
		}
		reviewed = item
		return nil
	})
	if errors.Is(err, errInvalidPortfolioRiskConfig) {
		return portfolioConfigParamError(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetPortfolioRiskConfigResp{
		Base: helper.OkResp(), Data: helpers.ToPortfolioRiskConfigProto(reviewed),
	}, nil
}

func validatePortfolioRiskConfigReview(
	item, latest *models.TOptionPortfolioRiskConfig,
	operatorID int64,
) error {
	if item == nil || latest == nil || operatorID <= 0 {
		return errInvalidPortfolioRiskConfig
	}
	if item.Status != int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING) ||
		item.Id != latest.Id || item.CreatedBy == operatorID {
		return errInvalidPortfolioRiskConfig
	}
	return nil
}

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
	optionrisk "wklive/services/option/internal/risk"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreatePortfolioRiskConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePortfolioRiskConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePortfolioRiskConfigLogic {
	return &CreatePortfolioRiskConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建不可覆盖的组合保证金参数草案
func (l *CreatePortfolioRiskConfigLogic) CreatePortfolioRiskConfig(in *option.CreatePortfolioRiskConfigReq) (*option.GetPortfolioRiskConfigResp, error) {
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
	now := time.Now().Unix()
	settleCoin := strings.ToUpper(strings.TrimSpace(in.SettleCoin))
	changeReason := strings.TrimSpace(in.ChangeReason)
	evidenceRef := strings.TrimSpace(in.EvidenceRef)
	effectiveFrom := in.EffectiveFrom
	if effectiveFrom <= 0 {
		effectiveFrom = now + 300
	}
	if in.TenantId <= 0 || operatorID <= 0 || !portfolioSettleCoinPattern.MatchString(settleCoin) ||
		changeReason == "" || evidenceRef == "" || len(changeReason) > 500 || len(evidenceRef) > 500 ||
		!portfolioRiskConfigEffectiveTimeValid(effectiveFrom, now) {
		return portfolioConfigParamError(l.ctx), nil
	}

	var created *models.TOptionPortfolioRiskConfig
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		configModel := models.NewTOptionPortfolioRiskConfigModel(conn, l.svcCtx.Config.CacheRedis)
		latest, findErr := configModel.FindLatestForUpdate(ctx, in.TenantId, settleCoin)
		if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		if latest != nil &&
			latest.Status == int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING) {
			return i18n.StatusError(ctx, i18n.OperationNotAllowed)
		}
		version := int64(1)
		if latest != nil {
			version = latest.Version + 1
		}

		modelMethod := in.ModelMethod
		var initialShockRate, maintenanceShockRate, concentrationThreshold,
			concentrationAddonRate, liquidityAddonRate decimal.Decimal
		var scenarioShocks string
		sourceConfigID := int64(0)
		if in.SourceConfigId > 0 {
			source, sourceErr := configModel.FindOneForUpdate(ctx, in.SourceConfigId)
			if sourceErr != nil {
				return sourceErr
			}
			if source.TenantId != in.TenantId || source.SettleCoin != settleCoin ||
				(source.Status != int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_APPROVED) &&
					source.Status != int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_SUPERSEDED)) {
				return i18n.StatusError(ctx, i18n.OperationNotAllowed)
			}
			modelMethod = option.PortfolioRiskMethod(source.ModelMethod)
			initialShockRate = source.InitialShockRate
			maintenanceShockRate = source.MaintenanceShockRate
			scenarioShocks = source.ScenarioShocks
			concentrationThreshold = source.ConcentrationThreshold
			concentrationAddonRate = source.ConcentrationAddonRate
			liquidityAddonRate = source.LiquidityAddonRate
			sourceConfigID = source.Id
		} else {
			if modelMethod != option.PortfolioRiskMethod_PORTFOLIO_RISK_METHOD_EXPIRY_SCENARIO_V1 {
				return errInvalidPortfolioRiskConfig
			}
			var parseErr error
			initialShockRate, parseErr = decimal.NewFromString(strings.TrimSpace(in.InitialShockRate))
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
			maintenanceShockRate, parseErr = decimal.NewFromString(strings.TrimSpace(in.MaintenanceShockRate))
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
			concentrationThreshold, parseErr = decimal.NewFromString(strings.TrimSpace(in.ConcentrationThreshold))
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
			concentrationAddonRate, parseErr = decimal.NewFromString(strings.TrimSpace(in.ConcentrationAddonRate))
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
			liquidityAddonRate, parseErr = decimal.NewFromString(strings.TrimSpace(in.LiquidityAddonRate))
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
			_, scenarioShocks, parseErr = optionrisk.ParseScenarioShocks(in.ScenarioShocks)
			if parseErr != nil {
				return errInvalidPortfolioRiskConfig
			}
		}
		config := optionrisk.PortfolioConfig{
			InitialShockRate: initialShockRate, MaintenanceShockRate: maintenanceShockRate,
			ConcentrationThreshold: concentrationThreshold,
			ConcentrationAddonRate: concentrationAddonRate, LiquidityAddonRate: liquidityAddonRate,
		}
		var configErr error
		config.ScenarioShocks, _, configErr = optionrisk.ParseScenarioShocks(scenarioShocks)
		if configErr != nil || optionrisk.ValidatePortfolioConfig(config) != nil {
			return errInvalidPortfolioRiskConfig
		}

		created = &models.TOptionPortfolioRiskConfig{
			TenantId: in.TenantId, SettleCoin: settleCoin, Version: version,
			Status:      int64(option.PortfolioRiskConfigStatus_PORTFOLIO_RISK_CONFIG_STATUS_PENDING),
			ModelMethod: int64(modelMethod), InitialShockRate: initialShockRate,
			MaintenanceShockRate: maintenanceShockRate, ScenarioShocks: scenarioShocks,
			ConcentrationThreshold: concentrationThreshold, ConcentrationAddonRate: concentrationAddonRate,
			LiquidityAddonRate: liquidityAddonRate, EffectiveFrom: effectiveFrom,
			SourceConfigId: sourceConfigID,
			ChangeReason:   changeReason, EvidenceRef: evidenceRef, CreatedBy: operatorID,
			CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := configModel.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		return insertErr
	})
	if errors.Is(err, errInvalidPortfolioRiskConfig) {
		return portfolioConfigParamError(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	return &option.GetPortfolioRiskConfigResp{
		Base: helper.OkResp(), Data: helpers.ToPortfolioRiskConfigProto(created),
	}, nil
}

package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AdminUpdateMarketLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpdateMarketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateMarketLogic {
	return &AdminUpdateMarketLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权行情数据
func (l *AdminUpdateMarketLogic) AdminUpdateMarket(in *option.UpdateMarketReq) (*option.AdminCommonResp, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	_, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, contract.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}

	now := time.Now().Unix()
	market, err := l.svcCtx.OptionMarketModel.FindOneByTenantIdContractId(l.ctx, contract.TenantId, in.ContractId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if market == nil {
		market = &models.TOptionMarket{
			TenantId:    contract.TenantId,
			ContractId:  in.ContractId,
			CreateTimes: now,
		}
	}
	if in.UnderlyingPrice != "" {
		value, err := conv.ParseFloatField(in.UnderlyingPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.UnderlyingPriceFormatError, i18n.Translate(i18n.UnderlyingPriceFormatError, l.ctx))}, nil
		}
		market.UnderlyingPrice = value
	}
	if in.MarkPrice != "" {
		value, err := conv.ParseFloatField(in.MarkPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.MarkPriceFormatError, i18n.Translate(i18n.MarkPriceFormatError, l.ctx))}, nil
		}
		market.MarkPrice = value
	}
	if in.LastPrice != "" {
		value, err := conv.ParseFloatField(in.LastPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.LastPriceFormatError, i18n.Translate(i18n.LastPriceFormatError, l.ctx))}, nil
		}
		market.LastPrice = value
	}
	if in.BidPrice != "" {
		value, err := conv.ParseFloatField(in.BidPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.BidPriceFormatError, i18n.Translate(i18n.BidPriceFormatError, l.ctx))}, nil
		}
		market.BidPrice = value
	}
	if in.AskPrice != "" {
		value, err := conv.ParseFloatField(in.AskPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.AskPriceFormatError, i18n.Translate(i18n.AskPriceFormatError, l.ctx))}, nil
		}
		market.AskPrice = value
	}
	if in.TheoreticalPrice != "" {
		value, err := conv.ParseFloatField(in.TheoreticalPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.TheoreticalPriceFormatError, i18n.Translate(i18n.TheoreticalPriceFormatError, l.ctx))}, nil
		}
		market.TheoreticalPrice = value
	}
	if in.IntrinsicValue != "" {
		value, err := conv.ParseFloatField(in.IntrinsicValue)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.IntrinsicValueFormatError, i18n.Translate(i18n.IntrinsicValueFormatError, l.ctx))}, nil
		}
		market.IntrinsicValue = value
	}
	if in.TimeValue != "" {
		value, err := conv.ParseFloatField(in.TimeValue)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.TimeValueFormatError, i18n.Translate(i18n.TimeValueFormatError, l.ctx))}, nil
		}
		market.TimeValue = value
	}
	if in.Iv != "" {
		value, err := conv.ParseFloatField(in.Iv)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.IVFormatError, i18n.Translate(i18n.IVFormatError, l.ctx))}, nil
		}
		market.Iv = value
	}
	if in.Delta != "" {
		value, err := conv.ParseFloatField(in.Delta)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.DeltaFormatError, i18n.Translate(i18n.DeltaFormatError, l.ctx))}, nil
		}
		market.Delta = value
	}
	if in.Gamma != "" {
		value, err := conv.ParseFloatField(in.Gamma)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.GammaFormatError, i18n.Translate(i18n.GammaFormatError, l.ctx))}, nil
		}
		market.Gamma = value
	}
	if in.Theta != "" {
		value, err := conv.ParseFloatField(in.Theta)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.ThetaFormatError, i18n.Translate(i18n.ThetaFormatError, l.ctx))}, nil
		}
		market.Theta = value
	}
	if in.Vega != "" {
		value, err := conv.ParseFloatField(in.Vega)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.VegaFormatError, i18n.Translate(i18n.VegaFormatError, l.ctx))}, nil
		}
		market.Vega = value
	}
	if in.Rho != "" {
		value, err := conv.ParseFloatField(in.Rho)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.RhoFormatError, i18n.Translate(i18n.RhoFormatError, l.ctx))}, nil
		}
		market.Rho = value
	}
	if in.RiskFreeRate != "" {
		value, err := conv.ParseFloatField(in.RiskFreeRate)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(i18n.RiskFreeRateFormatError, i18n.Translate(i18n.RiskFreeRateFormatError, l.ctx))}, nil
		}
		market.RiskFreeRate = value
	}
	if in.PricingModel != "" {
		market.PricingModel = in.PricingModel
	}
	if in.SnapshotTime != 0 {
		market.SnapshotTime = in.SnapshotTime
	} else {
		market.SnapshotTime = now
	}
	market.UpdateTimes = now

	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		marketModel := models.NewTOptionMarketModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionMarketModel)
		snapshotModel := models.NewTOptionMarketSnapshotModel(conn, l.svcCtx.Config.CacheRedis).(models.OptionMarketSnapshotModel)

		if market.Id == 0 {
			result, err := marketModel.Insert(ctx, market)
			if err != nil {
				return err
			}
			market.Id, _ = result.LastInsertId()
		} else if err := marketModel.Update(ctx, market); err != nil {
			return err
		}

		return insertMarketSnapshot(ctx, snapshotModel, market, now)
	})
	if err != nil {
		return nil, err
	}

	return &option.AdminCommonResp{Base: helper.OkResp()}, nil
}

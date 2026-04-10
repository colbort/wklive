package logic

import (
	"context"
	"errors"
	"time"

	commonconv "wklive/common/conv"
	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
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
			return &option.AdminCommonResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && contract.TenantId != in.TenantId {
		return &option.AdminCommonResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
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
		value, err := commonconv.ParseFloatField(in.UnderlyingPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "underlying_price格式错误")}, nil
		}
		market.UnderlyingPrice = value
	}
	if in.MarkPrice != "" {
		value, err := commonconv.ParseFloatField(in.MarkPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "mark_price格式错误")}, nil
		}
		market.MarkPrice = value
	}
	if in.LastPrice != "" {
		value, err := commonconv.ParseFloatField(in.LastPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "last_price格式错误")}, nil
		}
		market.LastPrice = value
	}
	if in.BidPrice != "" {
		value, err := commonconv.ParseFloatField(in.BidPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "bid_price格式错误")}, nil
		}
		market.BidPrice = value
	}
	if in.AskPrice != "" {
		value, err := commonconv.ParseFloatField(in.AskPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "ask_price格式错误")}, nil
		}
		market.AskPrice = value
	}
	if in.TheoreticalPrice != "" {
		value, err := commonconv.ParseFloatField(in.TheoreticalPrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "theoretical_price格式错误")}, nil
		}
		market.TheoreticalPrice = value
	}
	if in.IntrinsicValue != "" {
		value, err := commonconv.ParseFloatField(in.IntrinsicValue)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "intrinsic_value格式错误")}, nil
		}
		market.IntrinsicValue = value
	}
	if in.TimeValue != "" {
		value, err := commonconv.ParseFloatField(in.TimeValue)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "time_value格式错误")}, nil
		}
		market.TimeValue = value
	}
	if in.Iv != "" {
		value, err := commonconv.ParseFloatField(in.Iv)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "iv格式错误")}, nil
		}
		market.Iv = value
	}
	if in.Delta != "" {
		value, err := commonconv.ParseFloatField(in.Delta)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "delta格式错误")}, nil
		}
		market.Delta = value
	}
	if in.Gamma != "" {
		value, err := commonconv.ParseFloatField(in.Gamma)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "gamma格式错误")}, nil
		}
		market.Gamma = value
	}
	if in.Theta != "" {
		value, err := commonconv.ParseFloatField(in.Theta)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "theta格式错误")}, nil
		}
		market.Theta = value
	}
	if in.Vega != "" {
		value, err := commonconv.ParseFloatField(in.Vega)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "vega格式错误")}, nil
		}
		market.Vega = value
	}
	if in.Rho != "" {
		value, err := commonconv.ParseFloatField(in.Rho)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "rho格式错误")}, nil
		}
		market.Rho = value
	}
	if in.RiskFreeRate != "" {
		value, err := commonconv.ParseFloatField(in.RiskFreeRate)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "risk_free_rate格式错误")}, nil
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

	if market.Id == 0 {
		result, err := l.svcCtx.OptionMarketModel.Insert(l.ctx, market)
		if err != nil {
			return nil, err
		}
		market.Id, _ = result.LastInsertId()
	} else if err := l.svcCtx.OptionMarketModel.Update(l.ctx, market); err != nil {
		return nil, err
	}

	_, err = l.svcCtx.OptionMarketSnapshotModel.Insert(l.ctx, &models.TOptionMarketSnapshot{
		TenantId:         contract.TenantId,
		ContractId:       in.ContractId,
		UnderlyingPrice:  market.UnderlyingPrice,
		MarkPrice:        market.MarkPrice,
		LastPrice:        market.LastPrice,
		BidPrice:         market.BidPrice,
		AskPrice:         market.AskPrice,
		TheoreticalPrice: market.TheoreticalPrice,
		Iv:               market.Iv,
		Delta:            market.Delta,
		Gamma:            market.Gamma,
		Theta:            market.Theta,
		Vega:             market.Vega,
		Rho:              market.Rho,
		SnapshotTime:     market.SnapshotTime,
		CreateTimes:      now,
	})
	if err != nil {
		return nil, err
	}

	return &option.AdminCommonResp{Base: helper.OkResp()}, nil
}

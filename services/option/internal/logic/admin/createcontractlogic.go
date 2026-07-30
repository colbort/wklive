package adminlogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateContractLogic {
	return &CreateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建期权合约
func (l *CreateContractLogic) CreateContract(in *option.CreateContractReq) (*option.CreateContractResp, error) {
	if _, err := l.svcCtx.OptionContractModel.FindOneByTenantIdContractCode(l.ctx, in.TenantId, in.ContractCode); err == nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ContractCodeAlreadyExists, i18n.Translate(i18n.ContractCodeAlreadyExists, l.ctx))}, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	strikePrice, err := conv.ParseDecimalField(in.StrikePrice)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.StrikePriceFormatError, i18n.Translate(i18n.StrikePriceFormatError, l.ctx))}, nil
	}
	contractUnit, err := conv.ParseDecimalField(in.ContractUnit)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ContractUnitFormatError, i18n.Translate(i18n.ContractUnitFormatError, l.ctx))}, nil
	}
	minOrderQty, err := conv.ParseDecimalField(in.MinOrderQty)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.MinOrderQuantityFormatError, i18n.Translate(i18n.MinOrderQuantityFormatError, l.ctx))}, nil
	}
	maxOrderQty, err := conv.ParseDecimalField(in.MaxOrderQty)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.MaxOrderQuantityFormatError, i18n.Translate(i18n.MaxOrderQuantityFormatError, l.ctx))}, nil
	}
	priceTick, err := conv.ParseDecimalField(in.PriceTick)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.PriceTickFormatError, i18n.Translate(i18n.PriceTickFormatError, l.ctx))}, nil
	}
	qtyStep, err := conv.ParseDecimalField(in.QtyStep)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.QuantityStepFormatError, i18n.Translate(i18n.QuantityStepFormatError, l.ctx))}, nil
	}
	multiplier, err := conv.ParseDecimalField(in.Multiplier)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.MultiplierFormatError, i18n.Translate(i18n.MultiplierFormatError, l.ctx))}, nil
	}
	makerFeeRate, err := parseOptionalOptionRate(in.MakerFeeRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	takerFeeRate, err := parseOptionalOptionRate(in.TakerFeeRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	exerciseFeeRate, err := parseOptionalOptionRate(in.ExerciseFeeRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	initialMarginRate, err := parseOptionalOptionRate(in.InitialMarginRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	maintenanceMarginRate, err := parseOptionalOptionRate(in.MaintenanceMarginRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	minMarginRate, err := parseOptionalOptionRate(in.MinMarginRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	liquidationFeeRate, err := parseOptionalOptionRate(in.LiquidationFeeRate)
	if err != nil {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	sellerMarginMode := in.SellerMarginMode
	if sellerMarginMode == option.SellerMarginMode_SELLER_MARGIN_MODE_UNKNOWN {
		sellerMarginMode = option.SellerMarginMode_SELLER_MARGIN_MODE_DISABLED
	}
	deficitPolicy := in.LiquidationDeficitPolicy
	if deficitPolicy == option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_UNKNOWN {
		deficitPolicy = option.LiquidationDeficitPolicy_LIQUIDATION_DEFICIT_POLICY_MANUAL_REVIEW
	}

	now := time.Now().Unix()
	item := &models.TOptionContract{
		TenantId:          in.TenantId,
		ContractCode:      in.ContractCode,
		UnderlyingSymbol:  in.UnderlyingSymbol,
		UnderlyingCoin:    in.UnderlyingCoin,
		SettleCoin:        in.SettleCoin,
		QuoteCoin:         in.QuoteCoin,
		OptionType:        int64(in.OptionType),
		ExerciseStyle:     int64(in.ExerciseStyle),
		SettlementType:    int64(in.SettlementType),
		StrikePrice:       strikePrice,
		ContractUnit:      contractUnit,
		MinOrderQty:       minOrderQty,
		MaxOrderQty:       maxOrderQty,
		PriceTick:         priceTick,
		QtyStep:           qtyStep,
		Multiplier:        multiplier,
		ListTime:          in.ListTime,
		ExpireTime:        in.ExpireTime,
		DeliverTime:       in.DeliverTime,
		IsAutoExercise:    int64(in.IsAutoExercise),
		MakerFeeRate:      makerFeeRate,
		TakerFeeRate:      takerFeeRate,
		ExerciseFeeRate:   exerciseFeeRate,
		FeeUserId:         in.FeeUserId,
		FeeAccountId:      in.FeeAccountId,
		SellerMarginMode:  int64(sellerMarginMode),
		InitialMarginRate: initialMarginRate, MaintenanceMarginRate: maintenanceMarginRate,
		MinMarginRate: minMarginRate, LiquidationFeeRate: liquidationFeeRate,
		InsuranceUserId: in.InsuranceUserId, InsuranceAccountId: in.InsuranceAccountId,
		LiquidationDeficitPolicy: int64(deficitPolicy),
		PhysicalDeliveryPolicy:   int64(in.PhysicalDeliveryPolicy),
		Status:                   int64(in.Status),
		Sort:                     int64(in.Sort),
		Remark:                   in.Remark,
		IsDeleted:                int64(common.YesNo_YES_NO_NO),
		CreateTimes:              now,
		UpdateTimes:              now,
	}
	if !validateSupportedContract(item) {
		return &option.CreateContractResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	result, err := l.svcCtx.OptionContractModel.Insert(l.ctx, item)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	item.Id = id
	for _, enqueueErr := range enqueueContractSchedules(l.svcCtx, item) {
		l.Errorf("enqueue option contract schedule failed, contractId=%d err=%v", id, enqueueErr)
	}

	return &option.CreateContractResp{Id: id, Base: helper.OkResp()}, nil
}

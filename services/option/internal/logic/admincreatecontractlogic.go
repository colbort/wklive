package logic

import (
	"context"
	"errors"
	"github.com/zeromicro/go-zero/core/logx"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"
)

type AdminCreateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminCreateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminCreateContractLogic {
	return &AdminCreateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建期权合约
func (l *AdminCreateContractLogic) AdminCreateContract(in *option.CreateContractReq) (*option.CreateContractResp, error) {
	if _, err := l.svcCtx.OptionContractModel.FindOneByTenantIdContractCode(l.ctx, in.TenantId, in.ContractCode); err == nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ContractCodeAlreadyExists, l.ctx))}, nil
	} else if !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	strikePrice, err := conv.ParseFloatField(in.StrikePrice)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.StrikePriceFormatError, l.ctx))}, nil
	}
	contractUnit, err := conv.ParseFloatField(in.ContractUnit)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ContractUnitFormatError, l.ctx))}, nil
	}
	minOrderQty, err := conv.ParseFloatField(in.MinOrderQty)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.MinOrderQuantityFormatError, l.ctx))}, nil
	}
	maxOrderQty, err := conv.ParseFloatField(in.MaxOrderQty)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.MaxOrderQuantityFormatError, l.ctx))}, nil
	}
	priceTick, err := conv.ParseFloatField(in.PriceTick)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.PriceTickFormatError, l.ctx))}, nil
	}
	qtyStep, err := conv.ParseFloatField(in.QtyStep)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.QuantityStepFormatError, l.ctx))}, nil
	}
	multiplier, err := conv.ParseFloatField(in.Multiplier)
	if err != nil {
		return &option.CreateContractResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.MultiplierFormatError, l.ctx))}, nil
	}

	now := time.Now().Unix()
	item := &models.TOptionContract{
		TenantId:         in.TenantId,
		ContractCode:     in.ContractCode,
		UnderlyingSymbol: in.UnderlyingSymbol,
		SettleCoin:       in.SettleCoin,
		QuoteCoin:        in.QuoteCoin,
		OptionType:       int64(in.OptionType),
		ExerciseStyle:    int64(in.ExerciseStyle),
		SettlementType:   int64(in.SettlementType),
		StrikePrice:      strikePrice,
		ContractUnit:     contractUnit,
		MinOrderQty:      minOrderQty,
		MaxOrderQty:      maxOrderQty,
		PriceTick:        priceTick,
		QtyStep:          qtyStep,
		Multiplier:       multiplier,
		ListTime:         in.ListTime,
		ExpireTime:       in.ExpireTime,
		DeliverTime:      in.DeliverTime,
		IsAutoExercise:   int64(in.IsAutoExercise),
		Status:           int64(in.Status),
		Sort:             int64(in.Sort),
		Remark:           in.Remark,
		IsDeleted:        int64(option.YesNo_YES_NO_NO),
		CreateTimes:      now,
		UpdateTimes:      now,
	}

	result, err := l.svcCtx.OptionContractModel.Insert(l.ctx, item)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &option.CreateContractResp{Id: id, Base: helper.OkResp()}, nil
}

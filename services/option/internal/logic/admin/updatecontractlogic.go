package adminlogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateContractLogic {
	return &UpdateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权合约
func (l *UpdateContractLogic) UpdateContract(in *option.UpdateContractReq) (*option.CommonResp, error) {
	item, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &option.CommonResp{Base: helper.ErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &option.CommonResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	targetTenantId := item.TenantId
	if allowTenantUpdate {
		targetTenantId = in.TenantId
	}

	if targetTenantId != item.TenantId || (in.ContractCode != "" && in.ContractCode != item.ContractCode) {
		contractCode := item.ContractCode
		if in.ContractCode != "" {
			contractCode = in.ContractCode
		}
		dup, err := l.svcCtx.OptionContractModel.FindOneByTenantIdContractCode(l.ctx, targetTenantId, contractCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if dup != nil && dup.Id != item.Id {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractCodeAlreadyExists, i18n.Translate(i18n.ContractCodeAlreadyExists, l.ctx))}, nil
		}
		item.ContractCode = contractCode
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}
	if in.UnderlyingSymbol != "" {
		item.UnderlyingSymbol = in.UnderlyingSymbol
	}
	if in.SettleCoin != "" {
		item.SettleCoin = in.SettleCoin
	}
	if in.QuoteCoin != "" {
		item.QuoteCoin = in.QuoteCoin
	}
	if in.OptionType != 0 {
		item.OptionType = int64(in.OptionType)
	}
	if in.ExerciseStyle != 0 {
		item.ExerciseStyle = int64(in.ExerciseStyle)
	}
	if in.SettlementType != 0 {
		item.SettlementType = int64(in.SettlementType)
	}
	if in.StrikePrice != "" {
		value, err := conv.ParseDecimalField(in.StrikePrice)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.StrikePriceFormatError, i18n.Translate(i18n.StrikePriceFormatError, l.ctx))}, nil
		}
		item.StrikePrice = value
	}
	if in.ContractUnit != "" {
		value, err := conv.ParseDecimalField(in.ContractUnit)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.ContractUnitFormatError, i18n.Translate(i18n.ContractUnitFormatError, l.ctx))}, nil
		}
		item.ContractUnit = value
	}
	if in.MinOrderQty != "" {
		value, err := conv.ParseDecimalField(in.MinOrderQty)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MinOrderQuantityFormatError, i18n.Translate(i18n.MinOrderQuantityFormatError, l.ctx))}, nil
		}
		item.MinOrderQty = value
	}
	if in.MaxOrderQty != "" {
		value, err := conv.ParseDecimalField(in.MaxOrderQty)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MaxOrderQuantityFormatError, i18n.Translate(i18n.MaxOrderQuantityFormatError, l.ctx))}, nil
		}
		item.MaxOrderQty = value
	}
	if in.PriceTick != "" {
		value, err := conv.ParseDecimalField(in.PriceTick)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.PriceTickFormatError, i18n.Translate(i18n.PriceTickFormatError, l.ctx))}, nil
		}
		item.PriceTick = value
	}
	if in.QtyStep != "" {
		value, err := conv.ParseDecimalField(in.QtyStep)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.QuantityStepFormatError, i18n.Translate(i18n.QuantityStepFormatError, l.ctx))}, nil
		}
		item.QtyStep = value
	}
	if in.Multiplier != "" {
		value, err := conv.ParseDecimalField(in.Multiplier)
		if err != nil {
			return &option.CommonResp{Base: helper.ErrResp(i18n.MultiplierFormatError, i18n.Translate(i18n.MultiplierFormatError, l.ctx))}, nil
		}
		item.Multiplier = value
	}
	if in.ListTime != 0 {
		item.ListTime = in.ListTime
	}
	if in.ExpireTime != 0 {
		item.ExpireTime = in.ExpireTime
	}
	if in.DeliverTime != 0 {
		item.DeliverTime = in.DeliverTime
	}
	if in.IsAutoExercise != 0 {
		item.IsAutoExercise = int64(in.IsAutoExercise)
	}
	if in.Status != 0 {
		item.Status = int64(in.Status)
	}
	if in.Sort != 0 {
		item.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	if in.IsDeleted != 0 {
		item.IsDeleted = int64(in.IsDeleted)
	}
	item.UpdateTimes = time.Now().Unix()

	if err := l.svcCtx.OptionContractModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &option.CommonResp{Base: helper.OkResp()}, nil
}

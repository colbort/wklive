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

type AdminUpdateContractLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUpdateContractLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateContractLogic {
	return &AdminUpdateContractLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新期权合约
func (l *AdminUpdateContractLogic) AdminUpdateContract(in *option.UpdateContractReq) (*option.AdminCommonResp, error) {
	item, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AdminCommonResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && item.TenantId != in.TenantId {
		return &option.AdminCommonResp{Base: helper.GetErrResp(404, "合约不存在")}, nil
	}

	if in.ContractCode != "" && in.ContractCode != item.ContractCode {
		dup, err := l.svcCtx.OptionContractModel.FindOneByTenantIdContractCode(l.ctx, item.TenantId, in.ContractCode)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if dup != nil && dup.Id != item.Id {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "合约编码已存在")}, nil
		}
		item.ContractCode = in.ContractCode
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
		value, err := commonconv.ParseFloatField(in.StrikePrice)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "strike_price格式错误")}, nil
		}
		item.StrikePrice = value
	}
	if in.ContractUnit != "" {
		value, err := commonconv.ParseFloatField(in.ContractUnit)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "contract_unit格式错误")}, nil
		}
		item.ContractUnit = value
	}
	if in.MinOrderQty != "" {
		value, err := commonconv.ParseFloatField(in.MinOrderQty)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "min_order_qty格式错误")}, nil
		}
		item.MinOrderQty = value
	}
	if in.MaxOrderQty != "" {
		value, err := commonconv.ParseFloatField(in.MaxOrderQty)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "max_order_qty格式错误")}, nil
		}
		item.MaxOrderQty = value
	}
	if in.PriceTick != "" {
		value, err := commonconv.ParseFloatField(in.PriceTick)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "price_tick格式错误")}, nil
		}
		item.PriceTick = value
	}
	if in.QtyStep != "" {
		value, err := commonconv.ParseFloatField(in.QtyStep)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "qty_step格式错误")}, nil
		}
		item.QtyStep = value
	}
	if in.Multiplier != "" {
		value, err := commonconv.ParseFloatField(in.Multiplier)
		if err != nil {
			return &option.AdminCommonResp{Base: helper.GetErrResp(400, "multiplier格式错误")}, nil
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

	return &option.AdminCommonResp{Base: helper.OkResp()}, nil
}

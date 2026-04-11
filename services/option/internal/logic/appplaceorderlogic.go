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

type AppPlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAppPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AppPlaceOrderLogic {
	return &AppPlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交期权下单请求
func (l *AppPlaceOrderLogic) AppPlaceOrder(in *option.AppPlaceOrderReq) (*option.AppPlaceOrderResp, error) {
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.AppPlaceOrderResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if contract.TenantId != in.TenantId {
		return &option.AppPlaceOrderResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
		return &option.AppPlaceOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}

	price, err := conv.ParseFloatField(in.Price)
	if err != nil {
		return &option.AppPlaceOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.PriceFormatError, l.ctx))}, nil
	}
	qty, err := conv.ParseFloatField(in.Qty)
	if err != nil || qty <= 0 {
		return &option.AppPlaceOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}

	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.OptionOrderModel.FindOneByTenantIdUidClientOrderId(l.ctx, in.TenantId, in.Uid, in.ClientOrderId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			return &option.AppPlaceOrderResp{Base: helper.GetErrResp(400, i18n.Translate(i18n.ClientOrderIDAlreadyExists, l.ctx)), OrderNo: exists.OrderNo, OrderId: exists.Id}, nil
		}
	}

	now := time.Now().Unix()
	order := &models.TOptionOrder{
		TenantId:         in.TenantId,
		OrderNo:          conv.GenerateBizNo("OP"),
		Uid:              in.Uid,
		AccountId:        in.AccountId,
		ContractId:       in.ContractId,
		UnderlyingSymbol: contract.UnderlyingSymbol,
		Side:             int64(in.Side),
		PositionEffect:   int64(in.PositionEffect),
		OrderType:        int64(in.OrderType),
		Price:            price,
		Qty:              qty,
		FilledQty:        0,
		UnfilledQty:      qty,
		AvgPrice:         0,
		Turnover:         0,
		Fee:              0,
		FeeCoin:          contract.SettleCoin,
		MarginAmount:     0,
		Source:           int64(option.OrderSource_ORDER_SOURCE_APP),
		ClientOrderId:    in.ClientOrderId,
		ReduceOnly:       int64(in.ReduceOnly),
		Mmp:              int64(in.Mmp),
		Status:           int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	result, err := l.svcCtx.OptionOrderModel.Insert(l.ctx, order)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &option.AppPlaceOrderResp{Base: helper.OkResp(), OrderNo: order.OrderNo, OrderId: id}, nil
}

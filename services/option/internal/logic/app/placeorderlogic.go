package applogic

import (
	"context"
	"errors"
	"time"
	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/shopspring/decimal"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type PlaceOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交期权下单请求
func (l *PlaceOrderLogic) PlaceOrder(in *option.PlaceOrderReq) (*option.PlaceOrderResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	contract, err := l.svcCtx.OptionContractModel.FindOne(l.ctx, in.ContractId)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if contract.TenantId != tenantId {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotFound, i18n.Translate(i18n.ContractNotFound, l.ctx))}, nil
	}
	if contract.Status != int64(option.ContractStatus_CONTRACT_STATUS_TRADING) {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ContractNotTradable, i18n.Translate(i18n.ContractNotTradable, l.ctx))}, nil
	}

	price, err := conv.ParseDecimalField(in.Price)
	if err != nil {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PriceFormatError, i18n.Translate(i18n.PriceFormatError, l.ctx))}, nil
	}
	qty, err := conv.ParseDecimalField(in.Qty)
	if err != nil || !qty.IsPositive() {
		return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx))}, nil
	}

	marginAmount := decimal.Zero
	if in.PositionEffect == option.PositionEffect_POSITION_EFFECT_OPEN {
		multiplier := contract.Multiplier
		if !multiplier.IsPositive() {
			multiplier = contract.ContractUnit
		}
		if !multiplier.IsPositive() {
			multiplier = decimal.NewFromInt(1)
		}
		marginAmount = price.Mul(qty).Mul(multiplier)
	}

	if in.ClientOrderId != "" {
		exists, err := l.svcCtx.OptionOrderModel.FindOneByTenantIdUserIdClientOrderId(l.ctx, tenantId, userId, in.ClientOrderId)
		if err != nil && !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		if exists != nil {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.ClientOrderIDAlreadyExists, i18n.Translate(i18n.ClientOrderIDAlreadyExists, l.ctx)), Data: &option.PlaceOrderData{OrderNo: exists.OrderNo, OrderId: exists.Id}}, nil
		}
	}

	orderNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "OP", "")
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	order := &models.TOptionOrder{
		TenantId:         tenantId,
		OrderNo:          orderNo,
		UserId:           userId,
		AccountId:        in.AccountId,
		ContractId:       in.ContractId,
		UnderlyingSymbol: contract.UnderlyingSymbol,
		Side:             int64(in.Side),
		PositionEffect:   int64(in.PositionEffect),
		OrderType:        int64(in.OrderType),
		Price:            price,
		Qty:              qty,
		FilledQty:        decimal.Zero,
		UnfilledQty:      qty,
		AvgPrice:         decimal.Zero,
		Turnover:         decimal.Zero,
		Fee:              decimal.Zero,
		FeeCoin:          contract.SettleCoin,
		MarginAmount:     marginAmount,
		Source:           int64(option.OrderSource_ORDER_SOURCE_APP),
		ClientOrderId:    in.ClientOrderId,
		ReduceOnly:       int64(in.ReduceOnly),
		Mmp:              int64(in.Mmp),
		Status:           int64(option.OrderStatus_ORDER_STATUS_PENDING),
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	var id int64
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
		positionModel := models.NewTOptionPositionModel(conn, l.svcCtx.Config.CacheRedis)
		result, err := orderModel.Insert(ctx, order)
		if err != nil {
			return err
		}
		id, err = result.LastInsertId()
		if err != nil {
			return err
		}
		order.Id = id
		if err := freezeClosePosition(ctx, positionModel, order, now); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.PositionNotFound, i18n.Translate(i18n.PositionNotFound, l.ctx)), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
		}
		if i18n.IsStatusError(err, i18n.QuantityFormatError) {
			return &option.PlaceOrderResp{Base: helper.ErrResp(i18n.QuantityFormatError, i18n.Translate(i18n.QuantityFormatError, l.ctx)), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
		}
		return nil, err
	}

	if marginAmount.IsPositive() {
		resp, err := l.svcCtx.AssetClient.FreezeAsset(l.ctx, &asset.FreezeAssetReq{
			TenantId:   tenantId,
			UserId:     userId,
			WalletType: common.WalletType_WALLET_TYPE_OPTION,
			Coin:       contract.SettleCoin,
			Amount:     conv.FloatString(marginAmount),
			BizType:    asset.BizType_BIZ_TYPE_OPTION,
			SceneType:  asset.SceneType_SCENE_TYPE_PLACE_ORDER,
			BizId:      id,
			BizNo:      order.OrderNo,
			Remark:     "option place order freeze",
		})
		if err != nil {
			order.Status = int64(option.OrderStatus_ORDER_STATUS_REJECTED)
			order.CancelReason = err.Error()
			order.UpdateTimes = time.Now().Unix()
			if updateErr := l.svcCtx.OptionOrderModel.Update(l.ctx, order); updateErr != nil {
				l.Errorf("update rejected option order failed, orderNo=%s err=%v", order.OrderNo, updateErr)
			}
			return nil, err
		}
		if resp == nil || resp.Base == nil || resp.Base.Code != 200 {
			order.Status = int64(option.OrderStatus_ORDER_STATUS_REJECTED)
			if resp != nil && resp.Base != nil {
				order.CancelReason = resp.Base.Msg
			}
			order.UpdateTimes = time.Now().Unix()
			if updateErr := l.svcCtx.OptionOrderModel.Update(l.ctx, order); updateErr != nil {
				l.Errorf("update rejected option order failed, orderNo=%s err=%v", order.OrderNo, updateErr)
			}
			if resp != nil && resp.Base != nil {
				return &option.PlaceOrderResp{Base: resp.Base, Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
			}
			return nil, err
		}
	}

	if err := l.matchOrder(contract, order); err != nil {
		return nil, err
	}

	return &option.PlaceOrderResp{Base: helper.OkResp(), Data: &option.PlaceOrderData{OrderNo: order.OrderNo, OrderId: id}}, nil
}

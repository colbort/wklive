package applogic

import (
	"context"
	"errors"
	"time"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/go-sql-driver/mysql"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const comboOrderCreateMaxAttempts = 5

type PlaceComboOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPlaceComboOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceComboOrderLogic {
	return &PlaceComboOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建2至4腿独立策略簿组合订单
func (l *PlaceComboOrderLogic) PlaceComboOrder(in *option.PlaceComboOrderReq) (*option.PlaceComboOrderResp, error) {
	userID, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	payloadHash, err := comboRequestPayloadHash(in)
	if err != nil {
		return &option.PlaceComboOrderResp{
			Base: helper.ErrResp(i18n.ParamError, err.Error()),
		}, nil
	}
	existing, findErr := l.svcCtx.OptionComboOrderModel.
		FindOneByTenantIdUserIdClientComboIdNoCache(l.ctx, tenantID, userID, in.ClientComboId)
	if findErr != nil && !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	if existing != nil {
		return l.existingComboResponse(existing, payloadHash)
	}
	validated, err := validateComboOrder(l.ctx, l.svcCtx, tenantID, in)
	if err != nil {
		return &option.PlaceComboOrderResp{
			Base: helper.ErrResp(i18n.ParamError, err.Error()),
		}, nil
	}

	comboNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "combo_order_id", "OC", "")
	if err != nil {
		return nil, err
	}
	childOrderNos := make([]string, len(validated.legs))
	for index := range childOrderNos {
		childOrderNos[index], err = generate.GenerateNo(
			l.svcCtx.Redis, l.ctx, "order_id", "OP", "",
		)
		if err != nil {
			return nil, err
		}
	}
	now := time.Now().Unix()
	firstContract := validated.legs[0].contract
	parent := &models.TOptionComboOrder{
		TenantId: tenantID, ComboNo: comboNo, UserId: userID,
		AccountId: in.AccountId, ClientComboId: in.ClientComboId,
		StrategyKey:        validated.strategyKey,
		InverseStrategyKey: validated.inverseStrategyKey,
		UnderlyingSymbol:   firstContract.UnderlyingSymbol,
		ExpireTime:         firstContract.ExpireTime, SettleCoin: firstContract.SettleCoin,
		QuoteCoin: firstContract.QuoteCoin, OrderType: int64(in.OrderType),
		NetPrice: validated.netPrice, Qty: validated.qty,
		FilledQty: decimal.Zero, UnfilledQty: validated.qty,
		Status:      int64(option.ComboOrderStatus_COMBO_ORDER_STATUS_FUNDING),
		PayloadHash: validated.payloadHash, CreateTimes: now, UpdateTimes: now,
	}
	var controlRejection *orderControlRejection
	var controlRejectedOrder *models.TOptionOrder
	errControlRejected := errors.New("combo order rejected by trading controls")
	for attempt := 0; attempt < comboOrderCreateMaxAttempts; attempt++ {
		controlRejection = nil
		controlRejectedOrder = nil
		parent.Id = 0
		err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			conn := sqlx.NewSqlConnFromSession(session)
			comboModel := models.NewTOptionComboOrderModel(conn, l.svcCtx.Config.CacheRedis)
			comboLegModel := models.NewTOptionComboOrderLegModel(conn, l.svcCtx.Config.CacheRedis)
			orderModel := models.NewTOptionOrderModel(conn, l.svcCtx.Config.CacheRedis)
			instructionModel := models.NewTOptionAssetInstructionModel(conn, l.svcCtx.Config.CacheRedis)
			userControlModel := models.NewTOptionUserTradingControlModel(conn, l.svcCtx.Config.CacheRedis)

			result, insertErr := comboModel.Insert(ctx, parent)
			if insertErr != nil {
				return insertErr
			}
			parent.Id, insertErr = result.LastInsertId()
			if insertErr != nil {
				return insertErr
			}
			// Keep combo admission lock order aligned with matching: parent/user first,
			// then contracts in normalized contract-id order. evaluateOrderTradingControls
			// will reuse this already-held user row while processing every leg.
			if _, insertErr = userControlModel.EnsureForUpdate(ctx, tenantID, userID, now); insertErr != nil {
				return insertErr
			}
			for index, leg := range validated.legs {
				child := &models.TOptionOrder{
					TenantId: tenantID, OrderNo: childOrderNos[index],
					UserId: userID, AccountId: in.AccountId,
					ContractId:       leg.contract.Id,
					UnderlyingSymbol: leg.contract.UnderlyingSymbol,
					Side:             int64(leg.side),
					PositionEffect:   int64(option.PositionEffect_POSITION_EFFECT_OPEN),
					OrderType:        comboChildOrderType(in.OrderType), Price: leg.price,
					Qty: leg.qty, FilledQty: decimal.Zero, UnfilledQty: leg.qty,
					AvgPrice: decimal.Zero, Turnover: decimal.Zero, Fee: decimal.Zero,
					FeeCoin:      leg.contract.SettleCoin,
					MarginAmount: leg.marginAmount, MarginCoin: leg.marginCoin,
					Source:       int64(option.OrderSource_ORDER_SOURCE_APP),
					ReduceOnly:   int64(common.YesNo_YES_NO_NO),
					Mmp:          int64(common.YesNo_YES_NO_NO),
					ComboOrderId: parent.Id, ComboLegNo: leg.legNo,
					Status:      int64(option.OrderStatus_ORDER_STATUS_FUNDING),
					CreateTimes: now, UpdateTimes: now,
				}
				lockedContract, rejection, controlErr := evaluateOrderTradingControls(
					ctx, l.svcCtx, conn, child, now,
				)
				if controlErr != nil {
					return controlErr
				}
				if rejection != nil {
					controlRejection = rejection
					controlRejectedOrder = child
					return errControlRejected
				}
				if lockedContract.Id != leg.contract.Id {
					return errors.New("combo leg contract lock scope mismatch")
				}
				childResult, childErr := orderModel.Insert(ctx, child)
				if childErr != nil {
					return childErr
				}
				child.Id, childErr = childResult.LastInsertId()
				if childErr != nil {
					return childErr
				}
				if _, childErr = comboLegModel.Insert(ctx, &models.TOptionComboOrderLeg{
					TenantId: tenantID, ComboOrderId: parent.Id, LegNo: leg.legNo,
					ContractId: leg.contract.Id, Side: int64(leg.side),
					PositionEffect: int64(option.PositionEffect_POSITION_EFFECT_OPEN),
					Ratio:          leg.ratio, Price: leg.price, Qty: leg.qty,
					FilledQty: decimal.Zero, UnfilledQty: leg.qty,
					ChildOrderId: child.Id, CreateTimes: now, UpdateTimes: now,
				}); childErr != nil {
					return childErr
				}
				if _, childErr = instructionModel.Insert(ctx, &models.TOptionAssetInstruction{
					TenantId: tenantID, InstructionNo: child.OrderNo + "-FREEZE",
					BizNo: child.OrderNo, OrderId: child.Id, UserId: userID,
					AccountId:   in.AccountId,
					Action:      int64(option.AssetInstructionAction_ASSET_INSTRUCTION_ACTION_FREEZE),
					TargetBizNo: child.OrderNo, Coin: leg.marginCoin, Amount: leg.marginAmount,
					StepNo: 1,
					Status: int64(option.AssetInstructionStatus_ASSET_INSTRUCTION_STATUS_PENDING),
					ReconciliationStatus: int64(
						option.AssetReconciliationStatus_ASSET_RECONCILIATION_STATUS_PENDING,
					),
					CreateTimes: now, UpdateTimes: now,
				}); childErr != nil {
					return childErr
				}
			}
			return nil
		})
		if err == nil || errors.Is(err, errControlRejected) {
			break
		}
		// A unique-key winner or a transaction that committed while this request
		// was a deadlock victim is authoritative. Bypass the generated index cache:
		// its negative entry may predate the concurrent commit.
		existing, findErr = l.svcCtx.OptionComboOrderModel.
			FindOneByTenantIdUserIdClientComboIdNoCache(
				l.ctx, tenantID, userID, in.ClientComboId,
			)
		if findErr == nil {
			return l.existingComboResponse(existing, validated.payloadHash)
		}
		if !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		if !isRetryableComboOrderCreateError(err) || attempt == comboOrderCreateMaxAttempts-1 {
			break
		}
		timer := time.NewTimer(time.Duration(10*(1<<attempt)) * time.Millisecond)
		select {
		case <-l.ctx.Done():
			timer.Stop()
			return nil, l.ctx.Err()
		case <-timer.C:
		}
	}
	if errors.Is(err, errControlRejected) && controlRejection != nil {
		// The leg-level audit written by evaluateOrderTradingControls belonged to
		// the rolled-back combo transaction. Re-append it outside that transaction
		// so rejected combo requests remain visible to audit and ratio monitoring.
		if controlRejectedOrder == nil {
			return nil, errors.New("combo control rejection is missing the rejected leg")
		}
		auditErr := l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
			return recordOrderTradingControlAudit(
				ctx, sqlx.NewSqlConnFromSession(session), controlRejectedOrder, controlRejection, now,
			)
		})
		if auditErr != nil {
			return nil, auditErr
		}
		return &option.PlaceComboOrderResp{
			Base: helper.ErrResp(i18n.OperationNotAllowed, controlRejection.reason),
		}, nil
	}
	if err != nil {
		existing, findErr = l.svcCtx.OptionComboOrderModel.
			FindOneByTenantIdUserIdClientComboIdNoCache(
				l.ctx, tenantID, userID, in.ClientComboId,
			)
		if findErr == nil {
			return l.existingComboResponse(existing, validated.payloadHash)
		}
		if !errors.Is(findErr, models.ErrNotFound) {
			return nil, findErr
		}
		return nil, err
	}
	detail, err := buildComboOrderDetail(l.ctx, l.svcCtx, parent)
	if err != nil {
		return nil, err
	}
	return &option.PlaceComboOrderResp{Base: helper.OkResp(), Data: detail}, nil
}

func isRetryableComboOrderCreateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && (mysqlErr.Number == 1213 || mysqlErr.Number == 1205)
}

func (l *PlaceComboOrderLogic) existingComboResponse(
	item *models.TOptionComboOrder, payloadHash string,
) (*option.PlaceComboOrderResp, error) {
	if item.PayloadHash != payloadHash {
		return &option.PlaceComboOrderResp{
			Base: helper.ErrResp(
				i18n.ClientOrderIDAlreadyExists,
				"client_combo_id already exists with a different payload",
			),
		}, nil
	}
	detail, err := buildComboOrderDetail(l.ctx, l.svcCtx, item)
	if err != nil {
		return nil, err
	}
	return &option.PlaceComboOrderResp{Base: helper.OkResp(), Data: detail}, nil
}

func comboChildOrderType(orderType option.ComboOrderType) int64 {
	if orderType == option.ComboOrderType_COMBO_ORDER_TYPE_FOK {
		return int64(option.OrderType_ORDER_TYPE_FOK)
	}
	return int64(option.OrderType_ORDER_TYPE_LIMIT)
}

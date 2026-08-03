package adminlogic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRedeemLogic {
	return &ManualRedeemLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ManualRedeem supports full-order redemption only. It uses the same durable
// operation executor as the user endpoint, so an admin retry cannot duplicate
// principal, fee or reward movements.
func (l *ManualRedeemLogic) ManualRedeem(in *staking.ManualRedeemReq) (*staking.ManualRedeemResp, error) {
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if in.TenantId > 0 && order.TenantId != in.TenantId {
		return manualRedeemNotFound(l.ctx), nil
	}
	if base, scopeErr := helpers.AdminTenantWriteScopeResp(l.ctx, order.TenantId, i18n.OrderNotFound); scopeErr != nil {
		return nil, scopeErr
	} else if base != nil {
		return &staking.ManualRedeemResp{Page: base}, nil
	}
	operatorId, err := helpers.AdminOperatorUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	if in.RedeemType == staking.RedeemType_REDEEM_TYPE_EARLY && order.AllowEarlyRedeem != int64(common.YesNo_YES_NO_YES) {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.EarlyRedeemNotAllowed, i18n.Translate(i18n.EarlyRedeemNotAllowed, l.ctx))}, nil
	}
	if in.RedeemType != staking.RedeemType_REDEEM_TYPE_EARLY && in.RedeemType != staking.RedeemType_REDEEM_TYPE_MATURITY && in.RedeemType != staking.RedeemType_REDEEM_TYPE_MANUAL {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	redeemAmount, err := conv.ParseDecimalField(in.RedeemAmount)
	if err != nil || !redeemAmount.Equal(order.StakeAmount) {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.RedeemAmountInvalid, i18n.Translate(i18n.RedeemAmountInvalid, l.ctx))}, nil
	}
	rewardAmount, err := conv.ParseDecimalField(in.RewardAmount)
	if err != nil || rewardAmount.IsNegative() {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.RewardAmountInvalid, i18n.Translate(i18n.RewardAmountInvalid, l.ctx))}, nil
	}
	feeRate, err := conv.ParseDecimalField(in.FeeRate)
	if err != nil || feeRate.IsNegative() {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	feeAmount, err := conv.ParseDecimalField(in.FeeAmount)
	if err != nil || feeAmount.IsNegative() || feeAmount.GreaterThan(order.StakeAmount) {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	expectedFee := order.StakeAmount.Mul(feeRate).Div(decimal.NewFromInt(100)).RoundDown(8)
	if !feeAmount.Equal(expectedFee) {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	spec := helpers.RedeemOperationSpec{
		RequestNo: in.RequestNo, OperationType: helpers.StakeOperationTypeManualRedeem,
		RedeemType: in.RedeemType, RewardAmount: rewardAmount, FeeRate: feeRate,
		FeeAmount: feeAmount, OperatorId: operatorId, Remark: in.Remark,
	}
	if existing, findErr := l.svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(l.ctx, order.TenantId, order.UserId, helpers.StakeOperationTypeManualRedeem, in.RequestNo); findErr == nil {
		spec.OperationNo = existing.OperationNo
		operation, prepareErr := helpers.PrepareRedeemOperation(l.ctx, l.svcCtx, order, spec)
		if prepareErr != nil {
			return nil, prepareErr
		}
		if err := helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation); err != nil && !errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return nil, err
		}
		if operation.Status != helpers.StakeOperationStatusSucceeded {
			return manualRedeemProcessing(operation.OperationNo, l.ctx), nil
		}
		return &staking.ManualRedeemResp{Page: helper.OkResp(), Success: 1, RedeemNo: operation.OperationNo}, nil
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	if order.Status != int64(staking.OrderStatus_ORDER_STATUS_STAKING) && order.Status != int64(staking.OrderStatus_ORDER_STATUS_EXPIRED) {
		return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.StakingOrderCannotRedeem, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}

	operationNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "SKR", "")
	if err != nil {
		return nil, err
	}
	spec.OperationNo = operationNo
	operation, err := helpers.PrepareRedeemOperation(l.ctx, l.svcCtx, order, spec)
	if errors.Is(err, helpers.ErrStakeOperationProcessing) {
		return manualRedeemProcessing(operationNo, l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	if err := helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation); err != nil {
		if errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return manualRedeemProcessing(operation.OperationNo, l.ctx), nil
		}
		return nil, err
	}
	return &staking.ManualRedeemResp{Page: helper.OkResp(), Success: 1, RedeemNo: operation.OperationNo}, nil
}

func manualRedeemNotFound(ctx context.Context) *staking.ManualRedeemResp {
	return &staking.ManualRedeemResp{Page: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, ctx))}
}

func manualRedeemProcessing(operationNo string, ctx context.Context) *staking.ManualRedeemResp {
	return &staking.ManualRedeemResp{
		Page:     helper.ErrResp(i18n.AssetRequestProcessing, i18n.Translate(i18n.AssetRequestProcessing, ctx)),
		RedeemNo: operationNo,
	}
}

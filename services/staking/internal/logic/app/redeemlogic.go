package applogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

type RedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RedeemLogic {
	return &RedeemLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// Redeem creates a durable operation before changing any asset balance. A
// repeated request_no returns or resumes the same redemption.
func (l *RedeemLogic) Redeem(in *staking.RedeemReq) (*staking.RedeemResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if order.TenantId != tenantId {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, l.ctx))}, nil
	}
	if order.UserId != userId {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.NoPermissionAccessOrder, i18n.Translate(i18n.NoPermissionAccessOrder, l.ctx))}, nil
	}
	requestNo := strings.TrimSpace(in.RequestNo)
	if requestNo == "" || len(requestNo) > 96 {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	redeemType := in.RedeemType
	if redeemType == staking.RedeemType_REDEEM_TYPE_UNKNOWN {
		if order.Status == int64(staking.OrderStatus_ORDER_STATUS_EXPIRED) {
			redeemType = staking.RedeemType_REDEEM_TYPE_MATURITY
		} else {
			redeemType = staking.RedeemType_REDEEM_TYPE_EARLY
		}
	}
	if redeemType == staking.RedeemType_REDEEM_TYPE_EARLY && order.AllowEarlyRedeem != int64(common.YesNo_YES_NO_YES) {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.EarlyRedeemNotAllowed, i18n.Translate(i18n.EarlyRedeemNotAllowed, l.ctx))}, nil
	}
	if redeemType != staking.RedeemType_REDEEM_TYPE_EARLY && redeemType != staking.RedeemType_REDEEM_TYPE_MATURITY {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	feeRate := decimal.Zero
	operationType := helpers.StakeOperationTypeMaturityRedeem
	if redeemType == staking.RedeemType_REDEEM_TYPE_EARLY {
		feeRate = order.EarlyRedeemRate
		operationType = helpers.StakeOperationTypeEarlyRedeem
	}
	feeAmount := order.StakeAmount.Mul(feeRate).Div(decimal.NewFromInt(100)).RoundDown(8)
	spec := helpers.RedeemOperationSpec{
		RequestNo: requestNo, OperationType: operationType, RedeemType: redeemType,
		RewardAmount: order.PendingReward, FeeRate: feeRate, FeeAmount: feeAmount,
		OperatorId: userId, Remark: in.Remark,
	}
	if existing, findErr := l.svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(l.ctx, tenantId, userId, operationType, requestNo); findErr == nil {
		spec.OperationNo = existing.OperationNo
		operation, prepareErr := helpers.PrepareRedeemOperation(l.ctx, l.svcCtx, order, spec)
		if prepareErr != nil {
			return nil, prepareErr
		}
		if err := helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation); err != nil && !errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return nil, err
		}
		if operation.Status != helpers.StakeOperationStatusSucceeded {
			return redeemProcessingResp(operation.OperationNo, l.ctx), nil
		}
		return &staking.RedeemResp{Base: helper.OkResp(), Data: &staking.RedeemData{Success: 1, RedeemNo: operation.OperationNo}}, nil
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	if order.Status != int64(staking.OrderStatus_ORDER_STATUS_STAKING) && order.Status != int64(staking.OrderStatus_ORDER_STATUS_EXPIRED) {
		return &staking.RedeemResp{Base: helper.ErrResp(i18n.StakingOrderCannotRedeem, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}
	operationNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "SKR", "")
	if err != nil {
		return nil, err
	}
	spec.OperationNo = operationNo
	operation, err := helpers.PrepareRedeemOperation(l.ctx, l.svcCtx, order, spec)
	if errors.Is(err, helpers.ErrStakeOperationProcessing) {
		return redeemProcessingResp(operationNo, l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	if err := helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation); err != nil {
		if errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return redeemProcessingResp(operation.OperationNo, l.ctx), nil
		}
		return nil, err
	}

	if updated, findErr := l.svcCtx.StakeOrderModel.FindOne(l.ctx, order.Id); findErr == nil {
		publishStakingOrderChanged(l.ctx, l.svcCtx, updated)
	}
	return &staking.RedeemResp{
		Base: helper.OkResp(),
		Data: &staking.RedeemData{Success: 1, RedeemNo: operation.OperationNo},
	}, nil
}

func redeemProcessingResp(operationNo string, ctx context.Context) *staking.RedeemResp {
	return &staking.RedeemResp{
		Base: helper.ErrResp(i18n.AssetRequestProcessing, i18n.Translate(i18n.AssetRequestProcessing, ctx)),
		Data: &staking.RedeemData{RedeemNo: operationNo},
	}
}

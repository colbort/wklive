package adminlogic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/generate"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ManualRewardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewManualRewardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ManualRewardLogic {
	return &ManualRewardLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ManualReward gives every request its own durable operation number. Reusing
// request_no returns the original result and never records a reward without the
// matching Asset credit.
func (l *ManualRewardLogic) ManualReward(in *staking.ManualRewardReq) (*staking.ManualRewardResp, error) {
	order, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if in.TenantId > 0 && order.TenantId != in.TenantId {
		return manualRewardNotFound(l.ctx), nil
	}
	if base, scopeErr := helpers.AdminTenantWriteScopeResp(l.ctx, order.TenantId, i18n.OrderNotFound); scopeErr != nil {
		return nil, scopeErr
	} else if base != nil {
		return &staking.ManualRewardResp{Page: base}, nil
	}
	operatorId, err := helpers.AdminOperatorUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	rewardAmount, err := conv.ParseDecimalField(in.RewardAmount)
	if err != nil || !rewardAmount.IsPositive() {
		return &staking.ManualRewardResp{Page: helper.ErrResp(i18n.RewardAmountInvalid, i18n.Translate(i18n.RewardAmountInvalid, l.ctx))}, nil
	}
	if in.RewardType != staking.RewardType_REWARD_TYPE_MANUAL && in.RewardType != staking.RewardType_REWARD_TYPE_REISSUE {
		return &staking.ManualRewardResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	spec := helpers.RewardOperationSpec{
		RequestNo: in.RequestNo, OperationType: helpers.StakeOperationTypeManualReward,
		RewardType: in.RewardType, RewardAmount: rewardAmount, OperatorId: operatorId, Remark: in.Remark,
	}
	if existing, findErr := l.svcCtx.StakeOperationModel.FindOneByTenantIdUserIdOperationTypeRequestNo(l.ctx, order.TenantId, order.UserId, helpers.StakeOperationTypeManualReward, in.RequestNo); findErr == nil {
		spec.OperationNo = existing.OperationNo
		operation, prepareErr := helpers.PrepareRewardOperation(l.ctx, l.svcCtx, order, spec)
		if prepareErr != nil {
			return nil, prepareErr
		}
		if err := helpers.ExecuteRewardOperation(l.ctx, l.svcCtx, operation); err != nil && !errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return nil, err
		}
		if operation.Status != helpers.StakeOperationStatusSucceeded {
			return manualRewardProcessing(l.ctx), nil
		}
		return &staking.ManualRewardResp{Page: helper.OkResp(), Data: 1}, nil
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	}
	if order.Status != int64(staking.OrderStatus_ORDER_STATUS_STAKING) && order.Status != int64(staking.OrderStatus_ORDER_STATUS_EXPIRED) {
		return &staking.ManualRewardResp{Page: helper.ErrResp(i18n.StakingOrderCannotRedeem, i18n.Translate(i18n.StakingOrderCannotRedeem, l.ctx))}, nil
	}

	operationNo, err := generate.GenerateNo(l.svcCtx.Redis, l.ctx, "order_id", "SKW", "")
	if err != nil {
		return nil, err
	}
	spec.OperationNo = operationNo
	operation, err := helpers.PrepareRewardOperation(l.ctx, l.svcCtx, order, spec)
	if errors.Is(err, helpers.ErrStakeOperationProcessing) {
		return manualRewardProcessing(l.ctx), nil
	}
	if err != nil {
		return nil, err
	}
	if err := helpers.ExecuteRewardOperation(l.ctx, l.svcCtx, operation); err != nil {
		if errors.Is(err, helpers.ErrStakeOperationProcessing) {
			return manualRewardProcessing(l.ctx), nil
		}
		return nil, err
	}
	return &staking.ManualRewardResp{Page: helper.OkResp(), Data: 1}, nil
}

func manualRewardNotFound(ctx context.Context) *staking.ManualRewardResp {
	return &staking.ManualRewardResp{Page: helper.ErrResp(i18n.OrderNotFound, i18n.Translate(i18n.OrderNotFound, ctx))}
}

func manualRewardProcessing(ctx context.Context) *staking.ManualRewardResp {
	return &staking.ManualRewardResp{Page: helper.ErrResp(i18n.AssetRequestProcessing, i18n.Translate(i18n.AssetRequestProcessing, ctx))}
}

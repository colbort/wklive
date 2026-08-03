package tasklogic

import (
	"context"
	"errors"
	"fmt"

	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxRewardCatchUpPeriods = 3660

type ProcessRewardsAndSettleOrdersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProcessRewardsAndSettleOrdersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProcessRewardsAndSettleOrdersLogic {
	return &ProcessRewardsAndSettleOrdersLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ProcessRewardsAndSettleOrders first resumes durable operations and then
// discovers newly due periods/orders. The periodic scan is the authoritative
// recovery path; the delay queue is only a latency optimisation.
func (l *ProcessRewardsAndSettleOrdersLogic) ProcessRewardsAndSettleOrders(in *staking.StakingTaskReq) (*staking.StakingTaskResp, error) {
	lockName := fmt.Sprintf("process_rewards_and_settle_orders:%d", in.GetTenantId())
	return helpers.RunTaskWithLock(l.ctx, l.svcCtx, lockName, func() (*staking.StakingTaskResp, error) {
		now := utils.NowMillis()
		l.recoverOperations(in.GetTenantId(), now)

		cursor := int64(0)
		for {
			orders, _, err := l.svcCtx.StakeOrderModel.FindPage(l.ctx, models.StakeOrderPageFilter{
				TenantId: in.GetTenantId(), Status: int64(staking.OrderStatus_ORDER_STATUS_STAKING),
			}, cursor, 100)
			if err != nil {
				return nil, err
			}
			if len(orders) == 0 {
				break
			}
			for _, order := range orders {
				cursor = order.Id
				if err := l.processDailyRewards(order, now); err != nil && !errors.Is(err, helpers.ErrStakeOperationProcessing) {
					l.Errorf("staking daily reward processing deferred, orderId=%d err=%v", order.Id, err)
				}
				current, err := l.svcCtx.StakeOrderModel.FindOne(l.ctx, order.Id)
				if err != nil {
					l.Errorf("reload staking order failed, orderId=%d err=%v", order.Id, err)
					continue
				}
				if current.EndTimes > 0 && current.EndTimes <= now {
					if err := l.settleExpiredOrder(current, now); err != nil && !errors.Is(err, helpers.ErrStakeOperationProcessing) {
						l.Errorf("staking maturity settlement deferred, orderId=%d err=%v", current.Id, err)
					}
				}
			}
			if len(orders) < 100 {
				break
			}
		}
		return helpers.OkTaskResp(), nil
	})
}

func (l *ProcessRewardsAndSettleOrdersLogic) processDailyRewards(order *models.TStakeOrder, now int64) error {
	if order.RewardMode != int64(staking.RewardMode_REWARD_MODE_DAILY) || order.NextRewardTimes <= 0 {
		return nil
	}
	for periods := 0; periods < maxRewardCatchUpPeriods; periods++ {
		periodEnd := order.NextRewardTimes
		if periodEnd > now || (order.EndTimes > 0 && periodEnd > order.EndTimes) {
			return nil
		}
		rewardAmount := helpers.CalcTaskReward(order, 1)
		if !rewardAmount.IsPositive() {
			return nil
		}
		operationNo := dailyRewardBizNo(order.Id, periodEnd)
		operation, err := helpers.PrepareRewardOperation(l.ctx, l.svcCtx, order, helpers.RewardOperationSpec{
			OperationNo: operationNo, RequestNo: operationNo,
			OperationType: helpers.StakeOperationTypeDailyReward, RewardType: staking.RewardType_REWARD_TYPE_DAILY,
			RewardAmount: rewardAmount, PeriodEnd: periodEnd, Remark: "staking daily reward task",
		})
		if err != nil {
			return err
		}
		if err := helpers.ExecuteRewardOperation(l.ctx, l.svcCtx, operation); err != nil {
			return err
		}
		order, err = l.svcCtx.StakeOrderModel.FindOne(l.ctx, order.Id)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("daily reward catch-up exceeded %d periods", maxRewardCatchUpPeriods)
}

func (l *ProcessRewardsAndSettleOrdersLogic) settleExpiredOrder(order *models.TStakeOrder, _ int64) error {
	rewardAmount := order.PendingReward
	if order.RewardMode == int64(staking.RewardMode_REWARD_MODE_MATURITY) {
		days := order.LockDays
		if days <= 0 {
			days = 1
		}
		rewardAmount = rewardAmount.Add(helpers.CalcTaskReward(order, days))
	}
	operationNo := maturityRedeemBizNo(order.Id, order.EndTimes)
	operation, err := helpers.PrepareRedeemOperation(l.ctx, l.svcCtx, order, helpers.RedeemOperationSpec{
		OperationNo: operationNo, RequestNo: operationNo,
		OperationType: helpers.StakeOperationTypeMaturityRedeem,
		RedeemType:    staking.RedeemType_REDEEM_TYPE_MATURITY,
		RewardAmount:  rewardAmount, OperatorId: 0, Remark: "staking maturity redeem task",
	})
	if err != nil {
		return err
	}
	if err := helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation); err != nil {
		return err
	}
	if updated, findErr := l.svcCtx.StakeOrderModel.FindOne(l.ctx, order.Id); findErr == nil {
		publishStakingOrderChanged(l.ctx, l.svcCtx, updated)
	}
	return nil
}

func (l *ProcessRewardsAndSettleOrdersLogic) recoverOperations(tenantId, now int64) {
	cursor := int64(0)
	for {
		operations, err := l.svcCtx.StakeOperationModel.FindRetryablePage(l.ctx, tenantId, now, cursor, 100)
		if err != nil {
			l.Errorf("list staking retry operations failed: %v", err)
			return
		}
		if len(operations) == 0 {
			return
		}
		for _, operation := range operations {
			cursor = operation.Id
			var executeErr error
			switch operation.OperationType {
			case helpers.StakeOperationTypeSubscribe:
				executeErr = helpers.ExecuteSubscribeOperation(l.ctx, l.svcCtx, operation)
			case helpers.StakeOperationTypeDailyReward, helpers.StakeOperationTypeMaturityReward, helpers.StakeOperationTypeManualReward:
				executeErr = helpers.ExecuteRewardOperation(l.ctx, l.svcCtx, operation)
			case helpers.StakeOperationTypeMaturityRedeem, helpers.StakeOperationTypeEarlyRedeem, helpers.StakeOperationTypeManualRedeem:
				executeErr = helpers.ExecuteRedeemOperation(l.ctx, l.svcCtx, operation)
			default:
				executeErr = fmt.Errorf("unsupported staking operation type %d", operation.OperationType)
			}
			if executeErr != nil && !errors.Is(executeErr, helpers.ErrStakeOperationProcessing) {
				l.Errorf("recover staking operation failed, operationNo=%s err=%v", operation.OperationNo, executeErr)
			}
		}
		if len(operations) < 100 {
			return
		}
	}
}

func dailyRewardBizNo(orderId, periodEnd int64) string {
	return fmt.Sprintf("SKW_%d_%d", orderId, periodEnd)
}

func maturityRedeemBizNo(orderId, endTimes int64) string {
	return fmt.Sprintf("SKR_%d_%d", orderId, endTimes)
}

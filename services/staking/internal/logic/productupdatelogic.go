package logic

import (
	"context"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductUpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductUpdateLogic {
	return &ProductUpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新质押产品
func (l *ProductUpdateLogic) ProductUpdate(in *staking.AdminProductUpdateReq) (*staking.AdminProductUpdateResp, error) {
	item, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(l.ctx, item.TenantId)
	if err != nil {
		return nil, i18n.StatusError(l.ctx, i18n.UserNotFound)
	}
	if forbidden {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, l.ctx))}, nil
	}
	if !allowed {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}

	if in.ProductName != "" {
		item.ProductName = in.ProductName
	}
	if in.ProductType != staking.ProductType_PRODUCT_TYPE_UNKNOWN {
		item.ProductType = int64(in.ProductType)
	}
	if in.CoinName != "" {
		item.CoinName = in.CoinName
	}
	if in.CoinSymbol != "" {
		item.CoinSymbol = in.CoinSymbol
	}
	if in.RewardCoinName != "" {
		item.RewardCoinName = in.RewardCoinName
	}
	if in.RewardCoinSymbol != "" {
		item.RewardCoinSymbol = in.RewardCoinSymbol
	}
	if in.Apr != "" {
		apr, err := conv.ParseFloatField(in.Apr)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.Apr = apr
	}
	if in.LockDays != 0 {
		item.LockDays = int64(in.LockDays)
	}
	if in.MinAmount != "" {
		minAmount, err := conv.ParseFloatField(in.MinAmount)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MinAmount = minAmount
	}
	if in.MaxAmount != "" {
		maxAmount, err := conv.ParseFloatField(in.MaxAmount)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.MaxAmount = maxAmount
	}
	if in.StepAmount != "" {
		stepAmount, err := conv.ParseFloatField(in.StepAmount)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.StepAmount = stepAmount
	}
	if in.TotalAmount != "" {
		totalAmount, err := conv.ParseFloatField(in.TotalAmount)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.TotalAmount = totalAmount
	}
	if in.UserLimitAmount != "" {
		userLimitAmount, err := conv.ParseFloatField(in.UserLimitAmount)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.UserLimitAmount = userLimitAmount
	}
	if in.InterestMode != staking.InterestMode_INTEREST_MODE_UNKNOWN {
		item.InterestMode = int64(in.InterestMode)
	}
	if in.RewardMode != staking.RewardMode_REWARD_MODE_UNKNOWN {
		item.RewardMode = int64(in.RewardMode)
	}
	if in.AllowEarlyRedeem != 0 {
		item.AllowEarlyRedeem = int64(in.AllowEarlyRedeem)
	}
	if in.EarlyRedeemRate != "" {
		earlyRedeemRate, err := conv.ParseFloatField(in.EarlyRedeemRate)
		if err != nil {
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
		}
		item.EarlyRedeemRate = earlyRedeemRate
	}
	if in.Status != staking.ProductStatus_PRODUCT_STATUS_UNKNOWN {
		item.Status = int64(in.Status)
	}
	if in.Sort != 0 {
		item.Sort = int64(in.Sort)
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	if in.OperatorUid != 0 {
		item.UpdateUserId = in.OperatorUid
	}
	item.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.StakeProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &staking.AdminProductUpdateResp{Page: helper.OkResp(), Data: 1}, nil
}

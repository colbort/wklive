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
			return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}

	apr, err := conv.ParseFloatField(in.Apr)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	minAmount, err := conv.ParseFloatField(in.MinAmount)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	maxAmount, err := conv.ParseFloatField(in.MaxAmount)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	stepAmount, err := conv.ParseFloatField(in.StepAmount)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	totalAmount, err := conv.ParseFloatField(in.TotalAmount)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	userLimitAmount, err := conv.ParseFloatField(in.UserLimitAmount)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	earlyRedeemRate, err := conv.ParseFloatField(in.EarlyRedeemRate)
	if err != nil {
		return &staking.AdminProductUpdateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	item.ProductName = in.ProductName
	item.ProductType = int64(in.ProductType)
	item.CoinName = in.CoinName
	item.CoinSymbol = in.CoinSymbol
	item.RewardCoinName = in.RewardCoinName
	item.RewardCoinSymbol = in.RewardCoinSymbol
	item.Apr = apr
	item.LockDays = int64(in.LockDays)
	item.MinAmount = minAmount
	item.MaxAmount = maxAmount
	item.StepAmount = stepAmount
	item.TotalAmount = totalAmount
	item.UserLimitAmount = userLimitAmount
	item.InterestMode = int64(in.InterestMode)
	item.RewardMode = int64(in.RewardMode)
	item.AllowEarlyRedeem = int64(in.AllowEarlyRedeem)
	item.EarlyRedeemRate = earlyRedeemRate
	item.Status = int64(in.Status)
	item.Sort = int64(in.Sort)
	item.Remark = in.Remark
	item.UpdateUserId = in.OperatorUid
	item.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.StakeProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &staking.AdminProductUpdateResp{Page: helper.OkResp(), Data: 1}, nil
}

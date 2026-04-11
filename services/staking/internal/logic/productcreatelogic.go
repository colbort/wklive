package logic

import (
	"context"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductCreateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductCreateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductCreateLogic {
	return &ProductCreateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 创建质押产品
func (l *ProductCreateLogic) ProductCreate(in *staking.AdminProductCreateReq) (*staking.AdminProductCreateResp, error) {
	exists, err := l.svcCtx.StakeProductModel.FindOneByTenantIdProductNo(l.ctx, in.TenantId, in.ProductNo)
	if err == nil && exists != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ProductNoAlreadyExists, l.ctx))}, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	apr, err := conv.ParseFloatField(in.Apr)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	minAmount, err := conv.ParseFloatField(in.MinAmount)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	maxAmount, err := conv.ParseFloatField(in.MaxAmount)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	stepAmount, err := conv.ParseFloatField(in.StepAmount)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	totalAmount, err := conv.ParseFloatField(in.TotalAmount)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	userLimitAmount, err := conv.ParseFloatField(in.UserLimitAmount)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	earlyRedeemRate, err := conv.ParseFloatField(in.EarlyRedeemRate)
	if err != nil {
		return &staking.AdminProductCreateResp{Page: helper.GetErrResp(400, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	now := utils.NowMillis()
	res, err := l.svcCtx.StakeProductModel.Insert(l.ctx, &models.TStakeProduct{
		TenantId:         in.TenantId,
		ProductNo:        in.ProductNo,
		ProductName:      in.ProductName,
		ProductType:      int64(in.ProductType),
		CoinName:         in.CoinName,
		CoinSymbol:       in.CoinSymbol,
		RewardCoinName:   in.RewardCoinName,
		RewardCoinSymbol: in.RewardCoinSymbol,
		Apr:              apr,
		LockDays:         int64(in.LockDays),
		MinAmount:        minAmount,
		MaxAmount:        maxAmount,
		StepAmount:       stepAmount,
		TotalAmount:      totalAmount,
		StakedAmount:     0,
		UserLimitAmount:  userLimitAmount,
		InterestMode:     int64(in.InterestMode),
		RewardMode:       int64(in.RewardMode),
		AllowEarlyRedeem: int64(in.AllowEarlyRedeem),
		EarlyRedeemRate:  earlyRedeemRate,
		Status:           int64(in.Status),
		Sort:             int64(in.Sort),
		Remark:           in.Remark,
		CreateUserId:     in.OperatorUid,
		UpdateUserId:     in.OperatorUid,
		CreateTimes:      now,
		UpdateTimes:      now,
	})
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &staking.AdminProductCreateResp{Page: helper.OkResp(), Data: id}, nil
}

package adminlogic

import (
	"context"
	"strings"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/logic/helpers"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/shopspring/decimal"
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
func (l *ProductCreateLogic) ProductCreate(in *staking.ProductCreateReq) (*staking.ProductCreateResp, error) {
	if base, err := helpers.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.PermissionDenied); err != nil {
		return nil, err
	} else if base != nil {
		return &staking.ProductCreateResp{Page: base}, nil
	}
	operatorId, err := helpers.AdminOperatorUserID(l.ctx)
	if err != nil {
		return nil, err
	}
	exists, err := l.svcCtx.StakeProductModel.FindOneByTenantIdProductNo(l.ctx, in.TenantId, in.ProductNo)
	if err == nil && exists != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ProductNoAlreadyExists, i18n.Translate(i18n.ProductNoAlreadyExists, l.ctx))}, nil
	}
	if err != nil && err != models.ErrNotFound {
		return nil, err
	}

	apr, err := conv.ParseDecimalField(in.Apr)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	minAmount, err := conv.ParseDecimalField(in.MinAmount)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	maxAmount, err := conv.ParseDecimalField(in.MaxAmount)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	stepAmount, err := conv.ParseDecimalField(in.StepAmount)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	totalAmount, err := conv.ParseDecimalField(in.TotalAmount)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	userLimitAmount, err := conv.ParseDecimalField(in.UserLimitAmount)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	earlyRedeemRate, err := conv.ParseDecimalField(in.EarlyRedeemRate)
	if err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}

	now := utils.NowMillis()
	product := &models.TStakeProduct{
		TenantId:         in.TenantId,
		ProductNo:        strings.TrimSpace(in.ProductNo),
		ProductName:      strings.TrimSpace(in.ProductName),
		ProductType:      int64(in.ProductType),
		CoinName:         in.CoinName,
		CoinSymbol:       strings.ToUpper(strings.TrimSpace(in.CoinSymbol)),
		RewardCoinName:   in.RewardCoinName,
		RewardCoinSymbol: strings.ToUpper(strings.TrimSpace(in.RewardCoinSymbol)),
		Apr:              apr,
		LockDays:         int64(in.LockDays),
		MinAmount:        minAmount,
		MaxAmount:        maxAmount,
		StepAmount:       stepAmount,
		TotalAmount:      totalAmount,
		StakedAmount:     decimal.Zero,
		UserLimitAmount:  userLimitAmount,
		InterestMode:     int64(in.InterestMode),
		RewardMode:       int64(in.RewardMode),
		AllowEarlyRedeem: int64(in.AllowEarlyRedeem),
		EarlyRedeemRate:  earlyRedeemRate,
		Status:           int64(in.Status),
		Sort:             int64(in.Sort),
		Remark:           in.Remark,
		CreateUserId:     operatorId,
		UpdateUserId:     operatorId,
		CreateTimes:      now,
		UpdateTimes:      now,
	}
	if err := helpers.ValidateStakeProduct(product); err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	if err := helpers.ValidateStakeFundingAccounts(l.ctx, l.svcCtx, product); err != nil {
		return &staking.ProductCreateResp{Page: helper.ErrResp(i18n.ParamError, i18n.Translate(i18n.ParamError, l.ctx))}, nil
	}
	res, err := l.svcCtx.StakeProductModel.Insert(l.ctx, product)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &staking.ProductCreateResp{Page: helper.OkResp(), Data: id}, nil
}

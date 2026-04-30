package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetMyCryptoRechargeAddressLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetMyCryptoRechargeAddressLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetMyCryptoRechargeAddressLogic {
	return &GetMyCryptoRechargeAddressLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取/分配我的链上充值地址
func (l *GetMyCryptoRechargeAddressLogic) GetMyCryptoRechargeAddress(in *payment.GetMyCryptoRechargeAddressReq) (*payment.GetMyCryptoRechargeAddressResp, error) {
	userId, err := utils.GetUserIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	tenantId, err := utils.GetTenantIdFromMd(l.ctx)
	if err != nil {
		return nil, err
	}
	item, err := l.svcCtx.CryptoRechargeAddressModel.FindOneByTenantIdUserIdWalletTypeCoinChainCode(l.ctx, tenantId, userId, in.WalletType, in.Coin, int64(in.ChainCode))
	if err != nil {
		if !errors.Is(err, models.ErrNotFound) {
			return nil, err
		}
		item, err = l.svcCtx.CryptoRechargeAddressModel.FindOneAssignable(l.ctx, tenantId, in.WalletType, in.Coin, int64(in.ChainCode))
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return &payment.GetMyCryptoRechargeAddressResp{Base: helper.GetErrResp(404, "no available crypto recharge address")}, nil
			}
			return nil, err
		}

		now := utils.NowMillis()
		item.UserId = userId
		item.WalletType = in.WalletType
		item.Coin = in.Coin
		item.ChainCode = int64(in.ChainCode)
		item.Status = 1
		item.LastUsedTime = now
		item.UpdateTimes = now
		if item.CreateTimes == 0 {
			item.CreateTimes = now
		}
		if err := l.svcCtx.CryptoRechargeAddressModel.Update(l.ctx, item); err != nil {
			return nil, err
		}
	}

	return &payment.GetMyCryptoRechargeAddressResp{Base: helper.OkResp(), Data: toCryptoRechargeAddressProto(item)}, nil
}

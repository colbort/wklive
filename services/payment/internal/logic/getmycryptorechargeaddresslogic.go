package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

var errCryptoRechargeAddressNotConfigured = errors.New("crypto recharge address not configured")

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
		item, err = l.reserveAssignableAddress(tenantId, userId, in)
		if err != nil {
			if errors.Is(err, errCryptoRechargeAddressNotConfigured) {
				return &payment.GetMyCryptoRechargeAddressResp{Base: helper.GetErrResp(404, i18n.Translate(i18n.CryptoRechargeAddressNotConfigured, l.ctx))}, nil
			}
			if errors.Is(err, models.ErrNotFound) {
				return &payment.GetMyCryptoRechargeAddressResp{Base: helper.GetErrResp(201, i18n.Translate(i18n.CryptoRechargeAddressInUse, l.ctx))}, nil
			}
			return nil, err
		}
		return &payment.GetMyCryptoRechargeAddressResp{Base: helper.OkResp(), Data: toCryptoRechargeAddressProto(item)}, nil
	}

	ok, err := reserveCryptoRechargeAddress(l.ctx, l.svcCtx, item, tenantId, userId)
	if err != nil {
		return nil, err
	}
	if !ok {
		reservedByMe, err := cryptoRechargeAddressReservedBy(l.ctx, l.svcCtx, item.Id, tenantId, userId)
		if err != nil {
			return nil, err
		}
		if !reservedByMe {
			return &payment.GetMyCryptoRechargeAddressResp{Base: helper.GetErrResp(201, i18n.Translate(i18n.CryptoRechargeAddressInUse, l.ctx))}, nil
		}
		refreshCryptoRechargeAddressReservation(l.ctx, l.svcCtx, item.Id)
	}
	return &payment.GetMyCryptoRechargeAddressResp{Base: helper.OkResp(), Data: toCryptoRechargeAddressProto(item)}, nil
}

func (l *GetMyCryptoRechargeAddressLogic) reserveAssignableAddress(tenantId, userId int64, in *payment.GetMyCryptoRechargeAddressReq) (*models.TCryptoRechargeAddress, error) {
	now := utils.NowMillis()
	reusableBefore := now - cryptoRechargeAddressHoldSeconds*1000
	items, err := l.svcCtx.CryptoRechargeAddressModel.FindAssignableCandidates(l.ctx, tenantId, in.WalletType, in.Coin, int64(in.ChainCode), reusableBefore, 50)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		hasAddress, err := l.svcCtx.CryptoRechargeAddressModel.HasEnabledAddress(l.ctx, tenantId, in.WalletType, in.Coin, int64(in.ChainCode))
		if err != nil {
			return nil, err
		}
		if !hasAddress {
			return nil, errCryptoRechargeAddressNotConfigured
		}
		return nil, models.ErrNotFound
	}
	for _, item := range items {
		ok, err := reserveCryptoRechargeAddress(l.ctx, l.svcCtx, item, tenantId, userId)
		if err != nil {
			return nil, err
		}
		if !ok {
			reservedByMe, err := cryptoRechargeAddressReservedBy(l.ctx, l.svcCtx, item.Id, tenantId, userId)
			if err != nil {
				return nil, err
			}
			if !reservedByMe {
				continue
			}
			refreshCryptoRechargeAddressReservation(l.ctx, l.svcCtx, item.Id)
			return item, nil
		}

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
			releaseCryptoRechargeAddress(l.ctx, l.svcCtx, item.Id)
			return nil, err
		}
		return item, nil
	}
	return nil, models.ErrNotFound
}

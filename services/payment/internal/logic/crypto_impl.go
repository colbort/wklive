package logic

import (
	"context"
	"database/sql"
	"errors"

	"wklive/common/conv"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/pageutil"
	"wklive/common/utils"
	"wklive/proto/common"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

func createCryptoRechargeAddress(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.CreateCryptoRechargeAddressReq) (*payment.AdminCommonResp, error) {
	now := utils.NowMillis()
	status := toCryptoAddressStatusDB(in.Status, int64(payment.CryptoRechargeAddressStatus_CRYPTO_RECHARGE_ADDRESS_STATUS_ENABLED))
	if in.AddressSource == 0 {
		in.AddressSource = payment.CryptoRechargeAddressSource_CRYPTO_RECHARGE_ADDRESS_SOURCE_MANUAL
	}
	if in.AddressType == 0 {
		in.AddressType = payment.CryptoRechargeAddressType_CRYPTO_RECHARGE_ADDRESS_TYPE_EXCLUSIVE
	}
	_, err := svcCtx.CryptoRechargeAddressModel.Insert(ctx, &models.TCryptoRechargeAddress{
		TenantId:      in.TenantId,
		UserId:        in.UserId,
		WalletType:    int64(in.WalletType),
		Coin:          in.Coin,
		ChainCode:     int64(in.ChainCode),
		Address:       in.Address,
		Memo:          in.Memo,
		AddressSource: int64(in.AddressSource),
		AddressType:   int64(in.AddressType),
		Status:        status,
		CreateTimes:   now,
		UpdateTimes:   now,
	})
	if err != nil {
		return nil, err
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func updateCryptoRechargeAddress(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.UpdateCryptoRechargeAddressReq) (*payment.AdminCommonResp, error) {
	item, err := svcCtx.CryptoRechargeAddressModel.FindOne(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(ctx, item.TenantId, i18n.CryptoRechargeAddressNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}

	if in.Address != "" {
		item.Address = in.Address
	}
	if in.Memo != "" {
		item.Memo = in.Memo
	}
	if in.AddressSource != 0 {
		item.AddressSource = int64(in.AddressSource)
	}
	if in.AddressType != 0 {
		item.AddressType = int64(in.AddressType)
	}
	item.Status = toCryptoAddressStatusDB(in.Status, item.Status)
	item.UpdateTimes = utils.NowMillis()
	if err := svcCtx.CryptoRechargeAddressModel.Update(ctx, item); err != nil {
		return nil, err
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func getCryptoRechargeAddress(ctx context.Context, svcCtx *svc.ServiceContext, tenantId int64, id int64) (*models.TCryptoRechargeAddress, error) {
	item, err := svcCtx.CryptoRechargeAddressModel.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}
	if tenantId > 0 && item.TenantId != tenantId {
		return nil, models.ErrNotFound
	}
	return item, nil
}

func listCryptoRechargeAddresses(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.ListCryptoRechargeAddressesReq) (*payment.ListCryptoRechargeAddressesResp, error) {
	items, total, err := svcCtx.CryptoRechargeAddressModel.FindPage(ctx, models.CryptoRechargeAddressPageFilter{
		TenantId:    in.TenantId,
		UserId:      in.UserId,
		WalletType:  int64(in.WalletType),
		Coin:        in.Coin,
		ChainCode:   int64(in.ChainCode),
		Address:     in.Address,
		Status:      int64(in.Status),
		AddressType: int64(in.AddressType),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*payment.CryptoRechargeAddress, 0, len(items))
	for _, item := range items {
		data = append(data, toCryptoRechargeAddressProto(item))
	}
	return &payment.ListCryptoRechargeAddressesResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastCryptoAddressID(items)),
		Data: data,
	}, nil
}

func createCryptoWalletAccount(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.CreateCryptoWalletAccountReq) (*payment.AdminCommonResp, error) {
	now := utils.NowMillis()
	isDefault := int64(in.IsDefault)
	if common.YesNo(in.IsDefault) == common.YesNo_YES_NO_UNKNOWN {
		isDefault = int64(common.YesNo_YES_NO_NO)
	}
	_, err := svcCtx.CryptoWalletAccountModel.Insert(ctx, &models.TCryptoWalletAccount{
		TenantId:             in.TenantId,
		AccountCode:          in.AccountCode,
		AccountName:          in.AccountName,
		Provider:             in.Provider,
		ApiKeyCipher:         nullableString(in.ApiKeyCipher),
		ApiSecretCipher:      nullableString(in.ApiSecretCipher),
		CallbackSecretCipher: nullableString(in.CallbackSecretCipher),
		ExtConfig:            nullableString(in.ExtConfig),
		Enabled:              toCryptoWalletStatusDB(in.Enabled, int64(common.Enable_ENABLE_ENABLED)),
		IsDefault:            isDefault,
		CreateTimes:          now,
		UpdateTimes:          now,
	})
	if err != nil {
		return nil, err
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func updateCryptoWalletAccount(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.UpdateCryptoWalletAccountReq) (*payment.AdminCommonResp, error) {
	item, err := svcCtx.CryptoWalletAccountModel.FindOne(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(ctx, item.TenantId, i18n.CryptoWalletAccountNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}
	if in.AccountName != "" {
		item.AccountName = in.AccountName
	}
	if in.Provider != "" {
		item.Provider = in.Provider
	}
	if in.ApiKeyCipher != "" {
		item.ApiKeyCipher = nullableString(in.ApiKeyCipher)
	}
	if in.ApiSecretCipher != "" {
		item.ApiSecretCipher = nullableString(in.ApiSecretCipher)
	}
	if in.CallbackSecretCipher != "" {
		item.CallbackSecretCipher = nullableString(in.CallbackSecretCipher)
	}
	if in.ExtConfig != "" {
		item.ExtConfig = nullableString(in.ExtConfig)
	}
	item.Enabled = toCryptoWalletStatusDB(in.Enabled, item.Enabled)
	if common.YesNo(in.IsDefault) != common.YesNo_YES_NO_UNKNOWN {
		item.IsDefault = int64(in.IsDefault)
	}
	item.UpdateTimes = utils.NowMillis()
	if err := svcCtx.CryptoWalletAccountModel.Update(ctx, item); err != nil {
		return nil, err
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func listCryptoWalletAccounts(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.ListCryptoWalletAccountsReq) (*payment.ListCryptoWalletAccountsResp, error) {
	items, total, err := svcCtx.CryptoWalletAccountModel.FindPage(ctx, models.CryptoWalletAccountPageFilter{
		TenantId:  in.TenantId,
		Keyword:   in.Keyword,
		Provider:  in.Provider,
		Enabled:   int64(in.Enabled),
		IsDefault: int64(in.IsDefault),
	}, in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*payment.CryptoWalletAccount, 0, len(items))
	for _, item := range items {
		data = append(data, toCryptoWalletAccountProto(item))
	}
	return &payment.ListCryptoWalletAccountsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastCryptoWalletAccountID(items)),
		Data: data,
	}, nil
}

func createCryptoRechargeTx(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.CreateCryptoRechargeTxReq) (*payment.AdminCommonResp, error) {
	amount, err := conv.ParseFloatField(in.Amount)
	if err != nil {
		return nil, err
	}
	now := utils.NowMillis()
	_, err = svcCtx.CryptoRechargeTxModel.Insert(ctx, &models.TCryptoRechargeTx{
		TenantId:             in.TenantId,
		UserId:               in.UserId,
		OrderId:              in.OrderId,
		OrderNo:              in.OrderNo,
		Coin:                 in.Coin,
		ChainCode:            int64(in.ChainCode),
		TxHash:               in.TxHash,
		FromAddress:          in.FromAddress,
		ToAddress:            in.ToAddress,
		Memo:                 in.Memo,
		Amount:               amount,
		BlockHeight:          in.BlockHeight,
		ConfirmCount:         in.ConfirmCount,
		RequiredConfirmCount: in.RequiredConfirmCount,
		Status:               int64(in.Status),
		RawData:              nullableString(in.RawData),
		CreateTimes:          now,
		UpdateTimes:          now,
	})
	if err != nil {
		return nil, err
	}
	if in.Status == payment.CryptoRechargeTxStatus_CRYPTO_RECHARGE_TX_STATUS_CREDITED {
		if err := creditCryptoRechargeOrder(ctx, svcCtx, in.OrderNo, in.TxHash); err != nil {
			return nil, err
		}
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func updateCryptoRechargeTx(ctx context.Context, svcCtx *svc.ServiceContext, in *payment.UpdateCryptoRechargeTxReq) (*payment.AdminCommonResp, error) {
	item, err := svcCtx.CryptoRechargeTxModel.FindOne(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	allowTenantUpdate, resp, err := applyAdminTenantUpdateScope(ctx, item.TenantId, i18n.CryptoRechargeTxNotFound)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return resp, nil
	}
	if allowTenantUpdate {
		item.TenantId = in.TenantId
	}
	if in.OrderId > 0 {
		item.OrderId = in.OrderId
	}
	if in.OrderNo != "" {
		item.OrderNo = in.OrderNo
	}
	if in.ConfirmCount > 0 {
		item.ConfirmCount = in.ConfirmCount
	}
	if in.RequiredConfirmCount > 0 {
		item.RequiredConfirmCount = in.RequiredConfirmCount
	}
	if in.Status != 0 {
		item.Status = int64(in.Status)
	}
	if in.RawData != "" {
		item.RawData = nullableString(in.RawData)
	}
	if payment.CryptoRechargeTxStatus(item.Status) == payment.CryptoRechargeTxStatus_CRYPTO_RECHARGE_TX_STATUS_CREDITED {
		if err := creditCryptoRechargeOrder(ctx, svcCtx, item.OrderNo, item.TxHash); err != nil {
			return nil, err
		}
	}
	item.UpdateTimes = utils.NowMillis()
	if err := svcCtx.CryptoRechargeTxModel.Update(ctx, item); err != nil {
		return nil, err
	}
	return &payment.AdminCommonResp{Base: helper.OkResp()}, nil
}

func creditCryptoRechargeOrder(ctx context.Context, svcCtx *svc.ServiceContext, orderNo string, txHash string) error {
	if orderNo == "" {
		return nil
	}
	order, err := svcCtx.RechargeOrderModel.FindOneByOrderNo(ctx, orderNo)
	if err != nil {
		return err
	}
	return markRechargeOrderSuccessAndCredit(ctx, svcCtx, order, txHash, 0, "crypto recharge credited")
}

func listCryptoRechargeTxs(ctx context.Context, svcCtx *svc.ServiceContext, req listCryptoTxReq) ([]*models.TCryptoRechargeTx, int64, error) {
	return svcCtx.CryptoRechargeTxModel.FindPage(ctx, models.CryptoRechargeTxPageFilter{
		TenantId:        req.tenantId,
		UserId:          req.userId,
		OrderNo:         req.orderNo,
		Coin:            req.coin,
		ChainCode:       int64(req.chainCode),
		TxHash:          req.txHash,
		ToAddress:       req.toAddress,
		Status:          int64(req.status),
		CreateTimeStart: req.createTimeStart,
		CreateTimeEnd:   req.createTimeEnd,
	}, req.cursor, req.limit)
}

type listCryptoTxReq struct {
	tenantId        int64
	userId          int64
	orderNo         string
	coin            string
	chainCode       common.ChainCode
	txHash          string
	toAddress       string
	status          payment.CryptoRechargeTxStatus
	createTimeStart int64
	createTimeEnd   int64
	cursor          int64
	limit           int64
}

func applyAdminTenantUpdateScope(
	ctx context.Context,
	currentTenantId int64,
	notFoundCode int32,
) (bool, *payment.AdminCommonResp, error) {
	allowTenantUpdate, allowed, forbidden, err := utils.ResolveAdminTenantWriteScopeFromMd(ctx, currentTenantId)
	if err != nil {
		return false, nil, i18n.StatusError(ctx, i18n.UserNotFound)
	}
	if forbidden {
		return false, &payment.AdminCommonResp{
			Base: helper.GetErrResp(i18n.PermissionDenied, i18n.Translate(i18n.PermissionDenied, ctx)),
		}, nil
	}
	if !allowed {
		return false, &payment.AdminCommonResp{
			Base: helper.GetErrResp(notFoundCode, i18n.Translate(notFoundCode, ctx)),
		}, nil
	}
	return allowTenantUpdate, nil, nil
}

func nullableString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func lastCryptoAddressID(items []*models.TCryptoRechargeAddress) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[len(items)-1].Id
}

func lastCryptoWalletAccountID(items []*models.TCryptoWalletAccount) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[len(items)-1].Id
}

func lastCryptoTxID(items []*models.TCryptoRechargeTx) int64 {
	if len(items) == 0 {
		return 0
	}
	return items[len(items)-1].Id
}

func cryptoNotFoundResp() *payment.AdminCommonResp {
	return &payment.AdminCommonResp{Base: helper.GetErrResp(i18n.NotFound, i18n.Translate(i18n.NotFound, context.Background()))}
}

func isNotFound(err error) bool {
	return errors.Is(err, models.ErrNotFound)
}

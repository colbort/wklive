package wechat

import (
	"context"
	"errors"
	"fmt"

	"wklive/services/payment/internal/provider"
	"wklive/services/payment/models"
)

var ErrNotImplemented = errors.New("wechat pay adapter is not implemented")

var _ provider.Adapter = (*Adapter)(nil)

// Adapter is the WeChat Pay provider implementation template. Replace each
// ErrNotImplemented return with the corresponding WeChat Pay SDK call,
// signature verification and response mapping before registering it.
type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) CreatePayment(
	ctx context.Context,
	account *models.TTenantPayAccount,
	order *models.TRechargeOrder,
) (*provider.CreatePaymentResult, error) {
	if err := validatePayment(account, order); err != nil {
		return nil, err
	}
	return requestPayment(ctx, account, order)
}

func (a *Adapter) PaymentNotify(
	ctx context.Context,
	account *models.TTenantPayAccount,
	request provider.NotifyRequest,
) (*provider.PaymentNotifyResult, error) {
	if err := validateAccount(account); err != nil {
		return nil, err
	}
	if len(request.Body) == 0 {
		return nil, fmt.Errorf("wechat payment notification body is empty")
	}
	return handlePaymentNotify(ctx, account, request)
}

func (a *Adapter) QueryPayment(
	ctx context.Context,
	account *models.TTenantPayAccount,
	order *models.TRechargeOrder,
) (*provider.PaymentQueryResult, error) {
	if err := validatePayment(account, order); err != nil {
		return nil, err
	}
	return requestPaymentQuery(ctx, account, order)
}

func (a *Adapter) RefundPayment(
	ctx context.Context,
	account *models.TTenantPayAccount,
	order *models.TRechargeOrder,
	request provider.RefundRequest,
) (*provider.RefundResult, error) {
	if err := validatePayment(account, order); err != nil {
		return nil, err
	}
	if request.RefundNo == "" || !request.Amount.IsPositive() {
		return nil, fmt.Errorf("wechat refund number and amount are required")
	}
	return requestRefund(ctx, account, order, request)
}

func (a *Adapter) CreatePayout(
	ctx context.Context,
	account *models.TTenantPayAccount,
	order *models.TWithdrawOrder,
) (*provider.CreatePayoutResult, error) {
	if err := validatePayout(account, order); err != nil {
		return nil, err
	}
	return requestPayout(ctx, account, order)
}

func (a *Adapter) PayoutNotify(
	ctx context.Context,
	account *models.TTenantPayAccount,
	request provider.NotifyRequest,
) (*provider.PayoutNotifyResult, error) {
	if err := validateAccount(account); err != nil {
		return nil, err
	}
	if len(request.Body) == 0 {
		return nil, fmt.Errorf("wechat payout notification body is empty")
	}
	return handlePayoutNotify(ctx, account, request)
}

func (a *Adapter) QueryPayout(
	ctx context.Context,
	account *models.TTenantPayAccount,
	order *models.TWithdrawOrder,
) (*provider.PayoutQueryResult, error) {
	if err := validatePayout(account, order); err != nil {
		return nil, err
	}
	return requestPayoutQuery(ctx, account, order)
}

func validateAccount(account *models.TTenantPayAccount) error {
	if account == nil {
		return fmt.Errorf("wechat account is required")
	}
	if account.Enabled != 1 {
		return fmt.Errorf("wechat account is disabled")
	}
	if !account.AppId.Valid || account.AppId.String == "" {
		return fmt.Errorf("wechat app_id is required")
	}
	if !account.MerchantId.Valid || account.MerchantId.String == "" {
		return fmt.Errorf("wechat merchant_id is required")
	}
	if account.CredentialRef == "" && (!account.PrivateKeyCipher.Valid || account.PrivateKeyCipher.String == "") {
		return fmt.Errorf("wechat credential_ref or merchant private key is required")
	}
	return nil
}

func validatePayment(account *models.TTenantPayAccount, order *models.TRechargeOrder) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	if order == nil || order.OrderNo == "" || !order.OrderAmount.IsPositive() {
		return fmt.Errorf("valid wechat recharge order is required")
	}
	return nil
}

func validatePayout(account *models.TTenantPayAccount, order *models.TWithdrawOrder) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	if order == nil || order.OrderNo == "" || !order.Amount.IsPositive() {
		return fmt.Errorf("valid wechat withdraw order is required")
	}
	return nil
}

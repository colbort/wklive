package alipay

import (
	"context"
	"errors"
	"fmt"

	"wklive/services/payment/internal/provider"
	"wklive/services/payment/models"
)

var ErrNotImplemented = errors.New("alipay adapter is not implemented")

var _ provider.Adapter = (*Adapter)(nil)

// Adapter is the Alipay provider implementation template. Replace each
// ErrNotImplemented return with the corresponding Alipay SDK call, signature
// verification and response mapping before registering it in ServiceContext.
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
	if len(request.Body) == 0 && len(request.Form) == 0 && len(request.Query) == 0 {
		return nil, fmt.Errorf("alipay payment notification is empty")
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
	if request.RefundNo == "" || request.Amount <= 0 {
		return nil, fmt.Errorf("alipay refund number and amount are required")
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
	if len(request.Body) == 0 && len(request.Form) == 0 && len(request.Query) == 0 {
		return nil, fmt.Errorf("alipay payout notification is empty")
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
		return fmt.Errorf("alipay account is required")
	}
	if account.Enabled != 1 {
		return fmt.Errorf("alipay account is disabled")
	}
	if !account.AppId.Valid || account.AppId.String == "" {
		return fmt.Errorf("alipay app_id is required")
	}
	if account.CredentialRef == "" && (!account.PrivateKeyCipher.Valid || account.PrivateKeyCipher.String == "") {
		return fmt.Errorf("alipay credential_ref or private key is required")
	}
	return nil
}

func validatePayment(account *models.TTenantPayAccount, order *models.TRechargeOrder) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	if order == nil || order.OrderNo == "" || order.OrderAmount <= 0 {
		return fmt.Errorf("valid alipay recharge order is required")
	}
	return nil
}

func validatePayout(account *models.TTenantPayAccount, order *models.TWithdrawOrder) error {
	if err := validateAccount(account); err != nil {
		return err
	}
	if order == nil || order.OrderNo == "" || order.Amount <= 0 {
		return fmt.Errorf("valid alipay withdraw order is required")
	}
	return nil
}

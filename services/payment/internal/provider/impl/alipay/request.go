package alipay

import (
	"context"

	"wklive/services/payment/internal/provider"
	"wklive/services/payment/models"
)

func requestPayment(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder) (*provider.CreatePaymentResult, error) {
	return nil, ErrNotImplemented
}

func requestPaymentQuery(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder) (*provider.PaymentQueryResult, error) {
	return nil, ErrNotImplemented
}

func requestRefund(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder, provider.RefundRequest) (*provider.RefundResult, error) {
	return nil, ErrNotImplemented
}

func requestPayout(context.Context, *models.TTenantPayAccount, *models.TWithdrawOrder) (*provider.CreatePayoutResult, error) {
	return nil, ErrNotImplemented
}

func requestPayoutQuery(context.Context, *models.TTenantPayAccount, *models.TWithdrawOrder) (*provider.PayoutQueryResult, error) {
	return nil, ErrNotImplemented
}

func handlePaymentNotify(context.Context, *models.TTenantPayAccount, provider.NotifyRequest) (*provider.PaymentNotifyResult, error) {
	return nil, ErrNotImplemented
}

func handlePayoutNotify(context.Context, *models.TTenantPayAccount, provider.NotifyRequest) (*provider.PayoutNotifyResult, error) {
	return nil, ErrNotImplemented
}

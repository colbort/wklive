package provider

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/shopspring/decimal"

	"wklive/services/payment/models"
)

type CreatePaymentResult struct {
	ThirdOrderNo string
	PayURL       string
	QRContent    string
	ExpireTime   int64
	Status       int64
	RawRequest   string
	RawResponse  string
}

type PaymentQueryResult struct {
	ThirdOrderNo string
	ThirdTradeNo string
	Status       int64
	PayAmount    decimal.Decimal
	Currency     string
	PaidTime     int64
	RawResponse  string
}

type RefundRequest struct {
	RefundNo string
	Amount   decimal.Decimal
	Currency string
	Reason   string
}

type RefundResult struct {
	RefundNo      string
	ThirdRefundNo string
	Status        int64
	Amount        decimal.Decimal
	Currency      string
	FinishedTime  int64
	RawRequest    string
	RawResponse   string
}

type CreatePayoutResult struct {
	ThirdOrderNo string
	ThirdTradeNo string
	Status       int64
	Amount       decimal.Decimal
	Currency     string
	FinishedTime int64
	RawRequest   string
	RawResponse  string
}

type PayoutQueryResult struct {
	ThirdOrderNo string
	ThirdTradeNo string
	Status       int64
	Amount       decimal.Decimal
	Currency     string
	FinishedTime int64
	RawResponse  string
}

type NotifyRequest struct {
	Headers map[string][]string
	Query   map[string][]string
	Form    map[string][]string
	Body    []byte
}

type PaymentNotifyResult struct {
	NotifyID     string
	OrderNo      string
	ThirdOrderNo string
	ThirdTradeNo string
	Status       int64
	PayAmount    decimal.Decimal
	Currency     string
	PaidTime     int64
	RawBody      string
}

type PayoutNotifyResult struct {
	NotifyID     string
	OrderNo      string
	ThirdOrderNo string
	ThirdTradeNo string
	Status       int64
	Amount       decimal.Decimal
	Currency     string
	FinishedTime int64
	RawBody      string
}

type Adapter interface {
	CreatePayment(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder) (*CreatePaymentResult, error)
	PaymentNotify(context.Context, *models.TTenantPayAccount, NotifyRequest) (*PaymentNotifyResult, error)
	QueryPayment(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder) (*PaymentQueryResult, error)
	RefundPayment(context.Context, *models.TTenantPayAccount, *models.TRechargeOrder, RefundRequest) (*RefundResult, error)
	CreatePayout(context.Context, *models.TTenantPayAccount, *models.TWithdrawOrder) (*CreatePayoutResult, error)
	PayoutNotify(context.Context, *models.TTenantPayAccount, NotifyRequest) (*PayoutNotifyResult, error)
	QueryPayout(context.Context, *models.TTenantPayAccount, *models.TWithdrawOrder) (*PayoutQueryResult, error)
}

type Registry struct {
	mu       sync.RWMutex
	adapters map[string]Adapter
}

func NewRegistry() *Registry {
	return &Registry{adapters: make(map[string]Adapter)}
}

func (r *Registry) Register(platformCode string, adapter Adapter) error {
	code := strings.ToLower(strings.TrimSpace(platformCode))
	if code == "" || adapter == nil {
		return fmt.Errorf("payment platform code and adapter are required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.adapters[code]; exists {
		return fmt.Errorf("payment adapter already registered for %s", code)
	}
	r.adapters[code] = adapter
	return nil
}

func (r *Registry) Get(platformCode string) (Adapter, error) {
	if r == nil {
		return nil, fmt.Errorf("payment adapter registry is nil")
	}
	code := strings.ToLower(strings.TrimSpace(platformCode))
	r.mu.RLock()
	adapter := r.adapters[code]
	r.mu.RUnlock()
	if adapter == nil {
		return nil, fmt.Errorf("payment adapter is not configured for platform %s", code)
	}
	return adapter, nil
}

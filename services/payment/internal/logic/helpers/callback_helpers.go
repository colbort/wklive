package helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/proto/payment"
	"wklive/services/payment/internal/provider"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"
)

func ResolveNotifyAdapter(
	ctx context.Context,
	svcCtx *svc.ServiceContext,
	in *payment.ThirdPartyNotifyReq,
) (*models.TTenantPayAccount, provider.Adapter, error) {
	if in == nil || in.TenantId <= 0 || strings.TrimSpace(in.PlatformCode) == "" || strings.TrimSpace(in.AccountCode) == "" {
		return nil, nil, fmt.Errorf("notification platform, tenant and account are required")
	}
	account, err := svcCtx.TenantPayAccountModel.FindOneByTenantIdAccountCode(ctx, in.TenantId, in.AccountCode)
	if err != nil {
		return nil, nil, err
	}
	platform, err := svcCtx.PayPlatformModel.FindOne(ctx, account.PlatformId)
	if err != nil {
		return nil, nil, err
	}
	if !strings.EqualFold(platform.PlatformCode, in.PlatformCode) {
		return nil, nil, fmt.Errorf("notification platform does not match account")
	}
	adapter, err := svcCtx.PaymentAdapters.Get(platform.PlatformCode)
	return account, adapter, err
}

func ToNotifyRequest(in *payment.ThirdPartyNotifyReq) provider.NotifyRequest {
	headers := make(map[string][]string, len(in.Headers))
	for key, value := range in.Headers {
		headers[key] = []string{value}
	}
	query := make(map[string][]string, len(in.Query))
	for key, value := range in.Query {
		query[key] = []string{value}
	}
	form := make(map[string][]string, len(in.Form))
	for key, value := range in.Form {
		form[key] = []string{value}
	}
	return provider.NotifyRequest{Headers: headers, Query: query, Form: form, Body: in.Body}
}

func NotifyResponse(platformCode string, success bool) *payment.ThirdPartyNotifyResp {
	switch strings.ToLower(strings.TrimSpace(platformCode)) {
	case string(provider.PaymentCodeAlipay):
		body := []byte("failure")
		if success {
			body = []byte("success")
		}
		return &payment.ThirdPartyNotifyResp{HttpStatus: 200, ContentType: "text/plain; charset=utf-8", Body: body}
	case string(provider.PaymentCodeWechat):
		body := []byte(`{"code":"FAIL","message":"处理失败"}`)
		if success {
			body = []byte(`{"code":"SUCCESS","message":"成功"}`)
		}
		return &payment.ThirdPartyNotifyResp{HttpStatus: 200, ContentType: "application/json; charset=utf-8", Body: body}
	default:
		return &payment.ThirdPartyNotifyResp{HttpStatus: 400, ContentType: "text/plain; charset=utf-8", Body: []byte("unsupported platform")}
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, models.ErrNotFound)
}

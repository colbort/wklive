// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package trade

import (
	"context"
	"strings"

	"wklive/app-api/internal/logicutil"
	"wklive/app-api/internal/svc"
	"wklive/app-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PlaceOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPlaceOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PlaceOrderLogic {
	return &PlaceOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PlaceOrderLogic) PlaceOrder(req *types.PlaceOrderReq) (resp *types.PlaceOrderResp, err error) {
	nextReq := normalizePlaceOrderReq(req)
	return logicutil.Proxy[types.PlaceOrderResp](l.ctx, &nextReq, l.svcCtx.TradeCli.PlaceOrder)
}

func normalizePlaceOrderReq(req *types.PlaceOrderReq) types.PlaceOrderReq {
	nextReq := *req
	nextReq.Amount = scaleTradeMinorText(req.Amount)
	return nextReq
}

func scaleTradeMinorText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}

	sign := ""
	switch text[:1] {
	case "-":
		sign = "-"
		text = text[1:]
	case "+":
		text = text[1:]
	}
	if text == "" {
		return value
	}

	parts := strings.Split(text, ".")
	if len(parts) > 2 {
		return value
	}
	integerPart := parts[0]
	fractionPart := ""
	if len(parts) == 2 {
		fractionPart = parts[1]
	}
	if integerPart == "" {
		integerPart = "0"
	}
	if !isDigits(integerPart) || !isDigits(fractionPart) {
		return value
	}

	scaledInteger := integerPart
	scaledFraction := ""
	if len(fractionPart) >= 2 {
		scaledInteger += fractionPart[:2]
		scaledFraction = fractionPart[2:]
	} else {
		scaledInteger += fractionPart + strings.Repeat("0", 2-len(fractionPart))
	}

	scaledInteger = strings.TrimLeft(scaledInteger, "0")
	if scaledInteger == "" {
		scaledInteger = "0"
	}
	scaledFraction = strings.TrimRight(scaledFraction, "0")
	if scaledInteger == "0" && scaledFraction == "" {
		sign = ""
	}
	if scaledFraction != "" {
		return sign + scaledInteger + "." + scaledFraction
	}
	return sign + scaledInteger
}

func isDigits(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

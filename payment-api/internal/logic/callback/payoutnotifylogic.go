// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package callback

import (
	"context"

	"wklive/payment-api/internal/svc"
	"wklive/payment-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayoutNotifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayoutNotifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayoutNotifyLogic {
	return &PayoutNotifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayoutNotifyLogic) PayoutNotify(req *types.NotifyReq) (*types.ThirdPartyNotifyResp, error) {
	resp, err := l.svcCtx.PaymentCli.PayoutNotify(l.ctx, &payment.ThirdPartyNotifyReq{
		PlatformCode: req.PlatformCode,
		TenantId:     req.TenantId,
		AccountCode:  req.AccountCode,
		Headers:      req.Headers,
		Form:         req.Form,
		Query:        req.Query,
		Body:         req.Body,
	})

	if err != nil {
		return nil, err
	}
	return &types.ThirdPartyNotifyResp{
		HttpStatus:  resp.HttpStatus,
		ContentType: resp.ContentType,
		Body:        resp.Body,
	}, nil
}

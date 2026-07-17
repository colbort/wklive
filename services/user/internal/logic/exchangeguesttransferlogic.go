package logic

import (
	"context"
	"encoding/json"
	"strings"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/user"
	"wklive/services/user/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExchangeGuestTransferLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewExchangeGuestTransferLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExchangeGuestTransferLogic {
	return &ExchangeGuestTransferLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

// ExchangeGuestTransfer atomically consumes a guest transfer code and issues a token for the original guest.
func (l *ExchangeGuestTransferLogic) ExchangeGuestTransfer(in *user.ExchangeGuestTransferReq) (*user.ExchangeGuestTransferResp, error) {
	code := strings.TrimSpace(in.GetCode())
	currentOrigin, ok := normalizeTransferOrigin(in.GetCurrentOrigin())
	if code == "" || !ok {
		return exchangeGuestTransferError(l.ctx), nil
	}

	raw, err := l.svcCtx.Redis.GetDelCtx(l.ctx, guestTransferRedisKey(code))
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return exchangeGuestTransferError(l.ctx), nil
	}

	var payload guestTransferPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.TargetOrigin != currentOrigin || payload.UserID <= 0 {
		return exchangeGuestTransferError(l.ctx), nil
	}

	expand, err := json.Marshal(map[string]any{"tenantId": payload.TenantID})
	if err != nil {
		return nil, err
	}
	token, err := buildTokenInfo(l.svcCtx.Config.Jwt.AccessSecret, 24*3600, payload.UserID, payload.Username, string(expand))
	if err != nil {
		return nil, err
	}

	return &user.ExchangeGuestTransferResp{
		Base: helper.OkResp(),
		Data: &user.ExchangeGuestTransferData{
			Token: token.AccessToken, DeviceId: payload.DeviceID, UserId: payload.DeviceID,
		},
	}, nil
}

func exchangeGuestTransferError(ctx context.Context) *user.ExchangeGuestTransferResp {
	return &user.ExchangeGuestTransferResp{Base: helper.ErrResp(i18n.TokenExpiredOrInvalid, i18n.Translate(i18n.TokenExpiredOrInvalid, ctx))}
}

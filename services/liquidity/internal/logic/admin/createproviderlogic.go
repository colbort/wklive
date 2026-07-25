package adminlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateProviderLogic {
	return &CreateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateProviderLogic) CreateProvider(in *liquidity.CreateProviderReq) (*liquidity.ProviderResp, error) {
	code, name := strings.TrimSpace(in.ProviderCode), strings.TrimSpace(in.ProviderName)
	if code == "" || name == "" {
		return nil, fmt.Errorf("provider_code and provider_name are required")
	}
	if in.ProviderType != liquidity.ProviderType_PROVIDER_TYPE_INTERNAL &&
		in.ProviderType != liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL {
		return nil, fmt.Errorf("invalid provider_type")
	}
	if in.ProviderType == liquidity.ProviderType_PROVIDER_TYPE_INTERNAL && in.TradeUserId <= 0 {
		return nil, fmt.Errorf("trade_user_id is required for internal provider")
	}
	if in.ProviderType == liquidity.ProviderType_PROVIDER_TYPE_INTERNAL {
		if err := validateInternalTradingUser(l.ctx, l.svcCtx, in.TradeUserId); err != nil {
			return nil, err
		}
		if _, err := l.svcCtx.ProviderModel.FindOneByProviderTypeTradeUserId(
			l.ctx,
			int64(liquidity.ProviderType_PROVIDER_TYPE_INTERNAL),
			in.TradeUserId,
		); err == nil {
			return nil, fmt.Errorf("trade_user_id is already bound to an internal provider")
		} else if err != models.ErrNotFound {
			return nil, err
		}
	}
	if in.ProviderType == liquidity.ProviderType_PROVIDER_TYPE_EXTERNAL &&
		(strings.TrimSpace(in.VenueCode) == "" || strings.TrimSpace(in.CredentialRef) == "") {
		return nil, fmt.Errorf("venue_code and credential_ref are required for external provider")
	}
	if _, err := l.svcCtx.ProviderModel.FindOneByProviderCode(l.ctx, code); err == nil {
		return nil, fmt.Errorf("provider_code already exists")
	} else if err != models.ErrNotFound {
		return nil, err
	}
	now := time.Now().UnixMilli()
	status := in.Status
	if status == liquidity.ProviderStatus_PROVIDER_STATUS_UNKNOWN {
		status = liquidity.ProviderStatus_PROVIDER_STATUS_DISABLED
	}
	row := &models.TLiquidityProvider{
		ProviderCode: code, ProviderName: name,
		ProviderType: int64(in.ProviderType), TradeUserId: in.TradeUserId,
		VenueCode: strings.TrimSpace(in.VenueCode), Environment: int64(in.Environment),
		CredentialRef: strings.TrimSpace(in.CredentialRef), AccountRef: strings.TrimSpace(in.AccountRef),
		RateLimitPerSecond: int64(in.RateLimitPerSecond), Status: int64(status),
		LastHealthStatus: int64(liquidity.HealthStatus_HEALTH_STATUS_UNKNOWN),
		Version:          1, Remark: strings.TrimSpace(in.Remark), CreateTimes: now, UpdateTimes: now,
	}
	result, err := l.svcCtx.ProviderModel.Insert(l.ctx, row)
	if err != nil {
		return nil, err
	}
	row.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &liquidity.ProviderResp{Base: helper.OkResp(), Data: helpers.ProviderToProto(row)}, nil
}

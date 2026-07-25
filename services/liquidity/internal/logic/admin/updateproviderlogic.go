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

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProviderLogic {
	return &UpdateProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateProviderLogic) UpdateProvider(in *liquidity.UpdateProviderReq) (*liquidity.ProviderResp, error) {
	row, err := l.svcCtx.ProviderModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if row.Version != in.Version {
		return nil, fmt.Errorf("provider version conflict")
	}
	if strings.TrimSpace(in.ProviderName) == "" {
		return nil, fmt.Errorf("provider_name is required")
	}
	if row.ProviderType == int64(liquidity.ProviderType_PROVIDER_TYPE_INTERNAL) {
		if err := validateInternalTradingUser(l.ctx, l.svcCtx, in.TradeUserId); err != nil {
			return nil, err
		}
	}
	row.ProviderName = strings.TrimSpace(in.ProviderName)
	row.TradeUserId = in.TradeUserId
	row.VenueCode = strings.TrimSpace(in.VenueCode)
	row.Environment = int64(in.Environment)
	if strings.TrimSpace(in.CredentialRef) != "" {
		row.CredentialRef = strings.TrimSpace(in.CredentialRef)
	}
	row.AccountRef = strings.TrimSpace(in.AccountRef)
	row.RateLimitPerSecond = int64(in.RateLimitPerSecond)
	row.Remark = strings.TrimSpace(in.Remark)
	row.Version++
	row.UpdateTimes = time.Now().UnixMilli()
	if err := l.svcCtx.ProviderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.ProviderResp{Base: helper.OkResp(), Data: helpers.ProviderToProto(row)}, nil
}

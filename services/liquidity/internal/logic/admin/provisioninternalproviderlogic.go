package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/proto/trade"
	"wklive/proto/user"
	"wklive/services/liquidity/internal/logic/helpers"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProvisionInternalProviderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProvisionInternalProviderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProvisionInternalProviderLogic {
	return &ProvisionInternalProviderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ProvisionInternalProviderLogic) ProvisionInternalProvider(in *liquidity.ProvisionInternalProviderReq) (*liquidity.ProviderResp, error) {
	if in == nil || in.SymbolId <= 0 {
		return nil, fmt.Errorf("symbol_id is required")
	}
	code, name := strings.TrimSpace(in.ProviderCode), strings.TrimSpace(in.ProviderName)
	if code == "" || name == "" {
		return nil, fmt.Errorf("provider_code and provider_name are required")
	}
	baseAmount, err := parseNumber("base_amount", in.BaseAmount)
	if err != nil || !baseAmount.IsPositive() {
		return nil, fmt.Errorf("base_amount must be positive")
	}
	quoteAmount, err := parseNumber("quote_amount", in.QuoteAmount)
	if err != nil || !quoteAmount.IsPositive() {
		return nil, fmt.Errorf("quote_amount must be positive")
	}
	if existing, err := l.svcCtx.ProviderModel.FindOneByProviderCode(l.ctx, code); err == nil {
		return &liquidity.ProviderResp{Base: existingProviderOK(), Data: helpers.ProviderToProto(existing)}, nil
	} else if err != models.ErrNotFound {
		return nil, err
	}
	symbolResp, err := l.svcCtx.TradeClient.GetSymbolDetail(l.ctx, &trade.GetSymbolDetailReq{SymbolId: in.SymbolId})
	if err != nil {
		return nil, err
	}
	symbol := symbolResp.GetData().GetSymbol()
	if symbolResp.GetBase().GetCode() != 200 || symbol == nil {
		return nil, fmt.Errorf("trade symbol unavailable: %s", symbolResp.GetBase().GetMsg())
	}
	if symbol.TenantId <= 0 {
		return nil, fmt.Errorf("symbol tenant is unavailable")
	}
	if symbol.ProductType != common.ProductType_PRODUCT_TYPE_SPOT {
		return nil, fmt.Errorf("account provisioning currently supports spot symbols")
	}
	walletType := walletTypeForLiquidity(symbol.ProductType, symbol.CategoryType)
	account, err := l.svcCtx.UserClient.CreateInternalTradingUser(l.ctx, &user.CreateInternalTradingUserReq{
		TenantId: symbol.TenantId, AccountKey: "liquidity:" + code,
		Nickname: name, Source: "LIQUIDITY_ADMIN", Remark: strings.TrimSpace(in.Remark),
	})
	if err != nil {
		return nil, err
	}
	if account.GetBase().GetCode() != 200 || account.GetData() == nil {
		return nil, fmt.Errorf("create market-maker user failed: %s", account.GetBase().GetMsg())
	}
	for _, balance := range []struct {
		coin, amount string
	}{
		{symbol.BaseAsset, in.BaseAmount},
		{symbol.QuoteAsset, in.QuoteAmount},
	} {
		bizNo := fmt.Sprintf("LQ-PROVISION-%s-%s", code, strings.ToUpper(balance.coin))
		resp, addErr := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
			TenantId: symbol.TenantId, UserId: account.Data.UserId,
			WalletType: walletType,
			Coin:       strings.ToUpper(balance.coin), Amount: balance.amount,
			BizType: asset.BizType_BIZ_TYPE_SYSTEM, SceneType: asset.SceneType_SCENE_TYPE_SYSTEM_ADJUST,
			BizNo: bizNo, Remark: strings.TrimSpace(in.Remark),
		})
		if addErr != nil {
			return nil, addErr
		}
		if resp.GetBase().GetCode() != 200 {
			return nil, fmt.Errorf("initialize %s failed: %s", balance.coin, resp.GetBase().GetMsg())
		}
	}
	return NewCreateProviderLogic(l.ctx, l.svcCtx).CreateProvider(&liquidity.CreateProviderReq{
		ProviderCode: code, ProviderName: name,
		ProviderType: liquidity.ProviderType_PROVIDER_TYPE_INTERNAL,
		TradeUserId:  account.Data.UserId,
		Environment:  liquidity.ProviderEnvironment_PROVIDER_ENVIRONMENT_PRODUCTION,
		Status:       liquidity.ProviderStatus_PROVIDER_STATUS_DISABLED,
		Remark:       strings.TrimSpace(in.Remark),
	})
}

func existingProviderOK() *common.RespBase {
	return &common.RespBase{Code: 200, Msg: "OK"}
}

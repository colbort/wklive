package adminlogic

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"wklive/common/helper"
	"wklive/proto/common"
	"wklive/proto/liquidity"
	"wklive/proto/trade"
	"wklive/proto/user"
	"wklive/services/liquidity/internal/svc"
	"wklive/services/liquidity/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/mr"
)

type GetConfigOptionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetConfigOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetConfigOptionsLogic {
	return &GetConfigOptionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetConfigOptionsLogic) GetConfigOptions(in *liquidity.GetConfigOptionsReq) (*liquidity.GetConfigOptionsResp, error) {
	if in == nil {
		in = &liquidity.GetConfigOptionsReq{}
	}
	keyword := strings.TrimSpace(in.Keyword)
	var (
		symbolResp   *trade.GetSymbolListAdminResp
		userResp     *user.ListInternalTradingUsersResp
		providerRows []*models.TLiquidityProvider
		unavailable  []string
		warnings     []string
		mu           sync.Mutex
	)
	recordUnavailable := func(section string, err error) {
		l.Errorf("config options section unavailable: section=%s err=%v", section, err)
		mu.Lock()
		defer mu.Unlock()
		unavailable = append(unavailable, section)
		warnings = append(warnings, fmt.Sprintf("%s options are temporarily unavailable", section))
	}
	_ = mr.Finish(
		func() error {
			resp, err := l.svcCtx.TradeClient.GetSymbolList(l.ctx, &trade.GetSymbolListAdminReq{
				Status:  trade.SymbolStatus_SYMBOL_STATUS_ENABLED,
				Keyword: keyword,
				Page:    &common.PageReq{Limit: 100},
			})
			if err != nil {
				recordUnavailable("symbols", err)
				return nil
			}
			if resp.GetBase().GetCode() != 200 {
				recordUnavailable("symbols", fmt.Errorf("%s", resp.GetBase().GetMsg()))
				return nil
			}
			symbolResp = resp
			return nil
		},
		func() error {
			resp, err := l.svcCtx.UserClient.ListInternalTradingUsers(l.ctx, &user.ListInternalTradingUsersReq{
				Keyword: keyword,
				Page:    &common.PageReq{Limit: 100},
			})
			if err != nil {
				recordUnavailable("tradingUsers", err)
				return nil
			}
			if resp.GetBase().GetCode() != 200 {
				recordUnavailable("tradingUsers", fmt.Errorf("%s", resp.GetBase().GetMsg()))
				return nil
			}
			userResp = resp
			return nil
		},
		func() error {
			rows, _, err := l.svcCtx.ProviderModel.FindPage(l.ctx, models.LiquidityProviderPageFilter{
				Keyword: keyword,
			}, 0, 100)
			if err != nil {
				recordUnavailable("providers", err)
				return nil
			}
			providerRows = rows
			return nil
		},
	)
	sort.Strings(unavailable)
	sort.Strings(warnings)
	if len(unavailable) == 3 {
		return nil, fmt.Errorf("all config option sections are unavailable")
	}

	resp := &liquidity.GetConfigOptionsResp{
		Base: helper.OkResp(), UnavailableSections: unavailable, Warnings: warnings,
	}
	for _, symbol := range symbolResp.GetData() {
		isSpot := symbol.ProductType == common.ProductType_PRODUCT_TYPE_SPOT &&
			symbol.ContractType == common.ContractType_CONTRACT_TYPE_NOT_APPLICABLE
		isSupportedContract := symbol.ProductType == common.ProductType_PRODUCT_TYPE_DERIVATIVE &&
			(symbol.ContractType == common.ContractType_CONTRACT_TYPE_PERPETUAL ||
				symbol.ContractType == common.ContractType_CONTRACT_TYPE_DELIVERY) &&
			(symbol.ContractValueType == trade.ContractValueType_CONTRACT_VALUE_TYPE_LINEAR ||
				symbol.ContractValueType == trade.ContractValueType_CONTRACT_VALUE_TYPE_INVERSE)
		if !isSpot && !isSupportedContract {
			continue
		}
		walletType := common.WalletType_WALLET_TYPE_CONTRACT
		if isSpot {
			walletType = common.WalletType_WALLET_TYPE_SPOT
		}
		resp.Symbols = append(resp.Symbols, &liquidity.ConfigSymbolOption{
			SymbolId: symbol.Id, Symbol: symbol.Symbol, DisplaySymbol: symbol.DisplaySymbol,
			ProductType: symbol.ProductType, ContractType: symbol.ContractType, WalletType: walletType,
			ContractValueType: int32(symbol.ContractValueType),
		})
	}
	for _, provider := range providerRows {
		resp.Providers = append(resp.Providers, &liquidity.ConfigProviderOption{
			ProviderId: provider.Id, ProviderCode: provider.ProviderCode,
			ProviderName: provider.ProviderName, ProviderType: liquidity.ProviderType(provider.ProviderType),
			TradeUserId: provider.TradeUserId, Status: liquidity.ProviderStatus(provider.Status),
		})
	}
	for _, account := range userResp.GetData() {
		resp.TradingUsers = append(resp.TradingUsers, &liquidity.ConfigTradingUserOption{
			TradeUserId: account.UserId, Username: account.Username,
		})
	}
	return resp, nil
}

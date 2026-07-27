package adminlogic

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"sync"
	"time"

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

type providerProvisionLock struct {
	mu   sync.Mutex
	refs int
}

var internalProviderProvisionLocks = struct {
	sync.Mutex
	locks map[string]*providerProvisionLock
}{locks: make(map[string]*providerProvisionLock)}

func lockInternalProviderProvision(code string) func() {
	internalProviderProvisionLocks.Lock()
	lock := internalProviderProvisionLocks.locks[code]
	if lock == nil {
		lock = &providerProvisionLock{}
		internalProviderProvisionLocks.locks[code] = lock
	}
	lock.refs++
	internalProviderProvisionLocks.Unlock()
	lock.mu.Lock()
	return func() {
		lock.mu.Unlock()
		internalProviderProvisionLocks.Lock()
		lock.refs--
		if lock.refs == 0 {
			delete(internalProviderProvisionLocks.locks, code)
		}
		internalProviderProvisionLocks.Unlock()
	}
}

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
	unlock := lockInternalProviderProvision(code)
	defer unlock()
	baseAmount, err := parseNumber("base_amount", in.BaseAmount)
	if err != nil || !baseAmount.IsPositive() {
		return nil, fmt.Errorf("base_amount must be positive")
	}
	quoteAmount, err := parseNumber("quote_amount", in.QuoteAmount)
	if err != nil || !quoteAmount.IsPositive() {
		return nil, fmt.Errorf("quote_amount must be positive")
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
	requestHash := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf(
		"%d|%d|%s|%s|%s|%s|%s|%s",
		in.SymbolId, symbol.TenantId, strings.ToUpper(code), name,
		strings.ToUpper(symbol.BaseAsset), baseAmount.String(),
		strings.ToUpper(symbol.QuoteAsset), quoteAmount.String(),
	))))
	provision, err := l.svcCtx.ProviderProvisionModel.Reserve(l.ctx, code, requestHash, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}
	updateProgress := func(userID, step int64, lastError string) error {
		return l.svcCtx.ProviderProvisionModel.UpdateProgress(
			l.ctx, code, userID, step, lastError, time.Now().UnixMilli(),
		)
	}
	fail := func(userID int64, err error) (*liquidity.ProviderResp, error) {
		if persistErr := updateProgress(userID, models.ProvisionStepFailed, err.Error()); persistErr != nil {
			return nil, fmt.Errorf("%w; persist provisioning failure: %v", err, persistErr)
		}
		return nil, err
	}
	if existing, findErr := l.svcCtx.ProviderModel.FindOneByProviderCode(l.ctx, code); findErr == nil {
		if provision.TradeUserId > 0 && existing.TradeUserId != provision.TradeUserId {
			return fail(provision.TradeUserId, fmt.Errorf("provider trade user does not match provisioning record"))
		}
		if err := l.setInternalTradingUserStatus(existing.TradeUserId, user.UserStatus_USER_STATUS_NORMAL, "liquidity provider provisioning completed"); err != nil {
			return fail(existing.TradeUserId, err)
		}
		if err := updateProgress(existing.TradeUserId, models.ProvisionStepCompleted, ""); err != nil {
			return nil, err
		}
		return existingInternalProviderResp(existing)
	} else if findErr != models.ErrNotFound {
		return fail(provision.TradeUserId, findErr)
	}
	account, err := l.svcCtx.UserClient.CreateInternalTradingUser(l.ctx, &user.CreateInternalTradingUserReq{
		TenantId: symbol.TenantId, AccountKey: "liquidity:" + code,
		Nickname: name, Source: "LIQUIDITY_ADMIN", Remark: strings.TrimSpace(in.Remark),
	})
	if err != nil {
		return fail(provision.TradeUserId, err)
	}
	if account.GetBase().GetCode() != 200 || account.GetData() == nil {
		return fail(provision.TradeUserId, fmt.Errorf("create market-maker user failed: %s", account.GetBase().GetMsg()))
	}
	userID := account.Data.UserId
	if provision.TradeUserId > 0 && provision.TradeUserId != userID {
		return fail(provision.TradeUserId, fmt.Errorf("provisioning account does not match existing trade user"))
	}
	if err := l.setInternalTradingUserStatus(userID, user.UserStatus_USER_STATUS_DISABLED, "liquidity provider provisioning"); err != nil {
		return fail(userID, err)
	}
	if err := updateProgress(userID, models.ProvisionStepAccountCreated, ""); err != nil {
		return fail(userID, err)
	}
	for _, balance := range []struct {
		coin, amount string
	}{
		{symbol.BaseAsset, baseAmount.String()},
		{symbol.QuoteAsset, quoteAmount.String()},
	} {
		bizNo := fmt.Sprintf("LQ-PROVISION-%s-%s", code, strings.ToUpper(balance.coin))
		resp, addErr := l.svcCtx.AssetClient.AddAvailable(l.ctx, &asset.AddAvailableReq{
			TenantId: symbol.TenantId, UserId: account.Data.UserId,
			WalletType: common.WalletType_WALLET_TYPE_SPOT,
			Coin:       strings.ToUpper(balance.coin), Amount: balance.amount,
			BizType: asset.BizType_BIZ_TYPE_SYSTEM, SceneType: asset.SceneType_SCENE_TYPE_SYSTEM_ADJUST,
			BizNo: bizNo, Remark: strings.TrimSpace(in.Remark),
		})
		if addErr != nil {
			return fail(userID, addErr)
		}
		if resp.GetBase().GetCode() != 200 {
			return fail(userID, fmt.Errorf("initialize %s failed: %s", balance.coin, resp.GetBase().GetMsg()))
		}
	}
	if err := updateProgress(userID, models.ProvisionStepFunded, ""); err != nil {
		return fail(userID, err)
	}
	resp, err := NewCreateProviderLogic(l.ctx, l.svcCtx).CreateProvider(&liquidity.CreateProviderReq{
		ProviderCode: code, ProviderName: name,
		ProviderType: liquidity.ProviderType_PROVIDER_TYPE_INTERNAL,
		TradeUserId:  account.Data.UserId,
		Environment:  liquidity.ProviderEnvironment_PROVIDER_ENVIRONMENT_PRODUCTION,
		Status:       liquidity.ProviderStatus_PROVIDER_STATUS_DISABLED,
		Remark:       strings.TrimSpace(in.Remark),
	})
	if err != nil {
		// 跨实例并发时唯一索引只允许一个创建成功；读取胜者结果即可安全收敛。
		existing, findErr := l.svcCtx.ProviderModel.FindOneByProviderCode(l.ctx, code)
		if findErr != nil || existing.TradeUserId != userID {
			return fail(userID, err)
		}
		resp, err = existingInternalProviderResp(existing)
		if err != nil {
			return fail(userID, err)
		}
	}
	if err := l.setInternalTradingUserStatus(userID, user.UserStatus_USER_STATUS_NORMAL, "liquidity provider provisioning completed"); err != nil {
		return fail(userID, err)
	}
	if err := updateProgress(userID, models.ProvisionStepCompleted, ""); err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *ProvisionInternalProviderLogic) setInternalTradingUserStatus(userID int64, status user.UserStatus, remark string) error {
	resp, err := l.svcCtx.UserClient.SetInternalTradingUserStatus(l.ctx, &user.SetInternalTradingUserStatusReq{
		UserId: userID, Status: status, Remark: remark,
	})
	if err != nil {
		return err
	}
	if resp.GetBase().GetCode() != 200 {
		return fmt.Errorf("set internal trading user status failed: %s", resp.GetBase().GetMsg())
	}
	return nil
}

func existingProviderOK() *common.RespBase {
	return &common.RespBase{Code: 200, Msg: "OK"}
}

func existingInternalProviderResp(existing *models.TLiquidityProvider) (*liquidity.ProviderResp, error) {
	if existing.ProviderType != int64(liquidity.ProviderType_PROVIDER_TYPE_INTERNAL) ||
		existing.TradeUserId <= 0 {
		return nil, fmt.Errorf("provider_code is occupied by a non-internal provider")
	}
	return &liquidity.ProviderResp{Base: existingProviderOK(), Data: helpers.ProviderToProto(existing)}, nil
}

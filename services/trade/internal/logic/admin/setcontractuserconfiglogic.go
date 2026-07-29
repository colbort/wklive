package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/trade"
	"wklive/services/trade/internal/authz"
	"wklive/services/trade/internal/logic/helpers"
	"wklive/services/trade/internal/svc"
	"wklive/services/trade/internal/validation"
	"wklive/services/trade/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetContractUserConfigLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetContractUserConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetContractUserConfigLogic {
	return &SetContractUserConfigLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置用户合约偏好配置
func (l *SetContractUserConfigLogic) SetContractUserConfig(in *trade.SetContractUserConfigReq) (*trade.CommonResp, error) {
	crossEnabled := l.svcCtx.Config.CrossMarginTrading.Enabled &&
		l.svcCtx.Config.AutomaticLiquidation.Enabled
	if err := validateContractUserConfigInput(in, nil, crossEnabled); err != nil {
		return nil, err
	}
	if base, err := authz.AdminTenantWriteScopeResp(l.ctx, in.TenantId, i18n.BusinessDataNotFound); err != nil {
		return nil, err
	} else if base != nil {
		return &trade.CommonResp{Base: base}, nil
	}
	globalModeLockKey := fmt.Sprintf(
		"trade:contract-mode:%d:%d:0",
		in.TenantId, in.UserId,
	)
	globalModeLockValue := fmt.Sprintf("%d:%d:admin-global", in.UserId, time.Now().UnixNano())
	releaseGlobalModeLock, err := helpers.AcquireRenewingTaskLock(
		l.ctx, l.svcCtx.Redis, globalModeLockKey, globalModeLockValue,
	)
	if err != nil {
		return nil, errors.New("contract account mode is being updated; retry")
	}
	defer func() {
		if releaseErr := releaseGlobalModeLock(); releaseErr != nil {
			l.Errorf("release global contract mode lock failed, key=%s err=%v", globalModeLockKey, releaseErr)
		}
	}()
	if in.SymbolId > 0 {
		modeLockKey := fmt.Sprintf(
			"trade:contract-mode:%d:%d:%d",
			in.TenantId, in.UserId, in.SymbolId,
		)
		modeLockValue := fmt.Sprintf("%d:%d:admin-symbol", in.UserId, time.Now().UnixNano())
		releaseModeLock, lockErr := helpers.AcquireRenewingTaskLock(
			l.ctx, l.svcCtx.Redis, modeLockKey, modeLockValue,
		)
		if lockErr != nil {
			return nil, errors.New("contract account mode is being updated; retry")
		}
		defer func() {
			if releaseErr := releaseModeLock(); releaseErr != nil {
				l.Errorf("release contract mode lock failed, key=%s err=%v", modeLockKey, releaseErr)
			}
		}()
	}
	if in.SymbolId > 0 {
		contract, err := l.svcCtx.TradeSymbolContractModel.FindOneByTenantIdSymbolId(l.ctx, in.TenantId, in.SymbolId)
		if err != nil {
			return nil, err
		}
		if err = validateContractUserConfigInput(in, contract, crossEnabled); err != nil {
			return nil, err
		}
	}
	item, err := l.svcCtx.ContractUserConfigModel.FindOneByTenantIdUserIdSymbolId(l.ctx, in.TenantId, in.UserId, in.SymbolId)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}
	if item == nil || contractUserModeChanged(item, in) {
		blocked, blockErr := l.hasOpenContractRiskUnit(in.TenantId, in.UserId, in.SymbolId)
		if blockErr != nil {
			return nil, blockErr
		}
		if blocked {
			return nil, errors.New("position or active order must be closed before changing contract position or margin mode")
		}
	}
	now := utils.NowMillis()
	if item == nil {
		item = &models.TContractUserConfig{TenantId: in.TenantId, UserId: in.UserId, SymbolId: in.SymbolId, CreateTimes: now}
	}
	item.PositionMode, item.MarginMode, item.DefaultLeverage = int64(in.PositionMode), int64(in.MarginMode), in.DefaultLeverage
	item.UpdateTimes = now
	if item.Id == 0 {
		_, err = l.svcCtx.ContractUserConfigModel.Insert(l.ctx, item)
	} else {
		err = l.svcCtx.ContractUserConfigModel.Update(l.ctx, item)
	}
	if err != nil {
		return nil, err
	}
	return &trade.CommonResp{Base: helper.OkResp()}, nil
}

func validateContractUserConfigInput(in *trade.SetContractUserConfigReq, contract *models.TTradeSymbolContract, crossEnabled bool) error {
	if in == nil || in.TenantId <= 0 || in.UserId <= 0 || in.SymbolId < 0 {
		return errors.New("invalid contract user configuration identity")
	}
	if in.PositionMode != trade.PositionMode_POSITION_MODE_ONE_WAY &&
		in.PositionMode != trade.PositionMode_POSITION_MODE_HEDGE {
		return errors.New("invalid contract position mode")
	}
	if in.MarginMode != trade.MarginMode_MARGIN_MODE_ISOLATED &&
		in.MarginMode != trade.MarginMode_MARGIN_MODE_CROSS {
		return errors.New("invalid contract margin mode")
	}
	if in.DefaultLeverage <= 0 {
		return errors.New("default leverage must be positive")
	}
	if in.MarginMode == trade.MarginMode_MARGIN_MODE_CROSS && !crossEnabled {
		return errors.New("cross margin is temporarily unavailable: account-level liquidation is not enabled")
	}
	if contract != nil && !validation.ContractSupportsMarginMode(contract, in.MarginMode) {
		return errors.New("contract does not support the selected margin mode")
	}
	return nil
}

func contractUserModeChanged(item *models.TContractUserConfig, in *trade.SetContractUserConfigReq) bool {
	return item != nil && in != nil &&
		(item.PositionMode != int64(in.PositionMode) || item.MarginMode != int64(in.MarginMode))
}

func (l *SetContractUserConfigLogic) hasOpenContractRiskUnit(tenantID, userID, symbolID int64) (bool, error) {
	positions, err := l.svcCtx.ContractPositionModel.CountOpenRiskUnit(
		l.ctx, tenantID, userID, symbolID,
	)
	if err != nil {
		return false, err
	}
	if positions > 0 {
		return true, nil
	}
	orders, err := l.svcCtx.TradeOrderModel.CountOpenContractRiskUnit(
		l.ctx, tenantID, userID, symbolID,
	)
	if err != nil {
		return false, err
	}
	return orders > 0, nil
}

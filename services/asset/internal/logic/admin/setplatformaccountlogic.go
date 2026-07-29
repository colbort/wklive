package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/logic/helpers"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetPlatformAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPlatformAccountLogic {
	return &SetPlatformAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 设置平台账户（保险基金等，不属于任何用户）
func (l *SetPlatformAccountLogic) SetPlatformAccount(in *asset.SetPlatformAccountReq) (*asset.PlatformAccountResp, error) {
	typeName, coin := normalizePlatformAccount(in.GetAccountType(), in.GetCoin())
	if in.GetTenantId() <= 0 || !isConfigurablePlatformAccountType(typeName) || coin == "" {
		return nil, fmt.Errorf("invalid platform account")
	}
	status := int64(in.GetStatus())
	if status == 0 {
		status = 1
	}
	if status != 1 && status != 2 {
		return nil, fmt.Errorf("invalid platform account status")
	}
	m := models.NewTAssetPlatformAccountModel(l.svcCtx.DB, l.svcCtx.Config.CacheRedis)
	row, err := m.FindOneByTenantIdAccountTypeCoin(l.ctx, in.GetTenantId(), typeName, coin)
	now := utils.NowMillis()
	if errors.Is(err, models.ErrNotFound) {
		row = &models.TAssetPlatformAccount{TenantId: in.GetTenantId(), AccountType: typeName, Coin: coin, Status: status, Version: 1, CreateTimes: now, UpdateTimes: now}
		res, e := m.Insert(l.ctx, row)
		if e != nil {
			return nil, e
		}
		row.Id, _ = res.LastInsertId()
	} else if err != nil {
		return nil, err
	} else {
		if row.Version != in.GetVersion() {
			return nil, fmt.Errorf("platform account version conflict")
		}
		row.Status, row.Version, row.UpdateTimes = status, row.Version+1, now
		if err = m.Update(l.ctx, row); err != nil {
			return nil, err
		}
	}
	return &asset.PlatformAccountResp{Base: helper.OkResp(), Data: platformAccountProto(row)}, nil
}

func normalizePlatformAccount(t, coin string) (string, string) {
	return strings.ToUpper(strings.TrimSpace(t)), strings.ToUpper(strings.TrimSpace(coin))
}
func platformAccountProto(v *models.TAssetPlatformAccount) *asset.PlatformAccount {
	return &asset.PlatformAccount{Id: v.Id, TenantId: v.TenantId, AccountType: v.AccountType, Coin: v.Coin, AvailableAmount: v.AvailableAmount.String(), FrozenAmount: v.FrozenAmount.String(), Status: helpers.ToAssetStatus(v.Status), Version: v.Version, CreateTimes: v.CreateTimes, UpdateTimes: v.UpdateTimes}
}

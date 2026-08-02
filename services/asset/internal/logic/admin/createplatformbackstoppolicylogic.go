package adminlogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type CreatePlatformBackstopPolicyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePlatformBackstopPolicyLogic {
	return &CreatePlatformBackstopPolicyLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CreatePlatformBackstopPolicyLogic) CreatePlatformBackstopPolicy(in *asset.CreatePlatformBackstopPolicyReq) (*asset.PlatformBackstopPolicyResp, error) {
	now := utils.NowMillis()
	coin, requestNo, reason, perRequest, daily, floor, err := validateBackstopPolicyDraft(in, now)
	if err != nil {
		return backstopPolicyParamResp(l.ctx), nil
	}
	operatorID, permission, err := platformBackstopPolicyAdmin(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if permission != nil {
		return backstopPolicyPermissionResp(permission), nil
	}
	var created *models.TAssetBackstopPolicy
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		policies := models.NewTAssetBackstopPolicyModel(conn, l.svcCtx.Config.CacheRedis)
		account, lockErr := accounts.FindOneForUpdate(ctx, in.GetTenantId(), optionBackstopAccountType, coin)
		if lockErr != nil {
			return fmt.Errorf("active OPTION_BACKSTOP account is required: %w", lockErr)
		}
		if account.Status != 1 {
			return fmt.Errorf("active OPTION_BACKSTOP account is required")
		}
		existing, findErr := policies.FindOneByRequestNoForUpdate(ctx, in.GetTenantId(), requestNo)
		if findErr == nil {
			if !sameBackstopPolicyDraft(existing, in, coin, reason, perRequest, daily, floor) {
				return fmt.Errorf("backstop policy request_no parameters changed")
			}
			created = existing
			return nil
		}
		if !errors.Is(findErr, models.ErrNotFound) {
			return findErr
		}
		version, versionErr := policies.NextVersion(ctx, in.GetTenantId(), coin)
		if versionErr != nil {
			return versionErr
		}
		created = &models.TAssetBackstopPolicy{
			TenantId: in.GetTenantId(), Coin: coin, RequestNo: requestNo, Version: version,
			Mode: int64(in.GetMode()), PerRequestLimit: perRequest, DailyLimit: daily,
			BalanceFloor: floor, EffectiveFrom: in.GetEffectiveFrom(), EffectiveUntil: in.GetEffectiveUntil(),
			Status: 1, Reason: reason, EvidenceRef: strings.TrimSpace(in.GetEvidenceRef()),
			CreatedBy: operatorID, CreateTimes: now, UpdateTimes: now,
		}
		result, insertErr := policies.Insert(ctx, created)
		if insertErr != nil {
			return insertErr
		}
		created.Id, insertErr = result.LastInsertId()
		return insertErr
	})
	if err != nil {
		return nil, err
	}
	return &asset.PlatformBackstopPolicyResp{Base: helper.OkResp(), Data: platformBackstopPolicyProto(created)}, nil
}

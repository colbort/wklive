package adminlogic

import (
	"context"
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

type ReviewPlatformBackstopPolicyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewReviewPlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReviewPlatformBackstopPolicyLogic {
	return &ReviewPlatformBackstopPolicyLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ReviewPlatformBackstopPolicyLogic) ReviewPlatformBackstopPolicy(in *asset.ReviewPlatformBackstopPolicyReq) (*asset.PlatformBackstopPolicyResp, error) {
	if in == nil || in.GetTenantId() <= 0 || in.GetPolicyId() <= 0 ||
		strings.TrimSpace(in.GetReason()) == "" || len(strings.TrimSpace(in.GetReason())) > 255 {
		return backstopPolicyParamResp(l.ctx), nil
	}
	operatorID, permission, err := platformBackstopPolicyAdmin(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if permission != nil {
		return backstopPolicyPermissionResp(permission), nil
	}
	var reviewed *models.TAssetBackstopPolicy
	err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		conn := sqlx.NewSqlConnFromSession(session)
		policies := models.NewTAssetBackstopPolicyModel(conn, l.svcCtx.Config.CacheRedis)
		accounts := models.NewTAssetPlatformAccountModel(conn, l.svcCtx.Config.CacheRedis)
		candidate, findErr := policies.FindOne(ctx, in.GetPolicyId())
		if findErr != nil {
			return findErr
		}
		if candidate.TenantId != in.GetTenantId() {
			return fmt.Errorf("invalid backstop policy review scope")
		}
		if _, lockErr := accounts.FindOneForUpdate(ctx, candidate.TenantId, optionBackstopAccountType, candidate.Coin); lockErr != nil {
			return fmt.Errorf("active OPTION_BACKSTOP account is required: %w", lockErr)
		}
		row, findErr := policies.FindOneForUpdate(ctx, in.GetPolicyId())
		if findErr != nil {
			return findErr
		}
		if row.TenantId != in.GetTenantId() || row.Status != 1 || row.CreatedBy == operatorID {
			return fmt.Errorf("invalid or non-four-eyes backstop policy review")
		}
		if in.GetApprove() && row.EffectiveUntil <= utils.NowMillis() {
			return fmt.Errorf("expired backstop policy cannot be approved")
		}
		row.Status = 3
		if in.GetApprove() {
			row.Status = 2
		}
		row.ReviewedBy = operatorID
		row.ReviewReason = strings.TrimSpace(in.GetReason())
		row.UpdateTimes = utils.NowMillis()
		if updateErr := policies.Update(ctx, row); updateErr != nil {
			return updateErr
		}
		reviewed = row
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &asset.PlatformBackstopPolicyResp{Base: helper.OkResp(), Data: platformBackstopPolicyProto(reviewed)}, nil
}

package adminlogic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformBackstopPolicyLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPlatformBackstopPolicyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformBackstopPolicyLogic {
	return &GetPlatformBackstopPolicyLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *GetPlatformBackstopPolicyLogic) GetPlatformBackstopPolicy(in *asset.GetPlatformBackstopPolicyReq) (*asset.PlatformBackstopPolicyResp, error) {
	if in == nil || in.GetTenantId() <= 0 || in.GetPolicyId() <= 0 {
		return backstopPolicyParamResp(l.ctx), nil
	}
	_, permission, err := platformBackstopPolicyAdmin(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if permission != nil {
		return backstopPolicyPermissionResp(permission), nil
	}
	model := models.NewTAssetBackstopPolicyModel(l.svcCtx.DB, l.svcCtx.Config.CacheRedis)
	row, err := model.FindOne(l.ctx, in.GetPolicyId())
	if err != nil {
		return nil, err
	}
	if row.TenantId != in.GetTenantId() {
		return backstopPolicyParamResp(l.ctx), nil
	}
	return &asset.PlatformBackstopPolicyResp{Base: helper.OkResp(), Data: platformBackstopPolicyProto(row)}, nil
}

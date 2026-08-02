package adminlogic

import (
	"context"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPlatformBackstopPoliciesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPlatformBackstopPoliciesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPlatformBackstopPoliciesLogic {
	return &ListPlatformBackstopPoliciesLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *ListPlatformBackstopPoliciesLogic) ListPlatformBackstopPolicies(in *asset.ListPlatformBackstopPoliciesReq) (*asset.ListPlatformBackstopPoliciesResp, error) {
	if in == nil || in.GetTenantId() <= 0 {
		return &asset.ListPlatformBackstopPoliciesResp{Base: backstopPolicyParamResp(l.ctx).Base}, nil
	}
	if in.GetStatus() < asset.PlatformBackstopPolicyStatus_PLATFORM_BACKSTOP_POLICY_STATUS_UNKNOWN ||
		in.GetStatus() > asset.PlatformBackstopPolicyStatus_PLATFORM_BACKSTOP_POLICY_STATUS_REJECTED {
		return &asset.ListPlatformBackstopPoliciesResp{Base: backstopPolicyParamResp(l.ctx).Base}, nil
	}
	_, permission, err := platformBackstopPolicyAdmin(l.ctx, in.GetTenantId())
	if err != nil {
		return nil, err
	}
	if permission != nil {
		return &asset.ListPlatformBackstopPoliciesResp{Base: permission}, nil
	}
	cursor, limit := pageutil.Input(in.GetPage())
	model := models.NewTAssetBackstopPolicyModel(l.svcCtx.DB, l.svcCtx.Config.CacheRedis)
	rows, total, err := model.FindPage(l.ctx, in.GetTenantId(), strings.ToUpper(strings.TrimSpace(in.GetCoin())), int64(in.GetStatus()), cursor, limit)
	if err != nil {
		return nil, err
	}
	lastID := int64(0)
	data := make([]*asset.PlatformBackstopPolicy, 0, len(rows))
	for _, row := range rows {
		lastID = row.Id
		data = append(data, platformBackstopPolicyProto(row))
	}
	return &asset.ListPlatformBackstopPoliciesResp{
		Base: pageutil.Base(cursor, limit, len(rows), total, lastID),
		Data: data, Total: total,
	}, nil
}

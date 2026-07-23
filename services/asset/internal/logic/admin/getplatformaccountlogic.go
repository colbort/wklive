package adminlogic

import (
	"context"
	"fmt"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPlatformAccountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPlatformAccountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPlatformAccountLogic {
	return &GetPlatformAccountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询平台账户
func (l *GetPlatformAccountLogic) GetPlatformAccount(in *asset.GetPlatformAccountReq) (*asset.PlatformAccountResp, error) {
	typeName, coin := normalizePlatformAccount(in.GetAccountType(), in.GetCoin())
	if in.GetTenantId() <= 0 || typeName != insuranceFundAccountType || coin == "" {
		return nil, fmt.Errorf("invalid platform account")
	}
	row, err := models.NewTAssetPlatformAccountModel(l.svcCtx.DB, l.svcCtx.Config.CacheRedis).FindOneByTenantIdAccountTypeCoin(l.ctx, in.GetTenantId(), typeName, coin)
	if err != nil {
		return nil, err
	}
	return &asset.PlatformAccountResp{Base: helper.OkResp(), Data: platformAccountProto(row)}, nil
}

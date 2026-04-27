package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListAssetCoinConfigsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListAssetCoinConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListAssetCoinConfigsLogic {
	return &ListAssetCoinConfigsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询APP资产操作页币种配置
func (l *ListAssetCoinConfigsLogic) ListAssetCoinConfigs(in *asset.ListAssetCoinConfigsReq) (*asset.ListAssetCoinConfigsResp, error) {
	list, err := l.svcCtx.AssetCoinConfigModel.FindVisibleByOperation(l.ctx, in.TenantId, int64(in.WalletType), int64(in.OperationType), int64(in.CoinType))
	if err != nil {
		return nil, err
	}

	resp := &asset.ListAssetCoinConfigsResp{Base: helper.OkResp()}
	for _, item := range list {
		resp.Data = append(resp.Data, toAssetCoinConfigProto(item))
	}

	return resp, nil
}

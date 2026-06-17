package logic

import (
	"context"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/proto/common"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetCoinConfigsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetCoinConfigsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetCoinConfigsLogic {
	return &PageAssetCoinConfigsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询APP资产操作币种显示配置
func (l *PageAssetCoinConfigsLogic) PageAssetCoinConfigs(in *asset.PageAssetCoinConfigsReq) (*asset.PageAssetCoinConfigsResp, error) {
	page := in.Page
	if page == nil {
		page = &common.PageReq{}
	}

	list, total, err := l.svcCtx.AssetCoinConfigModel.FindPage(l.ctx, models.AssetCoinConfigPageFilter{
		TenantId:        in.TenantId,
		WalletType:      int64(in.WalletType),
		Coin:            in.Coin,
		Symbol:          in.Symbol,
		CoinType:        int64(in.CoinType),
		ChainCode:       int64(in.ChainCode),
		AppVisible:      int64(in.AppVisible),
		RechargeEnabled: int64(in.RechargeEnabled),
		WithdrawEnabled: int64(in.WithdrawEnabled),
		TransferEnabled: int64(in.TransferEnabled),
		Enabled:         int64(in.Enabled),
	}, page.Cursor, page.Limit)
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(list) > 0 {
		lastID = list[len(list)-1].Id
	}

	resp := &asset.PageAssetCoinConfigsResp{Base: pageutil.Base(page.Cursor, page.Limit, len(list), total, lastID)}
	for _, item := range list {
		resp.Data = append(resp.Data, toAssetCoinConfigProto(item))
	}

	return resp, nil
}

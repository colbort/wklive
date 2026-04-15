package asset

import (
	"context"

	"wklive/admin-api/internal/logicutil"
	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/asset"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAssetOptionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAssetOptionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAssetOptionsLogic {
	return &GetAssetOptionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAssetOptionsLogic) GetAssetOptions() (resp *types.GetAssetOptionsResp, err error) {
	return &types.GetAssetOptionsResp{
		RespBase: types.RespBase{Code: 200, Msg: "success"},
		Data: []types.OptionsGroup{
			logicutil.EnumGroup("walletType", "钱包类型", asset.WalletType_WALLET_TYPE_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("assetStatus", "资产状态", asset.AssetStatus_ASSET_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("freezeStatus", "冻结状态", asset.FreezeStatus_FREEZE_STATUS_UNKNOWN.Descriptor()),
			logicutil.EnumGroup("lockStatus", "锁定状态", asset.LockStatus_LOCK_STATUS_UNKNOWN.Descriptor()),
		},
	}, nil
}

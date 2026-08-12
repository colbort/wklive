package adminlogic

import (
	"context"
	"wklive/services/asset/internal/logic/helpers"

	"wklive/common/pageutil"
	"wklive/proto/asset"
	"wklive/services/asset/internal/svc"
	"wklive/services/asset/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type PageAssetFreezesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPageAssetFreezesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PageAssetFreezesLogic {
	return &PageAssetFreezesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 分页查询冻结明细
func (l *PageAssetFreezesLogic) PageAssetFreezes(in *asset.PageAssetFreezesReq) (*asset.PageAssetFreezesResp, error) {
	freezes, total, err := l.svcCtx.AssetFreezeModel.FindPage(l.ctx, models.AssetFreezePageFilter{
		TenantId:   in.TenantId,
		UserId:     in.UserId,
		WalletType: int64(in.WalletType),
		Coin:       in.Coin,
		BizType:    helpers.AssetBizType(in.BizType),
		BizNo:      in.BizNo,
		Status:     int64(in.Status),
	}, in.Page.Cursor, in.Page.Limit, pageutil.Count(in.Page))
	if err != nil {
		return nil, err
	}

	lastID := int64(0)
	if len(freezes) > 0 {
		lastID = freezes[len(freezes)-1].Id
	}

	resp := &asset.PageAssetFreezesResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(freezes), total, lastID)}

	for _, item := range freezes {
		resp.Data = append(resp.Data, helpers.ToAssetFreezeProto(item))
	}
	return resp, nil
}

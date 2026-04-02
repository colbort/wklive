// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package itick

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/itick"

	"github.com/zeromicro/go-zero/core/logx"
)

type SyncCategoryProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSyncCategoryProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SyncCategoryProductsLogic {
	return &SyncCategoryProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SyncCategoryProductsLogic) SyncCategoryProducts(req *types.SyncCategoryProductsReq) (resp *types.SyncCategoryProductsResp, err error) {
	reuslt, err := l.svcCtx.ItickCli.SyncCategoryProducts(l.ctx, &itick.SyncCategoryProductsReq{
		CategoryId: req.CategoryId,
	})
	if err != nil {
		return nil, err
	}
	return &types.SyncCategoryProductsResp{
		RespBase: types.RespBase{
			Code: reuslt.Base.Code,
			Msg:  reuslt.Base.Msg,
		},
		TaskNo: reuslt.TaskNo,
	}, nil
}

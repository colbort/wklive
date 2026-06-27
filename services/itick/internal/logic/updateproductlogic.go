package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/common/i18n"
	cutils "wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateProductLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateProductLogic {
	return &UpdateProductLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新产品仅允许更新名称、状态、排序、图标和备注，市场、品种、代码不允许修改
func (l *UpdateProductLogic) UpdateProduct(in *itick.UpdateProductReq) (*itick.AdminCommonResp, error) {
	item, err := l.svcCtx.ItickProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return &itick.AdminCommonResp{
			Base: helper.ErrResp(i18n.ProductNotFound, i18n.Translate(i18n.ProductNotFound, l.ctx)),
		}, nil
	}

	if in.Name != "" {
		item.Name = in.Name
	}
	if in.DisplayName != "" {
		item.DisplayName = in.DisplayName
	}
	if in.BaseCoin != "" {
		item.BaseCoin = in.BaseCoin
	}
	if in.QuoteCoin != "" {
		item.QuoteCoin = in.QuoteCoin
	}
	if in.Enabled != 0 {
		item.Enabled = int64(in.Enabled)
	}
	if in.AppVisible != 0 {
		item.AppVisible = int64(in.AppVisible)
	}
	if in.Sort != 0 {
		item.Sort = in.Sort
	}
	if in.Icon != "" {
		item.Icon = in.Icon
	}
	if in.Remark != "" {
		item.Remark = in.Remark
	}
	if in.SyncPriority != itick.SyncKlinePriority_SYNC_KLINE_PRIORITY_UNKNOWN {
		item.SyncPriority = int64(in.SyncPriority)
	}
	item.UpdateTimes = cutils.NowMillis()

	if err := l.svcCtx.ItickProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}

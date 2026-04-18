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
			Base: helper.GetErrResp(1, i18n.Translate(i18n.ProductNotFound, l.ctx)),
		}, nil
	}

	item.Name = in.Name
	item.DisplayName = in.DisplayName
	item.BaseCoin = in.BaseCoin
	item.QuoteCoin = in.QuoteCoin
	item.Enabled = in.Enabled
	item.AppVisible = in.AppVisible
	item.Sort = in.Sort
	item.Icon = in.Icon
	item.Remark = in.Remark
	item.UpdateTimes = cutils.NowMillis()

	if err := l.svcCtx.ItickProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}

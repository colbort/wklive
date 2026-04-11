package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/common/utils"
	"wklive/proto/staking"
	"wklive/services/staking/internal/svc"
	"wklive/services/staking/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ProductChangeStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewProductChangeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ProductChangeStatusLogic {
	return &ProductChangeStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 修改质押产品状态
func (l *ProductChangeStatusLogic) ProductChangeStatus(in *staking.AdminProductChangeStatusReq) (*staking.AdminProductChangeStatusResp, error) {
	item, err := l.svcCtx.StakeProductModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &staking.AdminProductChangeStatusResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if item.TenantId != in.TenantId {
		return &staking.AdminProductChangeStatusResp{Page: helper.GetErrResp(404, i18n.Translate(i18n.ProductNotFound, l.ctx))}, nil
	}

	item.Status = int64(in.Status)
	item.UpdateUserId = in.OperatorUid
	item.UpdateTimes = utils.NowMillis()
	if err := l.svcCtx.StakeProductModel.Update(l.ctx, item); err != nil {
		return nil, err
	}

	return &staking.AdminProductChangeStatusResp{Page: helper.OkResp(), Data: true}, nil
}

package adminlogic

import (
	"context"
	"errors"
	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/option"
	"wklive/services/option/internal/logic/helpers"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBillLogic {
	return &GetBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个资金流水详情
func (l *GetBillLogic) GetBill(in *option.GetBillReq) (*option.GetBillResp, error) {
	item, err := l.svcCtx.OptionBillModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetBillResp{Base: helper.ErrResp(i18n.BillNotFound, i18n.Translate(i18n.BillNotFound, l.ctx))}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && item.TenantId != in.TenantId {
		return &option.GetBillResp{Base: helper.ErrResp(i18n.BillNotFound, i18n.Translate(i18n.BillNotFound, l.ctx))}, nil
	}

	return &option.GetBillResp{Base: helper.OkResp(), Data: helpers.ToBillProto(item)}, nil
}

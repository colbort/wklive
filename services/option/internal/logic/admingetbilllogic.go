package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/proto/option"
	"wklive/services/option/internal/svc"
	"wklive/services/option/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminGetBillLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminGetBillLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminGetBillLogic {
	return &AdminGetBillLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 获取单个资金流水详情
func (l *AdminGetBillLogic) AdminGetBill(in *option.GetBillReq) (*option.GetBillResp, error) {
	item, err := l.svcCtx.OptionBillModel.FindOne(l.ctx, in.Id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return &option.GetBillResp{Base: helper.GetErrResp(404, "资金流水不存在")}, nil
		}
		return nil, err
	}
	if in.TenantId != 0 && item.TenantId != in.TenantId {
		return &option.GetBillResp{Base: helper.GetErrResp(404, "资金流水不存在")}, nil
	}

	return &option.GetBillResp{Base: helper.OkResp(), Data: toBillProto(item)}, nil
}

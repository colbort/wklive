// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayProductLogic {
	return &UpdatePayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePayProductLogic) UpdatePayProduct(req *types.UpdatePayProductReq) (resp *types.RespBase, err error) {
	result, err := l.svcCtx.PaymentCli.UpdatePayProduct(l.ctx, &payment.UpdatePayProductReq{
		Id:          req.Id,
		ProductName: req.ProductName,
		SceneType:   payment.SceneType(req.SceneType),
		Currency:    req.Currency,
		Status:      payment.CommonStatus(req.Status),
		Remark:      req.Remark,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.RespBase{
		Code: result.Base.Code,
		Msg:  result.Base.Msg,
	}
	return resp, nil
}

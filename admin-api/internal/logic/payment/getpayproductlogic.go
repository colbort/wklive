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

type GetPayProductLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPayProductLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPayProductLogic {
	return &GetPayProductLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPayProductLogic) GetPayProduct(req *types.GetPayProductReq) (resp *types.GetPayProductResp, err error) {
	result, err := l.svcCtx.PaymentCli.GetPayProduct(l.ctx, &payment.GetPayProductReq{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}

	resp = &types.GetPayProductResp{
		RespBase: types.RespBase{
			Code: result.Base.Code,
			Msg:  result.Base.Msg,
		},
		Data: types.PayProduct{
			Id:          result.Data.Id,
			PlatformId:  result.Data.PlatformId,
			ProductCode: result.Data.ProductCode,
			ProductName: result.Data.ProductName,
			SceneType:   int64(result.Data.SceneType),
			Currency:    result.Data.Currency,
			Status:      int64(result.Data.Status),
			Remark:      result.Data.Remark,
			CreateTime:  result.Data.CreateTime,
			UpdateTime:  result.Data.UpdateTime,
		},
	}
	return resp, nil
}

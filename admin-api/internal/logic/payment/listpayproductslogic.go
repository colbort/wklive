// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package payment

import (
	"context"

	"wklive/admin-api/internal/svc"
	"wklive/admin-api/internal/types"
	"wklive/proto/common"
	"wklive/proto/payment"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayProductsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayProductsLogic {
	return &ListPayProductsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPayProductsLogic) ListPayProducts(req *types.ListPayProductsReq) (resp *types.ListPayProductsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListPayProducts(l.ctx, &payment.ListPayProductsReq{
		Page: &common.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		PlatformId:  req.PlatformId,
		Keyword:     req.Keyword,
		ProductCode: req.ProductCode,
		Status:      payment.CommonStatus(req.Status),
		SceneType:   payment.SceneType(req.SceneType),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.PayProduct, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.PayProduct{
			Id:          item.Id,
			PlatformId:  item.PlatformId,
			ProductCode: item.ProductCode,
			ProductName: item.ProductName,
			SceneType:   int64(item.SceneType),
			Currency:    item.Currency,
			Status:      int64(item.Status),
			Remark:      item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		}
	}

	resp = &types.ListPayProductsResp{
		RespBase: types.RespBase{
			Code:       result.Base.Code,
			Msg:        result.Base.Msg,
			Total:      result.Base.Total,
			HasNext:    result.Base.HasNext,
			HasPrev:    result.Base.HasPrev,
			NextCursor: result.Base.NextCursor,
			PrevCursor: result.Base.PrevCursor,
		},
		Data: data,
	}
	return resp, nil
}

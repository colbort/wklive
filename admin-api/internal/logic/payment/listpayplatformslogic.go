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

type ListPayPlatformsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayPlatformsLogic {
	return &ListPayPlatformsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListPayPlatformsLogic) ListPayPlatforms(req *types.ListPayPlatformsReq) (resp *types.ListPayPlatformsResp, err error) {
	result, err := l.svcCtx.PaymentCli.ListPayPlatforms(l.ctx, &payment.ListPayPlatformsReq{
		Page: &payment.PageReq{
			Cursor: req.Cursor,
			Limit:  req.Limit,
		},
		Keyword:      req.Keyword,
		PlatformCode: req.PlatformCode,
		Status:       payment.CommonStatus(req.Status),
	})
	if err != nil {
		return nil, err
	}

	data := make([]types.PayPlatform, len(result.Data))
	for i, item := range result.Data {
		data[i] = types.PayPlatform{
			Id:           item.Id,
			PlatformCode: item.PlatformCode,
			PlatformName: item.PlatformName,
			PlatformType: int64(item.PlatformType),
			NotifyUrl:    item.NotifyUrl,
			ReturnUrl:    item.ReturnUrl,
			Icon:         item.Icon,
			Status:       int64(item.Status),
			Remark:       item.Remark,
			CreateTimes:  item.CreateTimes,
			UpdateTimes:  item.UpdateTimes,
		}
	}

	resp = &types.ListPayPlatformsResp{
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

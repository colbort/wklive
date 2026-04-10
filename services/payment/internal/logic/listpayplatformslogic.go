package logic

import (
	"context"

	"wklive/common/helper"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayPlatformsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayPlatformsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayPlatformsLogic {
	return &ListPayPlatformsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 平台列表
func (l *ListPayPlatformsLogic) ListPayPlatforms(in *payment.ListPayPlatformsReq) (*payment.ListPayPlatformsResp, error) {
	items, total, err := l.svcCtx.PayPlatformModel.FindPage(l.ctx, in.Keyword, in.PlatformCode, int64(in.PlatformType), int64(in.Status), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	prevCursor := in.Page.Cursor
	if prevCursor < 0 {
		prevCursor = 0
	}
	nextCursor := int64(0)
	if int64(len(items)) == in.Page.Limit {
		lastItem := items[len(items)-1]
		nextCursor = lastItem.Id
	}
	hasPrev := prevCursor > 0
	hasNext := int64(len(items)) == in.Page.Limit
	data := make([]*payment.PayPlatform, 0)
	for _, p := range items {
		data = append(data, &payment.PayPlatform{
			Id:           p.Id,
			PlatformCode: p.PlatformCode,
			PlatformName: p.PlatformName,
			PlatformType: payment.PlatformType(p.PlatformType),
			NotifyUrl:    p.NotifyUrl.String,
			ReturnUrl:    p.ReturnUrl.String,
			Icon:         p.Icon.String,
			Status:       payment.CommonStatus(p.Status),
			Remark:       p.Remark.String,
			CreateTimes:  p.CreateTimes,
			UpdateTimes:  p.UpdateTimes,
		})
	}

	return &payment.ListPayPlatformsResp{
		Base: helper.OkWithOthers(total, hasNext, hasPrev, nextCursor, prevCursor),
		Data: data,
	}, nil
}

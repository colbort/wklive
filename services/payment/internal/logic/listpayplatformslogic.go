package logic

import (
	"context"

	"wklive/common/pageutil"
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
	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}
	data := make([]*payment.PayPlatform, 0)
	for _, p := range items {
		data = append(data, toPayPlatformProto(p))
	}

	return &payment.ListPayPlatformsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}

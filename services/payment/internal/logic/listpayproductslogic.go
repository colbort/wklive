package logic

import (
	"context"
	"errors"

	"wklive/common/pageutil"
	"wklive/proto/payment"
	"wklive/services/payment/internal/svc"
	"wklive/services/payment/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListPayProductsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListPayProductsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListPayProductsLogic {
	return &ListPayProductsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 产品列表
func (l *ListPayProductsLogic) ListPayProducts(in *payment.ListPayProductsReq) (*payment.ListPayProductsResp, error) {
	items, total, err := l.svcCtx.PayProductModel.FindPage(l.ctx, in.PlatformId, in.Page.Cursor, in.Page.Limit)
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return nil, err
	}

	lastID := int64(0)
	if len(items) > 0 {
		lastID = items[len(items)-1].Id
	}

	data := make([]*payment.PayProduct, 0, len(items))
	for _, p := range items {
		data = append(data, toPayProductProto(p))
	}

	return &payment.ListPayProductsResp{
		Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(items), total, lastID),
		Data: data,
	}, nil
}

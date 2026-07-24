package adminlogic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"wklive/common/helper"
	"wklive/proto/liquidity"
	"wklive/services/liquidity/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CancelExternalOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelExternalOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelExternalOrderLogic {
	return &CancelExternalOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelExternalOrderLogic) CancelExternalOrder(in *liquidity.CancelExternalOrderReq) (*liquidity.CommonResp, error) {
	row, err := l.svcCtx.ExternalOrderModel.FindOne(l.ctx, in.OrderId)
	if err != nil {
		return nil, err
	}
	if row.Version != in.Version {
		return nil, fmt.Errorf("external order version conflict")
	}
	switch liquidity.ExternalOrderStatus(row.Status) {
	case liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_PENDING_SUBMIT:
		row.Status, row.FinishedAt = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELED), time.Now().UnixMilli()
	case liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_SUBMITTED,
		liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_PART_FILLED,
		liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_UNCERTAIN:
		row.Status = int64(liquidity.ExternalOrderStatus_EXTERNAL_ORDER_STATUS_CANCELING)
	default:
		return nil, fmt.Errorf("external order is already terminal")
	}
	row.LastErrorMsg = strings.TrimSpace(in.Reason)
	row.Version++
	row.UpdateTimes = time.Now().UnixMilli()
	if err := l.svcCtx.ExternalOrderModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}

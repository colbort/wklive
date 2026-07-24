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

type CancelHedgeTaskLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCancelHedgeTaskLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelHedgeTaskLogic {
	return &CancelHedgeTaskLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CancelHedgeTaskLogic) CancelHedgeTask(in *liquidity.CancelHedgeTaskReq) (*liquidity.CommonResp, error) {
	row, err := l.svcCtx.HedgeTaskModel.FindOne(l.ctx, in.HedgeTaskId)
	if err != nil {
		return nil, err
	}
	if row.Version != in.Version {
		return nil, fmt.Errorf("hedge task version conflict")
	}
	switch liquidity.HedgeStatus(row.Status) {
	case liquidity.HedgeStatus_HEDGE_STATUS_COMPLETED, liquidity.HedgeStatus_HEDGE_STATUS_CANCELED:
		return nil, fmt.Errorf("hedge task is already terminal")
	case liquidity.HedgeStatus_HEDGE_STATUS_EXECUTING:
		return nil, fmt.Errorf("executing hedge task cannot be canceled directly")
	}
	row.Status, row.LastErrorMsg = int64(liquidity.HedgeStatus_HEDGE_STATUS_CANCELED), strings.TrimSpace(in.Reason)
	row.Version++
	row.UpdateTimes = time.Now().UnixMilli()
	if err := l.svcCtx.HedgeTaskModel.Update(l.ctx, row); err != nil {
		return nil, err
	}
	return &liquidity.CommonResp{Base: helper.OkResp()}, nil
}

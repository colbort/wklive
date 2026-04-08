package logic

import (
	"context"

	"wklive/proto/trade"
	"wklive/services/trade/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordPositionHistoryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRecordPositionHistoryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordPositionHistoryLogic {
	return &RecordPositionHistoryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 记录持仓历史信息
func (l *RecordPositionHistoryLogic) RecordPositionHistory(in *trade.RecordPositionHistoryReq) (*trade.InternalCommonResp, error) {
	// todo: add your logic here and delete this line

	return &trade.InternalCommonResp{}, nil
}

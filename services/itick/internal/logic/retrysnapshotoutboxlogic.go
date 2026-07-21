package logic

import (
	"context"
	"errors"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type RetrySnapshotOutboxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRetrySnapshotOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetrySnapshotOutboxLogic {
	return &RetrySnapshotOutboxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RetrySnapshotOutboxLogic) RetrySnapshotOutbox(in *itick.RetrySnapshotOutboxReq) (*itick.AdminCommonResp, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("snapshot outbox id is required")
	}
	if err := l.svcCtx.SnapshotOutboxModel.RetryFailed(l.ctx, in.Id, utils.NowMillis()); err != nil {
		return nil, err
	}
	return &itick.AdminCommonResp{Base: helper.OkResp()}, nil
}

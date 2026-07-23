package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/pageutil"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSnapshotOutboxLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewListSnapshotOutboxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSnapshotOutboxLogic {
	return &ListSnapshotOutboxLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ListSnapshotOutboxLogic) ListSnapshotOutbox(in *itick.ListSnapshotOutboxReq) (*itick.ListSnapshotOutboxResp, error) {
	if in == nil || in.Page == nil || in.Status < 0 || in.Status > 5 {
		return nil, errors.New("invalid snapshot outbox query")
	}
	rows, count, err := l.svcCtx.SnapshotOutboxModel.FindPage(l.ctx, int64(in.Status), strings.TrimSpace(in.SnapshotId), in.Page.Cursor, in.Page.Limit)
	if err != nil {
		return nil, err
	}
	data := make([]*itick.SnapshotOutboxData, 0, len(rows))
	for _, row := range rows {
		data = append(data, &itick.SnapshotOutboxData{Id: row.Id, SnapshotId: row.SnapshotId, Status: int32(row.Status), RetryCount: row.RetryCount, NextRetryAt: row.NextRetryAt, LastErrorMsg: row.LastErrorMsg, CreateTimes: row.CreateTimes, UpdateTimes: row.UpdateTimes, RedisPublishedAt: row.RedisPublishedAt, OptionPublishedAt: row.OptionPublishedAt})
	}
	lastID := int64(0)
	if len(rows) > 0 {
		lastID = rows[len(rows)-1].Id
	}
	return &itick.ListSnapshotOutboxResp{Base: pageutil.Base(in.Page.Cursor, in.Page.Limit, len(rows), count, lastID), Data: data}, nil
}

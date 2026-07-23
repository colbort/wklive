package logic

import (
	"context"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuthoritativeSnapshotInternalLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuthoritativeSnapshotInternalLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthoritativeSnapshotInternalLogic {
	return &GetAuthoritativeSnapshotInternalLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Reads an authoritative snapshot from the permanent archive at or before
func (l *GetAuthoritativeSnapshotInternalLogic) GetAuthoritativeSnapshotInternal(in *itick.GetAuthoritativeSnapshotInternalReq) (*itick.GetAuthoritativeSnapshotInternalResp, error) {
	return NewGetAuthoritativeSnapshotLogic(l.ctx, l.svcCtx).GetAuthoritativeSnapshotInternal(in)
}

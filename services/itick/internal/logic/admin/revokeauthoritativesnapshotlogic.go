package adminlogic

import (
	"context"
	"errors"
	"strings"

	"wklive/common/helper"
	"wklive/common/utils"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type RevokeAuthoritativeSnapshotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRevokeAuthoritativeSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RevokeAuthoritativeSnapshotLogic {
	return &RevokeAuthoritativeSnapshotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RevokeAuthoritativeSnapshotLogic) RevokeAuthoritativeSnapshot(in *itick.RevokeAuthoritativeSnapshotReq) (*itick.CommonResp, error) {
	if in == nil || strings.TrimSpace(in.SnapshotId) == "" || strings.TrimSpace(in.Reason) == "" || in.SnapshotId == in.ReplacementSnapshotId {
		return nil, errors.New("invalid authoritative snapshot revocation")
	}
	original, err := l.svcCtx.AuthoritativeSnapshotModel.FindOneBySnapshotId(l.ctx, strings.TrimSpace(in.SnapshotId))
	if err != nil {
		return nil, err
	}
	if in.ReplacementSnapshotId != "" {
		replacement, findErr := l.svcCtx.AuthoritativeSnapshotModel.FindOneBySnapshotId(l.ctx, strings.TrimSpace(in.ReplacementSnapshotId))
		if findErr != nil {
			return nil, findErr
		}
		if replacement.Authority != original.Authority || replacement.SnapshotKind != original.SnapshotKind || replacement.CategoryCode != original.CategoryCode || replacement.Market != original.Market || replacement.Symbol != original.Symbol || replacement.SourceTimestamp != original.SourceTimestamp || replacement.Revision <= original.Revision {
			return nil, errors.New("replacement must be a higher revision of the same snapshot input")
		}
	}
	existing, findErr := l.svcCtx.SnapshotRevocationModel.FindOneBySnapshotId(l.ctx, original.SnapshotId)
	if findErr == nil {
		if existing.ReplacementSnapshotId != strings.TrimSpace(in.ReplacementSnapshotId) || existing.Reason != strings.TrimSpace(in.Reason) {
			return nil, errors.New("snapshot revocation already exists with different content")
		}
	} else if !errors.Is(findErr, models.ErrNotFound) {
		return nil, findErr
	} else if _, err = l.svcCtx.SnapshotRevocationModel.Insert(l.ctx, &models.TItickSnapshotRevocation{SnapshotId: original.SnapshotId, ReplacementSnapshotId: strings.TrimSpace(in.ReplacementSnapshotId), Reason: strings.TrimSpace(in.Reason), CreateTimes: utils.NowMillis()}); err != nil {
		return nil, err
	}
	if err = l.svcCtx.MarketDataCache.RevokeAuthoritativeSnapshot(l.ctx, original.SnapshotId, strings.TrimSpace(in.ReplacementSnapshotId), strings.TrimSpace(in.Reason)); err != nil {
		return nil, err
	}
	return &itick.CommonResp{Base: helper.OkResp()}, nil
}

package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	marketcache "wklive/common/market"
	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAuthoritativeSnapshotLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAuthoritativeSnapshotLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAuthoritativeSnapshotLogic {
	return &GetAuthoritativeSnapshotLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按业务时刻读取生产方永久归档的权威快照
func (l *GetAuthoritativeSnapshotLogic) GetAuthoritativeSnapshot(in *itick.GetAuthoritativeSnapshotReq) (*itick.GetAuthoritativeSnapshotResp, error) {
	if in == nil || in.TargetTime <= 0 || in.MaxLookbackMs <= 0 || strings.TrimSpace(in.Authority) == "" {
		return nil, fmt.Errorf("invalid authoritative snapshot query")
	}
	msg := marketcache.NormalizeClientMessage(marketcache.ClientMessage{Topic: marketcache.TopicQuote, CategoryCode: in.CategoryCode, Market: in.Market, Symbol: in.Symbol})
	if msg.CategoryCode == "" || msg.Market == "" || msg.Symbol == "" {
		return nil, fmt.Errorf("authoritative snapshot product is required")
	}
	row, err := l.svcCtx.AuthoritativeSnapshotModel.FindAtOrBefore(l.ctx, strings.ToLower(strings.TrimSpace(in.Authority)), msg.CategoryCode, msg.Market, msg.Symbol, in.TargetTime, in.TargetTime-in.MaxLookbackMs)
	if errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("authoritative snapshot unavailable at target time")
	}
	if err != nil {
		return nil, err
	}
	return &itick.GetAuthoritativeSnapshotResp{Base: helper.OkResp(), Data: &itick.AuthoritativeSnapshot{SnapshotId: row.SnapshotId, Authority: row.Authority, SnapshotKind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Price: row.Price.String(), SourceTimestamp: row.SourceTimestamp, SnapshotTimestamp: row.SnapshotTimestamp, Revision: row.Revision, FormulaVersion: row.FormulaVersion, RawPayload: row.RawPayload}}, nil
}

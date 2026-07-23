package internallogic

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"wklive/common/helper"
	cache "wklive/common/market"
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
	if in == nil {
		return nil, fmt.Errorf("invalid authoritative snapshot query")
	}
	data, err := l.getAuthoritativeSnapshot(in.Authority, in.CategoryCode, in.Market, in.Symbol, in.TargetTime, in.MaxLookbackMs, in.SnapshotKind)
	if err != nil {
		return nil, err
	}
	return &itick.GetAuthoritativeSnapshotResp{Base: helper.OkResp(), Data: data}, nil
}

func (l *GetAuthoritativeSnapshotLogic) getAuthoritativeSnapshot(authorityValue, categoryCode, market, symbol string, targetTime, maxLookbackMs int64, snapshotKind string) (*itick.AuthoritativeSnapshot, error) {
	if targetTime <= 0 || maxLookbackMs <= 0 || strings.TrimSpace(authorityValue) == "" {
		return nil, fmt.Errorf("invalid authoritative snapshot query")
	}
	msg := cache.NormalizeClientMessage(cache.ClientMessage{Topic: cache.TopicQuote, CategoryCode: categoryCode, Market: market, Symbol: symbol})
	if msg.CategoryCode == "" || msg.Market == "" || msg.Symbol == "" {
		return nil, fmt.Errorf("authoritative snapshot product is required")
	}
	kind := strings.ToUpper(strings.TrimSpace(snapshotKind))
	if kind == "" {
		return nil, fmt.Errorf("authoritative snapshot kind is required")
	}
	authorityName := strings.ToLower(strings.TrimSpace(authorityValue))
	authority, err := l.svcCtx.AuthorityRegistryModel.FindEnabled(l.ctx, authorityName)
	if errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("market authority is not registered or enabled")
	}
	if err != nil {
		return nil, err
	}
	if !authority.Allows(kind) {
		return nil, fmt.Errorf("market authority cannot provide %s", kind)
	}
	row, err := l.svcCtx.AuthoritativeSnapshotModel.FindAtOrBefore(l.ctx, authorityName, kind, msg.CategoryCode, msg.Market, msg.Symbol, targetTime, targetTime-maxLookbackMs)
	if errors.Is(err, models.ErrNotFound) {
		return nil, fmt.Errorf("authoritative snapshot unavailable at target time")
	}
	if err != nil {
		return nil, err
	}
	return &itick.AuthoritativeSnapshot{SnapshotId: row.SnapshotId, Authority: row.Authority, SnapshotKind: row.SnapshotKind, CategoryCode: row.CategoryCode, Market: row.Market, Symbol: row.Symbol, Price: row.Price.String(), SourceTimestamp: row.SourceTimestamp, SnapshotTimestamp: row.SnapshotTimestamp, Revision: row.Revision, FormulaVersion: row.FormulaVersion, RawPayload: row.RawPayload}, nil
}

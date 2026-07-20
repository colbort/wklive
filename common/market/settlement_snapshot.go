package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

func (b *MarketDataCache) LockPriceSnapshot(ctx context.Context, kind string, msg ClientMessage, maxAge time.Duration) (*SettlementSnapshot, error) {
	items, err := b.ReadMany(ctx, []ClientMessage{msg})
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, errors.New("settlement price unavailable")
	}
	q, ok := items[0].Payload.(*QuotePayload)
	if !ok || q == nil || q.LastPrice <= 0 || q.Ts <= 0 {
		return nil, errors.New("invalid settlement quote")
	}
	now := time.Now().UnixMilli()
	if q.Ts > now+1000 || now-q.Ts > maxAge.Milliseconds() {
		return nil, errors.New("stale settlement quote")
	}
	msg = NormalizeClientMessage(msg)
	s := &SettlementSnapshot{Kind: kind, CategoryCode: msg.CategoryCode, Market: msg.Market, Symbol: msg.Symbol, Price: fmt.Sprintf("%.18g", q.LastPrice), Source: msg.Market, SourceTimestamp: q.Ts, SnapshotTimestamp: now, Revision: q.Ts, Confirmed: true}
	s.SnapshotID = snapshotDigest(s)
	if err := b.PutSettlementSnapshot(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (b *MarketDataCache) PutSettlementSnapshot(ctx context.Context, s *SettlementSnapshot) error {
	if s == nil || !s.Confirmed || s.SourceTimestamp <= 0 {
		return errors.New("unconfirmed settlement snapshot")
	}
	if s.SnapshotTimestamp <= 0 {
		s.SnapshotTimestamp = time.Now().UnixMilli()
	}
	if s.Revision <= 0 {
		s.Revision = s.SourceTimestamp
	}
	if s.SnapshotID == "" {
		s.SnapshotID = snapshotDigest(s)
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("market:settlement:v1:%s", s.SnapshotID)
	return b.rdb.SetNX(ctx, key, raw, 30*24*time.Hour).Err()
}

func (b *MarketDataCache) GetSettlementSnapshot(ctx context.Context, id string) (*SettlementSnapshot, error) {
	raw, err := b.rdb.Get(ctx, fmt.Sprintf("market:settlement:v1:%s", id)).Bytes()
	if err != nil {
		return nil, err
	}
	var s SettlementSnapshot
	if err = json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func snapshotDigest(s *SettlementSnapshot) string {
	copy := *s
	copy.SnapshotID = ""
	// Reception time is audit metadata, not source identity. Re-reading the same
	// revision must resolve to the same immutable snapshot ID.
	copy.SnapshotTimestamp = 0
	raw, _ := json.Marshal(copy)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

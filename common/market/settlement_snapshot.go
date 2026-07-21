package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

const authoritativeSnapshotTTL = 365 * 24 * time.Hour

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

// PublishAuthoritativeQuote archives an immutable source-owned quote. Only the
// market-data producer may call this; consumers must query the archive and may
// not promote their local quote cache to an authoritative snapshot.
func (b *MarketDataCache) PublishAuthoritativeQuote(ctx context.Context, msg ClientMessage, q *QuotePayload) (*SettlementSnapshot, error) {
	s, err := BuildAuthoritativeQuoteSnapshot(msg, q)
	if err != nil {
		return nil, err
	}
	if err = b.PublishAuthoritativeSnapshot(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

func BuildAuthoritativeQuoteSnapshot(msg ClientMessage, q *QuotePayload) (*SettlementSnapshot, error) {
	msg = NormalizeClientMessage(msg)
	if q == nil || q.Ts <= 0 || strings.TrimSpace(q.LastPriceText) == "" || strings.TrimSpace(q.Authority) == "" {
		return nil, errors.New("authoritative quote metadata is incomplete")
	}
	priceText := strings.TrimSpace(q.LastPriceText)
	price, err := decimal.NewFromString(priceText)
	if err != nil || !price.IsPositive() {
		return nil, errors.New("authoritative quote price is invalid")
	}
	if err = validateArchiveDecimal(price); err != nil {
		return nil, err
	}
	s := &SettlementSnapshot{
		Kind:              "FINAL_QUOTE",
		CategoryCode:      msg.CategoryCode,
		Market:            msg.Market,
		Symbol:            msg.Symbol,
		Price:             priceText,
		Source:            msg.Market,
		SourceTimestamp:   q.Ts,
		SnapshotTimestamp: time.Now().UnixMilli(),
		Revision:          q.Ts,
		FormulaVersion:    "source-quote-v1",
		Authority:         strings.TrimSpace(q.Authority),
		Confirmed:         true,
	}
	s.SnapshotID = snapshotDigest(s)
	return s, nil
}

func validateArchiveDecimal(value decimal.Decimal) error {
	if value.Exponent() < -30 {
		return errors.New("authoritative quote exceeds 30 decimal places")
	}
	integerDigits := len(value.Abs().Truncate(0).StringFixed(0))
	if integerDigits > 35 {
		return errors.New("authoritative quote exceeds 35 integer digits")
	}
	return nil
}

func (b *MarketDataCache) PublishAuthoritativeSnapshot(ctx context.Context, s *SettlementSnapshot) error {
	if s == nil || s.SnapshotID == "" || !s.Confirmed || s.Authority == "" || s.SourceTimestamp <= 0 {
		return errors.New("invalid authoritative snapshot")
	}
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}
	dataKey := fmt.Sprintf("market:authoritative:v1:%s", s.SnapshotID)
	indexKey := authoritativeSnapshotKindIndex(ClientMessage{Topic: TopicQuote, CategoryCode: s.CategoryCode, Market: s.Market, Symbol: s.Symbol}, s.Authority, s.Kind)
	pipe := b.rdb.TxPipeline()
	pipe.SetNX(ctx, dataKey, raw, authoritativeSnapshotTTL)
	pipe.ZAdd(ctx, indexKey, redis.Z{Score: float64(s.SourceTimestamp), Member: s.SnapshotID})
	pipe.Expire(ctx, indexKey, authoritativeSnapshotTTL)
	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

// FindAuthoritativeQuoteAt returns the newest finalized source quote at or
// before targetTime, bounded by maxLookback.
func (b *MarketDataCache) FindAuthoritativeQuoteAt(ctx context.Context, msg ClientMessage, authority string, targetTime int64, maxLookback time.Duration) (*SettlementSnapshot, error) {
	return b.FindAuthoritativeSnapshotAt(ctx, msg, authority, "FINAL_QUOTE", targetTime, maxLookback)
}

// FindAuthoritativeSnapshotAt reads a purpose-specific immutable snapshot.
// Kind is part of the index so MARK, INDEX, FUNDING and DELIVERY cannot shadow
// one another when they share an authority and product.
func (b *MarketDataCache) FindAuthoritativeSnapshotAt(ctx context.Context, msg ClientMessage, authority, kind string, targetTime int64, maxLookback time.Duration) (*SettlementSnapshot, error) {
	msg = NormalizeClientMessage(msg)
	kind = strings.ToUpper(strings.TrimSpace(kind))
	if targetTime <= 0 || maxLookback <= 0 || strings.TrimSpace(authority) == "" || kind == "" {
		return nil, errors.New("invalid authoritative snapshot query")
	}
	ids, err := b.rdb.ZRevRangeByScore(ctx, authoritativeSnapshotKindIndex(msg, authority, kind), &redis.ZRangeBy{Max: fmt.Sprintf("%d", targetTime), Min: fmt.Sprintf("%d", targetTime-maxLookback.Milliseconds()), Offset: 0, Count: 100}).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("authoritative snapshot unavailable at target time")
	}
	var selected *SettlementSnapshot
	for _, id := range ids {
		raw, readErr := b.rdb.Get(ctx, fmt.Sprintf("market:authoritative:v1:%s", id)).Bytes()
		if readErr != nil {
			continue
		}
		var candidate SettlementSnapshot
		if json.Unmarshal(raw, &candidate) != nil || !candidate.Confirmed || !strings.EqualFold(candidate.Authority, strings.TrimSpace(authority)) || !strings.EqualFold(candidate.Kind, kind) || candidate.SourceTimestamp > targetTime {
			continue
		}
		if selected == nil || candidate.SourceTimestamp > selected.SourceTimestamp || (candidate.SourceTimestamp == selected.SourceTimestamp && candidate.Revision > selected.Revision) {
			copy := candidate
			selected = &copy
		}
	}
	if selected == nil {
		return nil, errors.New("valid authoritative snapshot unavailable at target time")
	}
	return selected, nil
}

func authoritativeSnapshotIndex(msg ClientMessage, authority string) string {
	return fmt.Sprintf("market:authoritative:v1:index:%s:%s:%s:%s", strings.ToLower(strings.TrimSpace(authority)), msg.CategoryCode, msg.Market, msg.Symbol)
}

func authoritativeSnapshotKindIndex(msg ClientMessage, authority, kind string) string {
	return fmt.Sprintf("market:authoritative:v2:index:%s:%s:%s:%s:%s", strings.ToLower(strings.TrimSpace(authority)), strings.ToUpper(strings.TrimSpace(kind)), msg.CategoryCode, msg.Market, msg.Symbol)
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

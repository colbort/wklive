package adminlogic

import (
	"context"
	"reflect"
	"testing"

	"wklive/proto/common"
	"wklive/proto/market"
	"wklive/services/market/internal/svc"
	"wklive/services/market/models"
)

type authorityRegistryAdminStub struct {
	row          *models.TMarketAuthorityRegistry
	enabled      map[string]*models.TMarketAuthorityRegistry
	references   int64
	countCalled  bool
	updateCalled bool
}

func (s *authorityRegistryAdminStub) FindEnabled(
	_ context.Context,
	authority string,
) (*models.TMarketAuthorityRegistry, error) {
	row := s.enabled[authority]
	if row == nil {
		return nil, models.ErrNotFound
	}
	copy := *row
	return &copy, nil
}

func (s *authorityRegistryAdminStub) Create(context.Context, *models.TMarketAuthorityRegistry) (int64, error) {
	return 9, nil
}

func (s *authorityRegistryAdminStub) FindOne(context.Context, int64) (*models.TMarketAuthorityRegistry, error) {
	if s.row == nil {
		return nil, models.ErrNotFound
	}
	copy := *s.row
	return &copy, nil
}

func (s *authorityRegistryAdminStub) FindOneByAuthority(context.Context, string) (*models.TMarketAuthorityRegistry, error) {
	return nil, models.ErrNotFound
}

func (s *authorityRegistryAdminStub) FindPage(
	context.Context,
	models.AuthorityRegistryFilter,
	int64,
	int64,
) ([]*models.TMarketAuthorityRegistry, int64, error) {
	return nil, 0, nil
}

func (s *authorityRegistryAdminStub) CountActiveFormulaReferences(context.Context, string) (int64, error) {
	s.countCalled = true
	return s.references, nil
}

func (s *authorityRegistryAdminStub) UpdateConfigVersioned(
	context.Context,
	int64,
	int64,
	string,
	int64,
	int64,
) (bool, error) {
	s.updateCalled = true
	return true, nil
}

func TestNormalizeAuthorityRegistry(t *testing.T) {
	t.Parallel()

	authority, provider, producer, kinds, raw, err := normalizeAuthorityRegistry(
		" Binance-WS ",
		"binance",
		"binance_ws",
		[]string{"mark", "FINAL_QUOTE", "mark"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if authority != "binance-ws" || provider != "BINANCE" || producer != "BINANCE_WS" {
		t.Fatalf("unexpected normalization: %q %q %q", authority, provider, producer)
	}
	if !reflect.DeepEqual(kinds, []string{"FINAL_QUOTE", "MARK"}) {
		t.Fatalf("unexpected kinds: %#v", kinds)
	}
	if raw != `["FINAL_QUOTE","MARK"]` {
		t.Fatalf("unexpected JSON: %s", raw)
	}
}

func TestNormalizeAuthorityRegistryRejectsInvalidValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		authority string
		provider  string
		producer  string
		kinds     []string
	}{
		{name: "empty authority", provider: "SOURCE", producer: "SOURCE_WS", kinds: []string{"FINAL_QUOTE"}},
		{name: "invalid authority", authority: "source ws", provider: "SOURCE", producer: "SOURCE_WS", kinds: []string{"FINAL_QUOTE"}},
		{name: "empty provider", authority: "source", producer: "SOURCE_WS", kinds: []string{"FINAL_QUOTE"}},
		{name: "invalid provider", authority: "source", provider: "source provider", producer: "SOURCE_WS", kinds: []string{"FINAL_QUOTE"}},
		{name: "invalid producer", authority: "source", provider: "SOURCE", producer: "source ws", kinds: []string{"FINAL_QUOTE"}},
		{name: "invalid kind", authority: "source", provider: "SOURCE", producer: "SOURCE_WS", kinds: []string{"TRADE"}},
		{name: "empty kinds", authority: "source", provider: "SOURCE", producer: "SOURCE_WS"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, _, _, _, _, err := normalizeAuthorityRegistry(tt.authority, tt.provider, tt.producer, tt.kinds); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestSetAuthorityRegistryRejectsBreakingActiveFormula(t *testing.T) {
	t.Parallel()

	store := &authorityRegistryAdminStub{
		row: &models.TMarketAuthorityRegistry{
			Id:           1,
			Authority:    "itick-ws",
			ProviderCode: "ITICK",
			ProducerType: "ITICK_WS",
			AllowedKinds: `["FINAL_QUOTE"]`,
			Status:       1,
			Version:      3,
		},
		references: 1,
	}
	logic := NewSetAuthorityRegistryLogic(
		context.Background(),
		&svc.ServiceContext{AuthorityRegistryAdminModel: store},
	)
	_, err := logic.SetAuthorityRegistry(&market.SetAuthorityRegistryReq{
		Id:           1,
		Authority:    "itick-ws",
		ProviderCode: "ITICK",
		ProducerType: "ITICK_WS",
		AllowedKinds: []string{"FINAL_QUOTE"},
		Status:       common.Enable_ENABLE_DISABLED,
		Version:      3,
	})
	if err == nil || err.Error() != "authority is referenced by active price formulas" {
		t.Fatalf("unexpected error: %v", err)
	}
	if !store.countCalled || store.updateCalled {
		t.Fatalf("unexpected calls: count=%v update=%v", store.countCalled, store.updateCalled)
	}
}

func TestSetAuthorityRegistryAllowsNonBreakingExpansion(t *testing.T) {
	t.Parallel()

	store := &authorityRegistryAdminStub{
		row: &models.TMarketAuthorityRegistry{
			Id:           2,
			Authority:    "source-ws",
			ProviderCode: "SOURCE",
			ProducerType: "SOURCE_WS",
			AllowedKinds: `["FINAL_QUOTE"]`,
			Status:       1,
			Version:      4,
		},
		references: 1,
	}
	logic := NewSetAuthorityRegistryLogic(
		context.Background(),
		&svc.ServiceContext{AuthorityRegistryAdminModel: store},
	)
	resp, err := logic.SetAuthorityRegistry(&market.SetAuthorityRegistryReq{
		Id:           2,
		Authority:    "source-ws",
		ProviderCode: "SOURCE",
		ProducerType: "SOURCE_WS",
		AllowedKinds: []string{"FINAL_QUOTE", "INDEX"},
		Status:       common.Enable_ENABLE_ENABLED,
		Version:      4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.GetData().GetVersion() != 5 {
		t.Fatalf("unexpected version: %d", resp.GetData().GetVersion())
	}
	if store.countCalled || !store.updateCalled {
		t.Fatalf("unexpected calls: count=%v update=%v", store.countCalled, store.updateCalled)
	}
}

func TestCreateIndexRejectsAuthoritiesFromSameProvider(t *testing.T) {
	t.Parallel()

	enabled := map[string]*models.TMarketAuthorityRegistry{
		"price-engine": {
			Authority:    "price-engine",
			ProviderCode: "PRICE_ENGINE",
			AllowedKinds: `["INDEX"]`,
			Status:       1,
		},
		"itick-ws": {
			Authority:    "itick-ws",
			ProviderCode: "ITICK",
			AllowedKinds: `["FINAL_QUOTE"]`,
			Status:       1,
		},
		"itick-rest": {
			Authority:    "itick-rest",
			ProviderCode: "ITICK",
			AllowedKinds: `["FINAL_QUOTE"]`,
			Status:       1,
		},
		"source-c": {
			Authority:    "source-c",
			ProviderCode: "SOURCE_C",
			AllowedKinds: `["FINAL_QUOTE"]`,
			Status:       1,
		},
	}
	store := &authorityRegistryAdminStub{enabled: enabled}
	logic := NewCreatePriceFormulaLogic(
		context.Background(),
		&svc.ServiceContext{AuthorityRegistryModel: store},
	)
	_, err := logic.CreatePriceFormula(&market.CreatePriceFormulaReq{
		FormulaNo:      "index-provider-guard-v1",
		Authority:      "price-engine",
		SnapshotKind:   "INDEX",
		CategoryCode:   "crypto",
		Market:         "BA",
		Symbol:         "BTCUSDT",
		Algorithm:      market.PriceAlgorithm_PRICE_ALGORITHM_MEDIAN,
		FormulaVersion: "provider-guard-v1",
		Components: []*market.PriceFormulaComponent{
			{Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", Symbol: "BTCUSDT", Weight: "1"},
			{Authority: "itick-rest", SnapshotKind: "FINAL_QUOTE", Symbol: "BTCUSDT", Weight: "1"},
			{Authority: "source-c", SnapshotKind: "FINAL_QUOTE", Symbol: "BTCUSDT", Weight: "1"},
		},
		MaxLookbackMs: 30000,
		MinInputCount: 3,
		IntervalMs:    1000,
	})
	if err == nil || err.Error() != "INDEX and DELIVERY components must use independent providers" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemovesAuthorityKindRejectsInvalidStoredJSON(t *testing.T) {
	t.Parallel()

	if _, err := removesAuthorityKind("{", []string{"FINAL_QUOTE"}); err == nil {
		t.Fatal("expected stored JSON error")
	}
}

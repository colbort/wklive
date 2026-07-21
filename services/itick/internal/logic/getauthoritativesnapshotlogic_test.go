package logic

import (
	"context"
	"testing"

	"wklive/proto/itick"
	"wklive/services/itick/internal/svc"
	"wklive/services/itick/models"

	"github.com/shopspring/decimal"
)

type authoritativeSnapshotModelStub struct {
	row       *models.TItickAuthoritativeSnapshot
	authority string
	category  string
	market    string
	symbol    string
	target    int64
	minimum   int64
	kind      string
}

func (s *authoritativeSnapshotModelStub) FindAfterID(context.Context, int64, int64) ([]*models.TItickAuthoritativeSnapshot, error) {
	return nil, nil
}

func (s *authoritativeSnapshotModelStub) FindOneBySnapshotId(context.Context, string) (*models.TItickAuthoritativeSnapshot, error) {
	return s.row, nil
}

type authorityRegistryModelStub struct {
	row *models.TItickAuthorityRegistry
}

func (s authorityRegistryModelStub) FindEnabled(context.Context, string) (*models.TItickAuthorityRegistry, error) {
	return s.row, nil
}

func (s *authoritativeSnapshotModelStub) InsertImmutable(context.Context, *models.TItickAuthoritativeSnapshot) error {
	return nil
}
func (s *authoritativeSnapshotModelStub) InsertImmutableAndEnqueue(context.Context, *models.TItickAuthoritativeSnapshot, string) error {
	return nil
}

func (s *authoritativeSnapshotModelStub) FindAtOrBefore(_ context.Context, authority, kind, category, market, symbol string, target, minimum int64) (*models.TItickAuthoritativeSnapshot, error) {
	s.authority, s.category, s.market, s.symbol = authority, category, market, symbol
	s.kind = kind
	s.target, s.minimum = target, minimum
	return s.row, nil
}

func TestGetAuthoritativeSnapshotNormalizesAndUsesPermanentArchive(t *testing.T) {
	model := &authoritativeSnapshotModelStub{row: &models.TItickAuthoritativeSnapshot{SnapshotId: "snap", Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT", Price: decimal.RequireFromString("123.456789012345678901"), SourceTimestamp: 900, SnapshotTimestamp: 901, Revision: 900, FormulaVersion: "source-quote-v1", RawPayload: `{}`}}
	logic := NewGetAuthoritativeSnapshotLogic(context.Background(), &svc.ServiceContext{AuthoritativeSnapshotModel: model, AuthorityRegistryModel: authorityRegistryModelStub{row: &models.TItickAuthorityRegistry{Authority: "itick-ws", AllowedKinds: `["FINAL_QUOTE"]`, Status: 1}}})
	resp, err := logic.GetAuthoritativeSnapshot(&itick.GetAuthoritativeSnapshotReq{Authority: " ITICK-WS ", SnapshotKind: " final_quote ", CategoryCode: " Crypto ", Market: "ba", Symbol: "btcusdt", TargetTime: 1000, MaxLookbackMs: 200})
	if err != nil {
		t.Fatal(err)
	}
	if model.authority != "itick-ws" || model.kind != "FINAL_QUOTE" || model.category != "crypto" || model.market != "BA" || model.symbol != "BTCUSDT" || model.target != 1000 || model.minimum != 800 {
		t.Fatalf("query was not normalized: %#v", model)
	}
	if resp.GetData().GetPrice() != "123.456789012345678901" {
		t.Fatalf("price precision changed: %s", resp.GetData().GetPrice())
	}
}

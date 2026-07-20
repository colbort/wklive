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
}

func (s *authoritativeSnapshotModelStub) InsertIgnore(context.Context, *models.TItickAuthoritativeSnapshot) error {
	return nil
}

func (s *authoritativeSnapshotModelStub) FindAtOrBefore(_ context.Context, authority, category, market, symbol string, target, minimum int64) (*models.TItickAuthoritativeSnapshot, error) {
	s.authority, s.category, s.market, s.symbol = authority, category, market, symbol
	s.target, s.minimum = target, minimum
	return s.row, nil
}

func TestGetAuthoritativeSnapshotNormalizesAndUsesPermanentArchive(t *testing.T) {
	model := &authoritativeSnapshotModelStub{row: &models.TItickAuthoritativeSnapshot{SnapshotId: "snap", Authority: "itick-ws", SnapshotKind: "FINAL_QUOTE", CategoryCode: "crypto", Market: "BA", Symbol: "BTCUSDT", Price: decimal.RequireFromString("123.456789012345678901"), SourceTimestamp: 900, SnapshotTimestamp: 901, Revision: 900, FormulaVersion: "source-quote-v1", RawPayload: `{}`}}
	logic := NewGetAuthoritativeSnapshotLogic(context.Background(), &svc.ServiceContext{AuthoritativeSnapshotModel: model})
	resp, err := logic.GetAuthoritativeSnapshot(&itick.GetAuthoritativeSnapshotReq{Authority: " ITICK-WS ", CategoryCode: " Crypto ", Market: "ba", Symbol: "btcusdt", TargetTime: 1000, MaxLookbackMs: 200})
	if err != nil {
		t.Fatal(err)
	}
	if model.authority != "itick-ws" || model.category != "crypto" || model.market != "BA" || model.symbol != "BTCUSDT" || model.target != 1000 || model.minimum != 800 {
		t.Fatalf("query was not normalized: %#v", model)
	}
	if resp.GetData().GetPrice() != "123.456789012345678901" {
		t.Fatalf("price precision changed: %s", resp.GetData().GetPrice())
	}
}

package applogic

import (
	"context"

	"wklive/services/market/models"
)

type authoritativeSnapshotModelStub struct {
	row       *models.TMarketAuthoritativeSnapshot
	authority string
	category  string
	market    string
	symbol    string
	target    int64
	minimum   int64
	kind      string
}

func (s *authoritativeSnapshotModelStub) FindAfterID(context.Context, int64, int64) ([]*models.TMarketAuthoritativeSnapshot, error) {
	return nil, nil
}

func (s *authoritativeSnapshotModelStub) FindProductKeys(context.Context) ([]models.AuthoritativeSnapshotProductKey, error) {
	return nil, nil
}

func (s *authoritativeSnapshotModelStub) FindOneBySnapshotId(context.Context, string) (*models.TMarketAuthoritativeSnapshot, error) {
	return s.row, nil
}

type authorityRegistryModelStub struct {
	row *models.TMarketAuthorityRegistry
}

func (s authorityRegistryModelStub) FindEnabled(context.Context, string) (*models.TMarketAuthorityRegistry, error) {
	return s.row, nil
}

func (s *authoritativeSnapshotModelStub) InsertImmutable(context.Context, *models.TMarketAuthoritativeSnapshot) error {
	return nil
}
func (s *authoritativeSnapshotModelStub) InsertImmutableAndEnqueue(context.Context, *models.TMarketAuthoritativeSnapshot, string) error {
	return nil
}

func (s *authoritativeSnapshotModelStub) FindAtOrBefore(_ context.Context, authority, kind, category, market, symbol string, target, minimum int64) (*models.TMarketAuthoritativeSnapshot, error) {
	s.authority, s.category, s.market, s.symbol = authority, category, market, symbol
	s.kind = kind
	s.target, s.minimum = target, minimum
	return s.row, nil
}

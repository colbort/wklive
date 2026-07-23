package applogic

import (
	"context"

	"wklive/services/itick/models"
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

func (s *authoritativeSnapshotModelStub) FindLatestPage(context.Context, int64, int64) ([]*models.TItickAuthoritativeSnapshot, error) {
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

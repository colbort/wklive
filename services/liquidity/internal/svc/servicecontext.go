package svc

import "wklive/services/liquidity/internal/config"

// ServiceContext intentionally contains only configuration in the SQL-review
// phase. Database models and RPC clients are added after the schema is approved.
type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{Config: c}
}

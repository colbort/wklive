package config

import (
	"path/filepath"
	"testing"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/prometheus"
)

func TestOptionConfigExposesInternalPrometheusEndpoint(t *testing.T) {
	var c struct {
		ListenOn   string
		Prometheus prometheus.Config
	}
	path := filepath.Join("..", "..", "etc", "option.yaml")
	if err := conf.Load(path, &c); err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	if c.Prometheus.Host != "0.0.0.0" ||
		c.Prometheus.Port != 9105 ||
		c.Prometheus.Path != "/metrics" {
		t.Fatalf("unexpected prometheus config: %+v", c.Prometheus)
	}
	if c.ListenOn != "0.0.0.0:8085" {
		t.Fatalf("business RPC listen address changed: %s", c.ListenOn)
	}
}

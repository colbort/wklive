package systemlogic

import (
	"context"
	"net"
	"net/url"
	"strings"

	"wklive/common/helper"
	"wklive/common/i18n"
	"wklive/proto/system"
)

func normalizeTenantDomainOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Hostname())
	ip := net.ParseIP(hostname)
	allowHTTP := hostname == "localhost" || (ip != nil && (ip.IsLoopback() || ip.IsPrivate()))
	if scheme != "https" && !(scheme == "http" && allowHTTP) {
		return "", false
	}
	return scheme + "://" + strings.ToLower(parsed.Host), true
}

func validTenantDomainStatus(status system.TenantDomainStatus) bool {
	return status == system.TenantDomainStatus_TENANT_DOMAIN_STATUS_ACTIVE ||
		status == system.TenantDomainStatus_TENANT_DOMAIN_STATUS_RETIRED ||
		status == system.TenantDomainStatus_TENANT_DOMAIN_STATUS_DISABLED
}

func tenantDomainInvalidResp(ctx context.Context) *system.RespBase {
	return &system.RespBase{Base: helper.ErrResp(i18n.InvalidRequest, i18n.Translate(i18n.InvalidRequest, ctx))}
}

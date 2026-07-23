package useradminlogic

import (
	"net"
	"net/url"
	"strings"
)

func getLastFourDigits(accountNo string) string {
	if len(accountNo) <= 4 {
		return accountNo
	}
	return accountNo[len(accountNo)-4:]
}

func normalizeTransferOrigin(raw string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.User != nil || parsed.Host == "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	hostname := strings.ToLower(parsed.Hostname())
	ip := net.ParseIP(hostname)
	allowHTTP := hostname == "localhost" || (ip != nil && (ip.IsLoopback() || ip.IsPrivate()))
	if scheme != "https" && !(scheme == "http" && allowHTTP) {
		return "", false
	}
	if parsed.Path != "" && parsed.Path != "/" {
		return "", false
	}
	return scheme + "://" + strings.ToLower(parsed.Host), true
}

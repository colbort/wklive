package ws

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

type Config struct {
	AllowedOrigins     []string
	AllowMissingOrigin bool
}

func OriginAllowed(request *http.Request, config Config) bool {
	rawOrigin := strings.TrimSpace(request.Header.Get("Origin"))
	if rawOrigin == "" {
		return config.AllowMissingOrigin
	}
	origin, err := url.Parse(rawOrigin)
	if err != nil || origin.Host == "" ||
		(origin.Scheme != "http" && origin.Scheme != "https") {
		return false
	}
	normalized := strings.TrimRight(strings.ToLower(origin.String()), "/")
	for _, allowed := range config.AllowedOrigins {
		if normalized == strings.TrimRight(strings.ToLower(strings.TrimSpace(allowed)), "/") {
			return true
		}
	}
	return sameHost(origin.Host, request.Host)
}

func sameHost(left, right string) bool {
	leftHost, leftPort, leftErr := net.SplitHostPort(left)
	rightHost, rightPort, rightErr := net.SplitHostPort(right)
	if leftErr != nil {
		leftHost, leftPort = left, ""
	}
	if rightErr != nil {
		rightHost, rightPort = right, ""
	}
	return strings.EqualFold(leftHost, rightHost) && leftPort == rightPort
}

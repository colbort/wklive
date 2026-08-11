package chat_token

import (
	"net"
	"strings"
)

// isIPAllowed accepts both exact IP addresses and CIDR network entries.
// Exact matches remain supported for existing non-container deployments.
func isIPAllowed(clientIP string, whitelist []string) bool {
	clientIP = strings.TrimSpace(clientIP)
	parsedIP := net.ParseIP(clientIP)

	for _, entry := range whitelist {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if entry == clientIP {
			return true
		}

		_, network, err := net.ParseCIDR(entry)
		if err == nil && parsedIP != nil && network.Contains(parsedIP) {
			return true
		}
	}

	return false
}

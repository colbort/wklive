package utils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

func GetClientIP(r *http.Request) string {
	// 1. X-Forwarded-For（可能多个）
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		ips := strings.Split(ip, ",")
		return strings.TrimSpace(ips[0])
	}

	// 2. X-Real-IP
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// 3. RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func GetUserInfoFromMd(ctx context.Context) (int64, string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, "", fmt.Errorf("no metadata in context")
	}

	var userId int64
	if vals := md.Get(CtxKeyUid); len(vals) > 0 && vals[0] != "" {
		var err error
		userId, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, "", fmt.Errorf("invalid x-user-id: %w", err)
		}
	}

	var username string
	if vals := md.Get(CtxKeyUsername); len(vals) > 0 {
		username = vals[0]
	}

	return userId, username, nil
}

func GetUserIdFromMd(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata in context")
	}

	var userId int64
	if vals := md.Get(CtxKeyUid); len(vals) > 0 && vals[0] != "" {
		var err error
		userId, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid x-user-id: %w", err)
		}
	}

	return userId, nil
}

func GetUsernameFromMd(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	var username string
	if vals := md.Get(CtxKeyUsername); len(vals) > 0 {
		username = vals[0]
	}

	return username, nil
}

func GetTenantIdFromMd(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata in context")
	}

	var tenantId int64
	if vals := md.Get(CtxKeyTenantId); len(vals) > 0 && vals[0] != "" {
		var err error
		tenantId, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid x-tenant-id: %w", err)
		}
	}

	return tenantId, nil
}

func GetTenantCodeFromMd(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	var tenantCode string
	if vals := md.Get(CtxKeyTenantCode); len(vals) > 0 {
		tenantCode = vals[0]
	}

	return tenantCode, nil
}

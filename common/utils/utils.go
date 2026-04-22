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

	var uid int64
	if vals := md.Get("uid"); len(vals) > 0 && vals[0] != "" {
		var err error
		uid, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, "", fmt.Errorf("invalid x-user-id: %w", err)
		}
	}

	var username string
	if vals := md.Get("username"); len(vals) > 0 {
		username = vals[0]
	}

	return uid, username, nil
}

func GetUidFromMd(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata in context")
	}

	var uid int64
	if vals := md.Get("uid"); len(vals) > 0 && vals[0] != "" {
		var err error
		uid, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid x-user-id: %w", err)
		}
	}

	return uid, nil
}

func GetUsernameFromMd(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	var username string
	if vals := md.Get("username"); len(vals) > 0 {
		username = vals[0]
	}

	return username, nil
}

func GetTidFromMd(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, fmt.Errorf("no metadata in context")
	}

	var uid int64
	if vals := md.Get("tid"); len(vals) > 0 && vals[0] != "" {
		var err error
		uid, err = strconv.ParseInt(vals[0], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid x-user-id: %w", err)
		}
	}

	return uid, nil
}

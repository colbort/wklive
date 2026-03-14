package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
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

func StructToGoStruct(pbStruct *structpb.Struct, out interface{}) error {
	if pbStruct == nil {
		return nil
	}

	m := pbStruct.AsMap()

	b, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

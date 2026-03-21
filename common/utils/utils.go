package utils

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"wklive/proto/system"

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

func StringToStruct(data string) (*structpb.Struct, error) {
	var m map[string]interface{}

	// 1. JSON字符串 → map
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}

	// 2. map → structpb.Struct
	return structpb.NewStruct(m)
}

func StructToString(pbStruct *structpb.Struct) (string, error) {
	if pbStruct == nil {
		return "{}", nil
	}
	b, err := json.Marshal(pbStruct)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func CheckConfig(key string, value string) error {
	if key == system.SysConfigType_OBJECT_STORAGE.String() {
		var objectStorageConfig system.ObjectStorageConfig
		return json.Unmarshal([]byte(value), &objectStorageConfig)
	}

	return nil
}

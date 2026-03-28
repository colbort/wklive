package utils

import (
	"encoding/json"
	"wklive/proto/system"
)

func CheckConfig(key string, value string) error {
	switch key {
	case system.SysConfigType_OBJECT_STORAGE.String():
		var objectStorageConfig system.ObjectStorageConfig
		return json.Unmarshal([]byte(value), &objectStorageConfig)
	case system.SysConfigType_SYSTEM_CORE.String():
		var systemCore system.SystemCore
		return json.Unmarshal([]byte(value), &systemCore)
	case system.SysConfigType_ITICK_CONFIG.String():
		var itickConfig system.ItickConfig
		return json.Unmarshal([]byte(value), &itickConfig)
	default:
		return nil
	}
}

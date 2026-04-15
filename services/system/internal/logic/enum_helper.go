package logic

import (
	"strings"

	"wklive/proto/system"
)

func commonStatusToProto(v int64) system.CommonStatus {
	return system.CommonStatus(v)
}

func commonStatusToModel(v system.CommonStatus) int64 {
	return int64(v)
}

func menuTypeToProto(v int64) system.MenuType {
	return system.MenuType(v)
}

func menuTypeToModel(v system.MenuType) int64 {
	return int64(v)
}

func visibleStatusToProto(v int64) system.VisibleStatus {
	return system.VisibleStatus(v)
}

func visibleStatusToModel(v system.VisibleStatus) int64 {
	return int64(v)
}

func jobStatusToProto(v int64) system.JobStatus {
	return system.JobStatus(v)
}

func jobStatusToModel(v system.JobStatus) int64 {
	return int64(v)
}

func requestMethodToProto(v string) system.RequestMethod {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "GET":
		return system.RequestMethod_REQUEST_METHOD_GET
	case "POST":
		return system.RequestMethod_REQUEST_METHOD_POST
	case "PUT":
		return system.RequestMethod_REQUEST_METHOD_PUT
	case "DELETE":
		return system.RequestMethod_REQUEST_METHOD_DELETE
	default:
		return system.RequestMethod_REQUEST_METHOD_UNKNOWN
	}
}

func requestMethodToString(v system.RequestMethod) string {
	switch v {
	case system.RequestMethod_REQUEST_METHOD_GET:
		return "GET"
	case system.RequestMethod_REQUEST_METHOD_POST:
		return "POST"
	case system.RequestMethod_REQUEST_METHOD_PUT:
		return "PUT"
	case system.RequestMethod_REQUEST_METHOD_DELETE:
		return "DELETE"
	default:
		return ""
	}
}

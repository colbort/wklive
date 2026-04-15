package system

import (
	"strings"

	pbsystem "wklive/proto/system"
)

func toCommonStatus(v int64) pbsystem.CommonStatus {
	return pbsystem.CommonStatus(v)
}

func fromCommonStatus(v pbsystem.CommonStatus) int64 {
	return int64(v)
}

func toMenuType(v int64) pbsystem.MenuType {
	return pbsystem.MenuType(v)
}

func fromMenuType(v pbsystem.MenuType) int64 {
	return int64(v)
}

func toVisibleStatus(v int64) pbsystem.VisibleStatus {
	return pbsystem.VisibleStatus(v)
}

func fromVisibleStatus(v pbsystem.VisibleStatus) int64 {
	return int64(v)
}

func toJobStatus(v int64) pbsystem.JobStatus {
	return pbsystem.JobStatus(v)
}

func fromJobStatus(v pbsystem.JobStatus) int64 {
	return int64(v)
}

func toRequestMethod(v string) pbsystem.RequestMethod {
	switch strings.ToUpper(strings.TrimSpace(v)) {
	case "GET":
		return pbsystem.RequestMethod_REQUEST_METHOD_GET
	case "POST":
		return pbsystem.RequestMethod_REQUEST_METHOD_POST
	case "PUT":
		return pbsystem.RequestMethod_REQUEST_METHOD_PUT
	case "DELETE":
		return pbsystem.RequestMethod_REQUEST_METHOD_DELETE
	default:
		return pbsystem.RequestMethod_REQUEST_METHOD_UNKNOWN
	}
}

func fromRequestMethod(v pbsystem.RequestMethod) string {
	switch v {
	case pbsystem.RequestMethod_REQUEST_METHOD_GET:
		return "GET"
	case pbsystem.RequestMethod_REQUEST_METHOD_POST:
		return "POST"
	case pbsystem.RequestMethod_REQUEST_METHOD_PUT:
		return "PUT"
	case pbsystem.RequestMethod_REQUEST_METHOD_DELETE:
		return "DELETE"
	default:
		return ""
	}
}

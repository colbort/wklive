package helper

import "wklive/proto/common"

func OkResp() *common.RespBase {
	return &common.RespBase{Code: 200, Msg: "OK"}
}

func FailResp() *common.RespBase {
	return &common.RespBase{Code: 1, Msg: "FAIL"}
}

func FailWithCode(code int32) *common.RespBase {
	return &common.RespBase{Code: code, Msg: "FAIL"}
}

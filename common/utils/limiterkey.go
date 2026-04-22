package utils

import (
	"strconv"
)

func BuildUidKey(prefix string, uid int64) string {
	return prefix + ":uid:" + strconv.FormatInt(uid, 10)
}

func BuildIPKey(prefix, ip string) string {
	return prefix + ":ip:" + ip
}

func BuildStringKey(prefix, value string) string {
	return prefix + ":" + value
}

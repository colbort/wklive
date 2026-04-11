package conv

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
)

func ParseFloatField(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

func FloatString(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func NullStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

func BoolInt64(value bool) int64 {
	if value {
		return 1
	}
	return 0
}

func Int64Bool(value int64) bool {
	return value > 0
}

func GenerateBizNo(prefix string) string {
	return fmt.Sprintf("%s%s%06d", prefix, time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

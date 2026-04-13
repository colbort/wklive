package conv

import (
	"database/sql"
	"strconv"
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

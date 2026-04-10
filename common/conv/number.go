package conv

import (
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

func GenerateBizNo(prefix string) string {
	return fmt.Sprintf("%s%s%06d", prefix, time.Now().Format("20060102150405"), time.Now().UnixNano()%1000000)
}

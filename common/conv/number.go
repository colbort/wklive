package conv

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

func ParseFloatField(value string) (float64, error) {
	if value == "" {
		return 0, nil
	}
	return strconv.ParseFloat(value, 64)
}

// ParseDecimalField parses an exact numeric field transported as a string.
// Empty input is treated as zero to preserve the existing optional-field behavior.
func ParseDecimalField(value string) (decimal.Decimal, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(value)
}

// FloatString serializes both legacy floating-point values and exact decimals.
// New financial code should pass decimal.Decimal.
func FloatString(value any) string {
	switch v := value.(type) {
	case decimal.Decimal:
		return v.String()
	case decimal.NullDecimal:
		if v.Valid {
			return v.Decimal.String()
		}
		return ""
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	default:
		return fmt.Sprint(value)
	}
}

func NullStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}

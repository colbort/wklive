package models

import "github.com/shopspring/decimal"

// decimalAggregate gives go-zero's struct mapper an explicit destination for
// DECIMAL aggregate expressions. Scanning directly into decimal.Decimal makes
// the mapper treat its internal fields as result columns.
type decimalAggregate struct {
	Total string `db:"total"`
}

func (a decimalAggregate) Decimal() (decimal.Decimal, error) {
	if a.Total == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(a.Total)
}

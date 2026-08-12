package adminlogic

import (
	"context"
	"testing"

	"wklive/services/market/models"
)

func TestValidateSelectableProduct(t *testing.T) {
	tests := []struct {
		name    string
		product *models.TItickProduct
		valid   bool
	}{
		{name: "correct Shenzhen stock", product: &models.TItickProduct{Id: 1, Enabled: 1, CategoryCode: "stock", Market: "SZ", Exchange: "SZSE", Symbol: "000062"}, valid: true},
		{name: "wrong Shanghai stock", product: &models.TItickProduct{Id: 2, Enabled: 1, CategoryCode: "stock", Market: "SH", Exchange: "SZSE", Symbol: "000062"}, valid: false},
		{name: "disabled product", product: &models.TItickProduct{Id: 3, Enabled: 2, CategoryCode: "stock", Market: "SZ", Exchange: "SZSE", Symbol: "000062"}, valid: false},
		{name: "manual stock without exchange", product: &models.TItickProduct{Id: 4, Enabled: 1, CategoryCode: "stock", Market: "SZ", Symbol: "000062"}, valid: true},
		{name: "non stock product", product: &models.TItickProduct{Id: 5, Enabled: 1, CategoryCode: "forex", Market: "GB", Exchange: "FOREX", Symbol: "USDCNY"}, valid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateSelectableProduct(context.Background(), test.product)
			if (got == nil) != test.valid {
				t.Fatalf("validateSelectableProduct() response=%v valid=%v", got, test.valid)
			}
		})
	}
}

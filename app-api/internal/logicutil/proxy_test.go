package logicutil

import (
	"reflect"
	"testing"
)

func TestCopyValueMatchesGeneratedProtoUnderscoreField(t *testing.T) {
	type protoStatistics struct {
		Volume_24H     string
		TradeCount_24H int64
	}
	type apiStatistics struct {
		Volume24h     string
		TradeCount24h int64
	}
	src := protoStatistics{Volume_24H: "12.5", TradeCount_24H: 7}
	var dst apiStatistics
	if err := copyValue(reflect.ValueOf(&dst), reflect.ValueOf(&src)); err != nil {
		t.Fatal(err)
	}
	if dst.Volume24h != "12.5" || dst.TradeCount24h != 7 {
		t.Fatalf("canonical field mapping failed: %+v", dst)
	}
}

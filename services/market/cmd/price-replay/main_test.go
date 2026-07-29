package main

import "testing"

func TestDecodeAuditRecordsSupportsObjectArrayAndJSONL(t *testing.T) {
	object := []byte(`{"formula_no":"a"}`)
	records, err := decodeAuditRecords(object)
	if err != nil || len(records) != 1 {
		t.Fatalf("object records=%d err=%v", len(records), err)
	}
	records, err = decodeAuditRecords([]byte(`[{"formula_no":"a"},{"formula_no":"b"}]`))
	if err != nil || len(records) != 2 {
		t.Fatalf("array records=%d err=%v", len(records), err)
	}
	records, err = decodeAuditRecords([]byte("{\"formula_no\":\"a\"}\n{\"formula_no\":\"b\"}\n"))
	if err != nil || len(records) != 2 {
		t.Fatalf("jsonl records=%d err=%v", len(records), err)
	}
}

func TestDecodeAuditRecordsRejectsEmptyInput(t *testing.T) {
	if _, err := decodeAuditRecords(nil); err == nil {
		t.Fatal("empty audit input accepted")
	}
	if _, err := decodeAuditRecords([]byte(`[]`)); err == nil {
		t.Fatal("empty audit array accepted")
	}
}

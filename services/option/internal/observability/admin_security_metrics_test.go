package observability

import (
	"reflect"
	"testing"
)

func TestAdminRejectedMutationUsesBoundedLabels(t *testing.T) {
	counter := &fakeCounterVec{}
	original := adminRejectedMutation
	adminRejectedMutation = counter
	t.Cleanup(func() {
		adminRejectedMutation = original
	})

	RecordAdminRejectedMutation(9, "contract", "listed_economics")
	RecordAdminRejectedMutation(0, "unbounded-object", "unbounded-reason")

	want := []metricCall{
		{value: 1, labels: []string{"9", "contract", "listed_economics"}},
		{value: 1, labels: []string{"all", "other", "other"}},
	}
	if !reflect.DeepEqual(counter.calls, want) {
		t.Fatalf("admin rejected mutation calls=%+v want=%+v", counter.calls, want)
	}
}

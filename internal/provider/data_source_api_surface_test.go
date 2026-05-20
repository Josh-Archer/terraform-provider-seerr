package provider

import "testing"

func TestJSONFieldString(t *testing.T) {
	got, ok := jsonFieldString([]byte(`{"results":[{"id":1}],"pageInfo":{"page":1}}`), "results")
	if !ok {
		t.Fatal("expected results field")
	}
	if got != `[{"id":1}]` {
		t.Fatalf("unexpected field JSON %s", got)
	}
}

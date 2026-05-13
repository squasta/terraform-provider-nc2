package dsshared

import "testing"

// TestExtractList covers the canonical decoder.
func TestExtractList(t *testing.T) {
	t.Parallel()
	got := ExtractList(map[string]any{
		"data": []any{
			map[string]any{"id": "a"},
			map[string]any{"id": "b"},
		},
	})
	if len(got) != 2 {
		t.Errorf("len=%d", len(got))
	}
}

// TestExtractList_Missing returns nil.
func TestExtractList_Missing(t *testing.T) {
	t.Parallel()
	if ExtractList(map[string]any{}) != nil {
		t.Errorf("expected nil")
	}
}

// TestSortByID sorts ascending.
func TestSortByID(t *testing.T) {
	t.Parallel()
	items := []map[string]any{{"id": "b"}, {"id": "a"}}
	SortByID(items)
	if items[0]["id"] != "a" {
		t.Errorf("got %v", items)
	}
}

// TestStringOrEmpty handles nil and non-string.
func TestStringOrEmpty(t *testing.T) {
	t.Parallel()
	if StringOrEmpty(nil) != "" || StringOrEmpty(42) != "" || StringOrEmpty("x") != "x" {
		t.Errorf("StringOrEmpty unexpected behavior")
	}
}

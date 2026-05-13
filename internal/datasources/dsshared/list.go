package dsshared

import "sort"

// ExtractList decodes the canonical NC2 envelope `{"data": [...]}`
// into a slice of generic maps. Returns nil when the data field is
// absent or not an array.
//
// All US3 data sources share this decoder; per-data-source code is
// limited to mapping each map to a typed framework Object.
func ExtractList(body map[string]any) []map[string]any {
	raw, ok := body["data"].([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(raw))
	for _, r := range raw {
		if m, ok := r.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// StringOrEmpty pulls a typed string out of a generic map value
// graph. Returns "" on missing or non-string. Pure helper, used
// across every data-source decoder.
func StringOrEmpty(v any) string {
	s, _ := v.(string)
	return s
}

// SortByID sorts the slice in ascending order by the `id` field.
// Items missing an `id` field land at the front; this keeps the
// ordering deterministic without panicking on partial responses.
func SortByID(items []map[string]any) {
	sort.Slice(items, func(i, j int) bool {
		return StringOrEmpty(items[i]["id"]) < StringOrEmpty(items[j]["id"])
	})
}

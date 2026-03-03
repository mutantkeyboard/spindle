package spindle

import (
	"fmt"
	"math"
	"testing"
)

func TestSortOrderFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected SortOrder
	}{
		{"asc", ASC},
		{"desc", DESC},
		{"invalid", ASC},
		{"", ASC},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := SortOrderFromString(tt.input)
			if result != tt.expected {
				t.Errorf("SortOrderFromString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPageInfoStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pageInfo PageInfo
		expected int
	}{
		{"Page 1, limit 10", PageInfo{Page: 1, Limit: 10}, 0},
		{"Page 2, limit 10", PageInfo{Page: 2, Limit: 10}, 10},
		{"Page 3, limit 20", PageInfo{Page: 3, Limit: 20}, 40},
		{"With offset", PageInfo{Page: 2, Limit: 10, Offset: 25}, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pageInfo.Start()
			if result != tt.expected {
				t.Errorf("Start() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestPageInfoSortBy(t *testing.T) {
	t.Parallel()

	p := NewPageInfo(1, 10, 0, nil)
	p.SortBy("name", ASC).SortBy("date", DESC)

	if len(p.Sort) != 2 {
		t.Fatalf("expected 2 sort fields, got %d", len(p.Sort))
	}
	if p.Sort[0].Field != "name" || p.Sort[0].Order != ASC {
		t.Errorf("Sort[0] = %v, want {name asc}", p.Sort[0])
	}
	if p.Sort[1].Field != "date" || p.Sort[1].Order != DESC {
		t.Errorf("Sort[1] = %v, want {date desc}", p.Sort[1])
	}
}

func TestPageInfoNextPageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pageInfo PageInfo
		baseURL  string
		expected string
	}{
		{
			"Middle page",
			PageInfo{Page: 2, Limit: 10},
			"https://example.com/users",
			"https://example.com/users?page=3&limit=10",
		},
		{
			"First page",
			PageInfo{Page: 1, Limit: 20},
			"https://example.com/users",
			"https://example.com/users?page=2&limit=20",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pageInfo.NextPageURL(tt.baseURL)
			if result != tt.expected {
				t.Errorf("NextPageURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestPageInfoCursorFields(t *testing.T) {
	t.Parallel()

	p := &PageInfo{
		Cursor:     "abc123",
		HasMore:    true,
		NextCursor: "def456",
	}

	if p.Cursor != "abc123" {
		t.Errorf("Cursor = %q, want %q", p.Cursor, "abc123")
	}
	if !p.HasMore {
		t.Error("HasMore = false, want true")
	}
	if p.NextCursor != "def456" {
		t.Errorf("NextCursor = %q, want %q", p.NextCursor, "def456")
	}
}

func TestPageInfoPreviousPageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pageInfo PageInfo
		baseURL  string
		expected string
	}{
		{
			"Middle page",
			PageInfo{Page: 2, Limit: 10},
			"https://example.com/users",
			"https://example.com/users?page=1&limit=10",
		},
		{
			"First page returns empty",
			PageInfo{Page: 1, Limit: 20},
			"https://example.com/users",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pageInfo.PreviousPageURL(tt.baseURL)
			if result != tt.expected {
				t.Errorf("PreviousPageURL() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCursorValuesRoundTrip(t *testing.T) {
	t.Parallel()

	original := map[string]any{
		"id":         float64(42),
		"created_at": "2026-01-01T00:00:00Z",
	}

	p := &PageInfo{}
	p.SetNextCursor(original)

	if !p.HasMore {
		t.Error("HasMore = false, want true after SetNextCursor")
	}
	if p.NextCursor == "" {
		t.Fatal("NextCursor is empty after SetNextCursor")
	}

	// Simulate next request: cursor from previous response becomes input
	p2 := &PageInfo{Cursor: p.NextCursor}
	decoded := p2.CursorValues()

	if decoded == nil {
		t.Fatal("CursorValues() returned nil")
	}
	if decoded["id"] != float64(42) {
		t.Errorf("decoded[id] = %v, want 42", decoded["id"])
	}
	if decoded["created_at"] != "2026-01-01T00:00:00Z" {
		t.Errorf("decoded[created_at] = %v, want 2026-01-01T00:00:00Z", decoded["created_at"])
	}
}

func TestCursorValuesEmptyCursor(t *testing.T) {
	t.Parallel()

	p := &PageInfo{Cursor: ""}
	if vals := p.CursorValues(); vals != nil {
		t.Errorf("CursorValues() = %v, want nil for empty cursor", vals)
	}
}

func TestCursorValuesInvalidBase64(t *testing.T) {
	t.Parallel()

	p := &PageInfo{Cursor: "not-valid-base64!!!"}
	if vals := p.CursorValues(); vals != nil {
		t.Errorf("CursorValues() = %v, want nil for invalid base64", vals)
	}
}

func TestCursorValuesInvalidJSON(t *testing.T) {
	t.Parallel()

	// Valid base64 but not valid JSON
	p := &PageInfo{Cursor: "bm90LWpzb24"}
	if vals := p.CursorValues(); vals != nil {
		t.Errorf("CursorValues() = %v, want nil for invalid JSON", vals)
	}
}

func TestSetNextCursorMarshalError(t *testing.T) {
	t.Parallel()

	p := &PageInfo{Limit: 10}
	result := p.SetNextCursor(map[string]any{"bad": math.NaN()})

	if result != p {
		t.Error("SetNextCursor should return the same PageInfo on error")
	}
	if p.HasMore {
		t.Error("HasMore should remain false on marshal error")
	}
	if p.NextCursor != "" {
		t.Errorf("NextCursor = %q, want empty on marshal error", p.NextCursor)
	}
}

func TestNextCursorURL(t *testing.T) {
	t.Parallel()

	t.Run("with HasMore", func(t *testing.T) {
		p := &PageInfo{Limit: 20}
		p.SetNextCursor(map[string]any{"id": float64(42)})

		url := p.NextCursorURL("https://example.com/users")

		expected := fmt.Sprintf("https://example.com/users?cursor=%s&limit=20", p.NextCursor)
		if url != expected {
			t.Errorf("NextCursorURL() = %q, want %q", url, expected)
		}
	})

	t.Run("without HasMore", func(t *testing.T) {
		p := &PageInfo{Limit: 20}

		url := p.NextCursorURL("https://example.com/users")
		if url != "" {
			t.Errorf("NextCursorURL() = %q, want empty string when HasMore is false", url)
		}
	})
}

func TestSetNextCursorChainable(t *testing.T) {
	t.Parallel()

	p := &PageInfo{Limit: 10}
	result := p.SetNextCursor(map[string]any{"id": float64(1)})

	if result != p {
		t.Error("SetNextCursor should return the same PageInfo for chaining")
	}
}

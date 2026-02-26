package spindle

import "testing"

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

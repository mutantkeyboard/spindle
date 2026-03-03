package spindle

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gofiber/fiber/v3"
)

type Response struct {
	Page            int         `json:"page"`
	Limit           int         `json:"limit"`
	Offset          int         `json:"offset"`
	Start           int         `json:"start"`
	Sort            []SortField `json:"sort"`
	NextPageURL     string      `json:"next_PageURL"`
	PreviousPageURL string      `json:"prev_PageURL"`
}

func Test_PaginateWithQueries(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Use(New(Config{
		DefaultSort: "id",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}

		return c.JSON(Response{
			Page:            pageInfo.Page,
			Limit:           pageInfo.Limit,
			Offset:          pageInfo.Offset,
			Start:           pageInfo.Start(),
			Sort:            pageInfo.Sort,
			NextPageURL:     pageInfo.NextPageURL(c.BaseURL()),
			PreviousPageURL: pageInfo.PreviousPageURL(c.BaseURL()),
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?page=2&limit=20", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	defer resp.Body.Close() //nolint:errcheck

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 2 {
		t.Errorf("Page = %d, want 2", respBody.Page)
	}
	if respBody.Limit != 20 {
		t.Errorf("Limit = %d, want 20", respBody.Limit)
	}
	if respBody.Offset != 0 {
		t.Errorf("Offset = %d, want 0", respBody.Offset)
	}
	if respBody.Start != 20 {
		t.Errorf("Start = %d, want 20", respBody.Start)
	}
	if respBody.NextPageURL != "http://example.com?page=3&limit=20" {
		t.Errorf("NextPageURL = %q, want %q", respBody.NextPageURL, "http://example.com?page=3&limit=20")
	}
	if respBody.PreviousPageURL != "http://example.com?page=1&limit=20" {
		t.Errorf("PreviousPageURL = %q, want %q", respBody.PreviousPageURL, "http://example.com?page=1&limit=20")
	}
	expectedSort := []SortField{{Field: "id", Order: ASC}}
	if !reflect.DeepEqual(respBody.Sort, expectedSort) {
		t.Errorf("Sort = %v, want %v", respBody.Sort, expectedSort)
	}
}

func Test_PaginateWithOffset(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:   pageInfo.Page,
			Limit:  pageInfo.Limit,
			Offset: pageInfo.Offset,
			Start:  pageInfo.Start(),
			Sort:   pageInfo.Sort,
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?offset=20&limit=20", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 1 {
		t.Errorf("Page = %d, want 1", respBody.Page)
	}
	if respBody.Limit != 20 {
		t.Errorf("Limit = %d, want 20", respBody.Limit)
	}
	if respBody.Offset != 20 {
		t.Errorf("Offset = %d, want 20", respBody.Offset)
	}
	if respBody.Start != 20 {
		t.Errorf("Start = %d, want 20", respBody.Start)
	}
}

func Test_PaginateCheckDefaultsWhenNoQueries(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:   pageInfo.Page,
			Limit:  pageInfo.Limit,
			Offset: pageInfo.Offset,
			Start:  pageInfo.Start(),
			Sort:   pageInfo.Sort,
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, fiber.StatusOK)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 1 {
		t.Errorf("Page = %d, want 1", respBody.Page)
	}
	if respBody.Limit != 10 {
		t.Errorf("Limit = %d, want 10", respBody.Limit)
	}
	if respBody.Offset != 0 {
		t.Errorf("Offset = %d, want 0", respBody.Offset)
	}
	if respBody.Start != 0 {
		t.Errorf("Start = %d, want 0", respBody.Start)
	}
	expectedSort := []SortField{{Field: "id", Order: ASC}}
	if !reflect.DeepEqual(respBody.Sort, expectedSort) {
		t.Errorf("Sort = %v, want %v", respBody.Sort, expectedSort)
	}
}

func Test_PaginateCheckDefaultsWhenNoPage(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?limit=20", nil))
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 1 {
		t.Errorf("Page = %d, want 1", respBody.Page)
	}
	if respBody.Limit != 20 {
		t.Errorf("Limit = %d, want 20", respBody.Limit)
	}
	if respBody.Start != 0 {
		t.Errorf("Start = %d, want 0", respBody.Start)
	}
}

func Test_PaginateCheckDefaultsWhenNoLimit(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?page=2", nil))
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 2 {
		t.Errorf("Page = %d, want 2", respBody.Page)
	}
	if respBody.Limit != 10 {
		t.Errorf("Limit = %d, want 10", respBody.Limit)
	}
	if respBody.Start != 10 {
		t.Errorf("Start = %d, want 10", respBody.Start)
	}
}

func Test_PaginateConfigDefaultPageDefaultLimit(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		DefaultPage:  100,
		DefaultLimit: MaxLimit,
		DefaultSort:  "name",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
			Sort:  pageInfo.Sort,
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 100 {
		t.Errorf("Page = %d, want 100", respBody.Page)
	}
	if respBody.Limit != MaxLimit {
		t.Errorf("Limit = %d, want %d", respBody.Limit, MaxLimit)
	}
	if respBody.Start != 9900 {
		t.Errorf("Start = %d, want 9900", respBody.Start)
	}
	expectedSort := []SortField{{Field: "name", Order: ASC}}
	if !reflect.DeepEqual(respBody.Sort, expectedSort) {
		t.Errorf("Sort = %v, want %v", respBody.Sort, expectedSort)
	}
}

func Test_PaginateConfigPageKeyLimitKey(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		PageKey:     "site",
		LimitKey:    "size",
		DefaultSort: "id",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
			Sort:  pageInfo.Sort,
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/?site=2&size=5", nil))
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 2 {
		t.Errorf("Page = %d, want 2", respBody.Page)
	}
	if respBody.Limit != 5 {
		t.Errorf("Limit = %d, want 5", respBody.Limit)
	}
	if respBody.Start != 5 {
		t.Errorf("Start = %d, want 5", respBody.Start)
	}
}

func Test_PaginateNegativeDefaultPageDefaultLimitValues(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		DefaultPage:  -1,
		DefaultLimit: -1,
		DefaultSort:  "id",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
		})
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatal(err)
	}

	if respBody.Page != 1 {
		t.Errorf("Page = %d, want 1", respBody.Page)
	}
	if respBody.Limit != 10 {
		t.Errorf("Limit = %d, want 10", respBody.Limit)
	}
	if respBody.Start != 0 {
		t.Errorf("Start = %d, want 0", respBody.Start)
	}
}

func Test_PaginateFromContextWithoutNew(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		_, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(nil)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func Test_PaginateNextSkip(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		Next: func(c fiber.Ctx) bool {
			return true
		},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		_, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(nil)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("status = %d, want %d (middleware should be skipped)", resp.StatusCode, fiber.StatusBadRequest)
	}
}

func Test_PaginateEdgeCases(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		DefaultSort:  "id",
		DefaultLimit: 10,
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(pageInfo)
	})

	testCases := []struct {
		name          string
		url           string
		expectedPage  int
		expectedLimit int
	}{
		{"Negative page", "/?page=-1", 1, 10},
		{"Page zero", "/?page=0", 1, 10},
		{"Negative limit", "/?limit=-10", 1, 10},
		{"Limit zero", "/?limit=0", 1, 10},
		{"Limit exceeds max", "/?limit=200", 1, MaxLimit},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", tc.url, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 200 {
				t.Fatalf("status = %d, want 200", resp.StatusCode)
			}

			var result PageInfo
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
			if result.Page != tc.expectedPage {
				t.Errorf("Page = %d, want %d", result.Page, tc.expectedPage)
			}
			if result.Limit != tc.expectedLimit {
				t.Errorf("Limit = %d, want %d", result.Limit, tc.expectedLimit)
			}
		})
	}
}

func Test_PaginateWithMultipleSorting(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		SortKey:      "sort",
		DefaultSort:  "id",
		AllowedSorts: []string{"id", "name", "date"},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Sort: pageInfo.Sort,
		})
	})

	testCases := []struct {
		name         string
		url          string
		expectedSort []SortField
	}{
		{"Default Sort", "/", []SortField{{Field: "id", Order: ASC}}},
		{"Single Sort", "/?sort=name", []SortField{{Field: "name", Order: ASC}}},
		{"Multiple Sort", "/?sort=name,-date", []SortField{{Field: "name", Order: ASC}, {Field: "date", Order: DESC}}},
		{"Invalid Sort", "/?sort=invalid", []SortField{{Field: "id", Order: ASC}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tc.url, nil))
			if err != nil {
				t.Fatal(err)
			}

			var result Response
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Sort, tc.expectedSort) {
				t.Errorf("Sort = %v, want %v", result.Sort, tc.expectedSort)
			}
		})
	}
}

func TestParseSortQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		query        string
		allowedSorts []string
		defaultSort  string
		expected     []SortField
	}{
		{
			"Empty query",
			"",
			[]string{"id", "name", "date"},
			"id",
			[]SortField{{Field: "id", Order: ASC}},
		},
		{
			"Single allowed field",
			"name",
			[]string{"id", "name", "date"},
			"id",
			[]SortField{{Field: "name", Order: ASC}},
		},
		{
			"Multiple fields with mixed order",
			"name,-date,id",
			[]string{"id", "name", "date"},
			"id",
			[]SortField{
				{Field: "name", Order: ASC},
				{Field: "date", Order: DESC},
				{Field: "id", Order: ASC},
			},
		},
		{
			"Disallowed field",
			"email,name",
			[]string{"id", "name", "date"},
			"id",
			[]SortField{{Field: "name", Order: ASC}},
		},
		{
			"All disallowed fields",
			"email,phone",
			[]string{"id", "name", "date"},
			"id",
			[]SortField{{Field: "id", Order: ASC}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseSortQuery(tt.query, tt.allowedSorts, tt.defaultSort)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseSortQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkPaginateMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, _ := FromContext(c)
		return c.JSON(pageInfo)
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/?page=2&limit=20&sort=name,-date", nil)
		_, err := app.Test(req, fiber.TestConfig{Timeout: 0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaginateMiddlewareWithCustomConfig(b *testing.B) {
	app := fiber.New()
	app.Use(New(Config{
		PageKey:      "p",
		LimitKey:     "l",
		SortKey:      "s",
		DefaultPage:  1,
		DefaultLimit: 30,
		DefaultSort:  "id",
		AllowedSorts: []string{"id", "name", "date"},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, _ := FromContext(c)
		return c.JSON(pageInfo)
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/?p=3&l=25&s=name,-id", nil)
		_, err := app.Test(req, fiber.TestConfig{Timeout: 0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

type CursorResponse struct {
	Cursor     string      `json:"cursor"`
	Limit      int         `json:"limit"`
	HasMore    bool        `json:"has_more"`
	NextCursor string      `json:"next_cursor"`
	Sort       []SortField `json:"sort"`
}

func Test_PaginateWithCursor(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		DefaultSort: "id",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(CursorResponse{
			Cursor: pageInfo.Cursor,
			Limit:  pageInfo.Limit,
			Sort:   pageInfo.Sort,
		})
	})

	// Encode a valid cursor: {"id": 42}
	cursorJSON := `{"id":42}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	resp, err := app.Test(httptest.NewRequest("GET", "/?cursor="+cursor+"&limit=20", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result CursorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Cursor != cursor {
		t.Errorf("Cursor = %q, want %q", result.Cursor, cursor)
	}
	if result.Limit != 20 {
		t.Errorf("Limit = %d, want 20", result.Limit)
	}
}

func Test_PaginateCursorPriorityOverPage(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(pageInfo)
	})

	cursorJSON := `{"id":42}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	// Both cursor and page present — cursor should win
	resp, err := app.Test(httptest.NewRequest("GET", "/?cursor="+cursor+"&page=5&limit=10", nil))
	if err != nil {
		t.Fatal(err)
	}

	var result PageInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Cursor != cursor {
		t.Errorf("Cursor = %q, want %q", result.Cursor, cursor)
	}
	// Page should be 0 since cursor mode ignores it
	if result.Page != 0 {
		t.Errorf("Page = %d, want 0 in cursor mode", result.Page)
	}
}

func Test_PaginateEmptyCursorIsFirstPage(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(pageInfo)
	})

	// Empty cursor = first page, should not error
	resp, err := app.Test(httptest.NewRequest("GET", "/?cursor=&limit=10", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200 for empty cursor", resp.StatusCode)
	}

	var result PageInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Cursor != "" {
		t.Errorf("Cursor = %q, want empty string", result.Cursor)
	}
}

func Test_PaginateInvalidCursorReturns400(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New())

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, _ := FromContext(c)
		return c.JSON(pageInfo)
	})

	testCases := []struct {
		name   string
		cursor string
	}{
		{"Invalid base64", "not-valid!!!"},
		{"Valid base64 but invalid JSON", base64.RawURLEncoding.EncodeToString([]byte("not-json"))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest("GET", "/?cursor="+tc.cursor, nil))
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != 400 {
				t.Errorf("status = %d, want 400 for %s", resp.StatusCode, tc.name)
			}
		})
	}
}

func Test_PaginateCursorWithSort(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		SortKey:      "sort",
		DefaultSort:  "id",
		AllowedSorts: []string{"id", "name"},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(CursorResponse{
			Cursor: pageInfo.Cursor,
			Sort:   pageInfo.Sort,
		})
	})

	cursorJSON := `{"id":42}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	resp, err := app.Test(httptest.NewRequest("GET", "/?cursor="+cursor+"&sort=name,-id", nil))
	if err != nil {
		t.Fatal(err)
	}

	var result CursorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	expectedSort := []SortField{{Field: "name", Order: ASC}, {Field: "id", Order: DESC}}
	if !reflect.DeepEqual(result.Sort, expectedSort) {
		t.Errorf("Sort = %v, want %v", result.Sort, expectedSort)
	}
}

func Test_PaginateCursorWithCustomKey(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		CursorKey: "after",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(CursorResponse{
			Cursor: pageInfo.Cursor,
			Limit:  pageInfo.Limit,
		})
	})

	cursorJSON := `{"id":1}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	resp, err := app.Test(httptest.NewRequest("GET", "/?after="+cursor, nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result CursorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Cursor != cursor {
		t.Errorf("Cursor = %q, want %q", result.Cursor, cursor)
	}
}

func Test_PaginateCursorWithParamAlias(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		CursorParam: "starting_after",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(CursorResponse{
			Cursor: pageInfo.Cursor,
		})
	})

	cursorJSON := `{"id":1}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	// Use the alias param name
	resp, err := app.Test(httptest.NewRequest("GET", "/?starting_after="+cursor, nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var result CursorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Cursor != cursor {
		t.Errorf("Cursor = %q, want %q", result.Cursor, cursor)
	}
}

func Test_PaginateNoCursorFallsBackToPageMode(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	app.Use(New(Config{
		DefaultSort: "id",
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, ok := FromContext(c)
		if !ok {
			return fiber.ErrBadRequest
		}
		return c.JSON(Response{
			Page:  pageInfo.Page,
			Limit: pageInfo.Limit,
			Start: pageInfo.Start(),
		})
	})

	// No cursor param — should behave exactly as before
	resp, err := app.Test(httptest.NewRequest("GET", "/?page=3&limit=15", nil))
	if err != nil {
		t.Fatal(err)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	if result.Page != 3 {
		t.Errorf("Page = %d, want 3", result.Page)
	}
	if result.Limit != 15 {
		t.Errorf("Limit = %d, want 15", result.Limit)
	}
	if result.Start != 30 {
		t.Errorf("Start = %d, want 30", result.Start)
	}
}

func BenchmarkPaginateCursorMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(New(Config{
		SortKey:      "sort",
		DefaultSort:  "id",
		AllowedSorts: []string{"id", "name", "date"},
	}))

	app.Get("/", func(c fiber.Ctx) error {
		pageInfo, _ := FromContext(c)
		return c.JSON(pageInfo)
	})

	cursorJSON := `{"id":42,"created_at":"2026-01-01T00:00:00Z"}`
	cursor := base64.RawURLEncoding.EncodeToString([]byte(cursorJSON))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/?cursor="+cursor+"&limit=20&sort=name,-id", nil)
		_, err := app.Test(req, fiber.TestConfig{Timeout: 0})
		if err != nil {
			b.Fatal(err)
		}
	}
}

package spindle

import (
	"encoding/json"
	"io"
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

	body := resp.Body
	defer body.Close()

	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}

	var respBody Response
	if err := json.Unmarshal(bodyBytes, &respBody); err != nil {
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

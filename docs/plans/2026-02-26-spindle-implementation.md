# Spindle Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a Fiber v3 pagination middleware called Spindle, ported from fiberpaginate.

**Architecture:** Single-package Go module with middleware pattern — `New()` returns a `fiber.Handler` that parses query params into a `PageInfo` stored in context via `c.Locals()`. Consumers retrieve it with `FromContext()`.

**Tech Stack:** Go 1.25+, GoFiber v3 (`github.com/gofiber/fiber/v3`), standard library testing.

---

### Task 1: Initialize the Go module

**Files:**
- Create: `go.mod`

**Step 1: Initialize module**

Run:
```bash
cd /Users/tony/Projects/personal/spindle
go mod init github.com/mutantkeyboard/spindle
```

**Step 2: Add Fiber v3 dependency**

Run:
```bash
cd /Users/tony/Projects/personal/spindle
go get github.com/gofiber/fiber/v3
```

**Step 3: Commit**

```bash
cd /Users/tony/Projects/personal/spindle
git add go.mod go.sum
git commit -m "Initialize spindle Go module with Fiber v3 dependency"
```

---

### Task 2: Create PageInfo types

**Files:**
- Create: `page_info.go`

**Step 1: Write the failing test**

Create `page_info_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v -run "TestSortOrder|TestPageInfo"`
Expected: Compilation error — types not defined yet.

**Step 3: Write minimal implementation**

Create `page_info.go`:

```go
package spindle

import "fmt"

// SortOrder represents sort order.
type SortOrder string

const (
	ASC  SortOrder = "asc"
	DESC SortOrder = "desc"
)

// SortField represents a sort field with direction.
type SortField struct {
	Field string
	Order SortOrder
}

// SortOrderFromString returns a SortOrder from a string.
func SortOrderFromString(s string) SortOrder {
	switch s {
	case "asc":
		return ASC
	case "desc":
		return DESC
	default:
		return ASC
	}
}

// PageInfo contains pagination information.
type PageInfo struct {
	Page   int         `json:"page"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Sort   []SortField `json:"sort"`
}

// NewPageInfo creates a new PageInfo.
func NewPageInfo(page, limit, offset int, sort []SortField) *PageInfo {
	return &PageInfo{
		Page:   page,
		Limit:  limit,
		Offset: offset,
		Sort:   sort,
	}
}

// Start returns the start index based on page/limit or offset.
func (p *PageInfo) Start() int {
	if p.Offset > 0 {
		return p.Offset
	}
	return (p.Page - 1) * p.Limit
}

// SortBy adds a sort field. Chainable.
func (p *PageInfo) SortBy(field string, order SortOrder) *PageInfo {
	p.Sort = append(p.Sort, SortField{Field: field, Order: order})
	return p
}

// NextPageURL returns the URL for the next page.
func (p *PageInfo) NextPageURL(baseURL string) string {
	return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, p.Page+1, p.Limit)
}

// PreviousPageURL returns the URL for the previous page.
// Returns empty string if on page 1.
func (p *PageInfo) PreviousPageURL(baseURL string) string {
	if p.Page > 1 {
		return fmt.Sprintf("%s?page=%d&limit=%d", baseURL, p.Page-1, p.Limit)
	}
	return ""
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v -run "TestSortOrder|TestPageInfo"`
Expected: All PASS.

**Step 5: Commit**

```bash
cd /Users/tony/Projects/personal/spindle
git add page_info.go page_info_test.go
git commit -m "Add PageInfo types with tests"
```

---

### Task 3: Create Config

**Files:**
- Create: `config.go`

**Step 1: Write the failing test**

Create `config_test.go`:

```go
package spindle

import "testing"

func TestConfigDefault(t *testing.T) {
	t.Parallel()

	cfg := configDefault()
	if cfg.PageKey != "page" {
		t.Errorf("PageKey = %q, want %q", cfg.PageKey, "page")
	}
	if cfg.DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 1)
	}
	if cfg.LimitKey != "limit" {
		t.Errorf("LimitKey = %q, want %q", cfg.LimitKey, "limit")
	}
	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 10)
	}
}

func TestConfigOverride(t *testing.T) {
	t.Parallel()

	cfg := configDefault(Config{
		PageKey:      "p",
		LimitKey:     "l",
		DefaultPage:  5,
		DefaultLimit: 50,
	})
	if cfg.PageKey != "p" {
		t.Errorf("PageKey = %q, want %q", cfg.PageKey, "p")
	}
	if cfg.LimitKey != "l" {
		t.Errorf("LimitKey = %q, want %q", cfg.LimitKey, "l")
	}
	if cfg.DefaultPage != 5 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 5)
	}
	if cfg.DefaultLimit != 50 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 50)
	}
}

func TestConfigNegativeDefaults(t *testing.T) {
	t.Parallel()

	cfg := configDefault(Config{
		DefaultPage:  -1,
		DefaultLimit: -1,
	})
	if cfg.DefaultPage != 1 {
		t.Errorf("DefaultPage = %d, want %d", cfg.DefaultPage, 1)
	}
	if cfg.DefaultLimit != 10 {
		t.Errorf("DefaultLimit = %d, want %d", cfg.DefaultLimit, 10)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v -run "TestConfig"`
Expected: Compilation error — Config type not defined yet.

**Step 3: Write minimal implementation**

Create `config.go`:

```go
package spindle

import "github.com/gofiber/fiber/v3"

// Config defines the config for the pagination middleware.
type Config struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c fiber.Ctx) bool

	// PageKey is the query string key for page number.
	PageKey string

	// DefaultPage is the default page number.
	DefaultPage int

	// LimitKey is the query string key for limit.
	LimitKey string

	// DefaultLimit is the default items per page.
	DefaultLimit int

	// SortKey is the query string key for sort.
	SortKey string

	// DefaultSort is the default sort field.
	DefaultSort string

	// AllowedSorts is the list of allowed sort fields.
	AllowedSorts []string
}

// ConfigDefault is the default config.
var ConfigDefault = Config{
	Next:         nil,
	PageKey:      "page",
	DefaultPage:  1,
	LimitKey:     "limit",
	DefaultLimit: 10,
}

func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]

	if cfg.Next == nil {
		cfg.Next = ConfigDefault.Next
	}
	if cfg.PageKey == "" {
		cfg.PageKey = ConfigDefault.PageKey
	}
	if cfg.DefaultLimit < 1 {
		cfg.DefaultLimit = ConfigDefault.DefaultLimit
	}
	if cfg.LimitKey == "" {
		cfg.LimitKey = ConfigDefault.LimitKey
	}
	if cfg.DefaultPage < 1 {
		cfg.DefaultPage = ConfigDefault.DefaultPage
	}

	return cfg
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v -run "TestConfig"`
Expected: All PASS.

**Step 5: Commit**

```bash
cd /Users/tony/Projects/personal/spindle
git add config.go config_test.go
git commit -m "Add Config with defaults and tests"
```

---

### Task 4: Create middleware and FromContext

**Files:**
- Create: `paginate.go`

**Step 1: Write the failing test**

Create `paginate_test.go`:

```go
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
```

**Step 2: Run tests to verify they fail**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v`
Expected: Compilation error — `New` and `FromContext` not defined.

**Step 3: Write minimal implementation**

Create `paginate.go`:

```go
package spindle

import (
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type contextKey struct{}

var pageInfoKey = contextKey{}

// MaxLimit is the maximum limit allowed.
const MaxLimit = 100

// New creates a new pagination middleware handler.
func New(config ...Config) fiber.Handler {
	cfg := configDefault(config...)
	if cfg.DefaultSort == "" {
		cfg.DefaultSort = "id"
	}

	return func(c fiber.Ctx) error {
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		page := fiber.Query[int](c, cfg.PageKey, cfg.DefaultPage)
		if page < 1 {
			page = 1
		}

		limit := fiber.Query[int](c, cfg.LimitKey, cfg.DefaultLimit)
		if limit < 1 {
			limit = cfg.DefaultLimit
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		offset := fiber.Query[int](c, "offset", 0)
		if offset < 0 {
			offset = 0
		}

		sorts := parseSortQuery(c.Query(cfg.SortKey), cfg.AllowedSorts, cfg.DefaultSort)

		c.Locals(pageInfoKey, NewPageInfo(page, limit, offset, sorts))

		return c.Next()
	}
}

// FromContext returns the PageInfo from the context.
func FromContext(c fiber.Ctx) (*PageInfo, bool) {
	if pageInfo, ok := c.Locals(pageInfoKey).(*PageInfo); ok {
		return pageInfo, true
	}
	return nil, false
}

func parseSortQuery(query string, allowedSorts []string, defaultSort string) []SortField {
	if query == "" {
		return []SortField{{Field: defaultSort, Order: ASC}}
	}

	fields := strings.Split(query, ",")
	sortFields := make([]SortField, 0, len(fields))

	for _, field := range fields {
		order := ASC
		if strings.HasPrefix(field, "-") {
			order = DESC
			field = field[1:]
		}
		if slices.Contains(allowedSorts, field) {
			sortFields = append(sortFields, SortField{Field: field, Order: order})
		}
	}

	if len(sortFields) == 0 {
		return []SortField{{Field: defaultSort, Order: ASC}}
	}

	return sortFields
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v`
Expected: All PASS.

**Step 5: Commit**

```bash
cd /Users/tony/Projects/personal/spindle
git add paginate.go paginate_test.go
git commit -m "Add pagination middleware with full test suite"
```

---

### Task 5: Add benchmarks

**Files:**
- Modify: `paginate_test.go`

**Step 1: Append benchmarks to paginate_test.go**

Add to the end of `paginate_test.go`:

```go
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
```

**Step 2: Run benchmarks**

Run: `cd /Users/tony/Projects/personal/spindle && go test -bench=. -benchmem`
Expected: Benchmarks run and show allocations/op.

**Step 3: Commit**

```bash
cd /Users/tony/Projects/personal/spindle
git add paginate_test.go
git commit -m "Add benchmark tests"
```

---

### Task 6: Final verification

**Step 1: Run full test suite**

Run: `cd /Users/tony/Projects/personal/spindle && go test ./... -v -race`
Expected: All PASS, no race conditions.

**Step 2: Run go vet**

Run: `cd /Users/tony/Projects/personal/spindle && go vet ./...`
Expected: No issues.

**Step 3: Verify module is tidy**

Run: `cd /Users/tony/Projects/personal/spindle && go mod tidy`
Expected: No changes needed.

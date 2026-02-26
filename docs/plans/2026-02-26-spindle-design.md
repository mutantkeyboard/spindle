# Spindle — Fiber v3 Pagination Middleware

## Overview

Spindle is a pagination middleware for [GoFiber v3](https://github.com/gofiber/fiber), providing page/limit/offset and sort field extraction from query parameters. It is a direct port of [fiberpaginate](https://github.com/garrettladley/fiberpaginate) adapted to Fiber v3's API.

- **Module path:** `github.com/mutantkeyboard/spindle`
- **Go version:** 1.25+
- **Package:** `spindle` (single package)

## Approach

Direct port of fiberpaginate (Approach A). Copy the existing code, update all Fiber v2 → v3 API calls, and modernize with Go generics where applicable. No logic changes — same proven API shape.

## Module & Structure

```
spindle/
├── paginate.go        # Middleware handler (New)
├── page_info.go       # PageInfo type, SortOrder, SortField
├── config.go          # Config struct and defaults
├── paginate_test.go   # Test suite
├── go.mod
└── go.sum
```

## Public API

### Types

- `SortOrder` (string) — constants `ASC`, `DESC`
- `SortField` — `{Field string, Order SortOrder}`
- `PageInfo` — `{Page, Limit, Offset int, Sort []SortField}`
- `Config` — middleware configuration with `Next func(c fiber.Ctx) bool`

### Functions

- `New(config ...Config) fiber.Handler` — middleware factory
- `FromContext(c fiber.Ctx) (*PageInfo, bool)` — extract PageInfo from context
- `NewPageInfo(page, limit, offset int, sort []SortField) *PageInfo` — constructor
- `SortOrderFromString(s string) SortOrder` — parse sort order

### PageInfo Methods

- `Start() int` — calculate start index
- `SortBy(field string, order SortOrder) *PageInfo` — add sort field (chainable)
- `NextPageURL(baseURL string) string` — generate next page URL
- `PreviousPageURL(baseURL string) string` — generate previous page URL

### Constants

- `MaxLimit = 100`

## Implementation Changes (from fiberpaginate)

1. **Import path:** `github.com/gofiber/fiber/v2` → `github.com/gofiber/fiber/v3`
2. **Handler signatures:** `*fiber.Ctx` → `fiber.Ctx` (interface, not pointer)
3. **Query parsing:** `strconv.Atoi(c.Query("page"))` → `fiber.Query[int](c, "page", defaultValue)`
4. **Locals key:** Same `pageInfoKey` pattern — `c.Locals()` unchanged in v3
5. **Tests:** Updated to Fiber v3 `app.Test()` and `fiber.Ctx` interface

No logic changes to sorting, offset calculation, URL generation, or config defaults.

## Testing

- `go test ./...` — no local Fiber install needed, `go mod tidy` fetches dependencies
- Same test approach as fiberpaginate using `app.Test()` with `httptest.NewRequest`
- Coverage: page/limit defaults, offset calculation, sort parsing, MaxLimit enforcement, FromContext retrieval, URL generation, config Next skip

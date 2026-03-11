# Spindle

# IMPORTANT - Spindle has been fully integrated into the Fiber itelf as an official Paginate middleware, and as of March 11, 2026, this repo is archived


[![Go Reference](https://pkg.go.dev/badge/github.com/mutantkeyboard/spindle.svg)](https://pkg.go.dev/github.com/mutantkeyboard/spindle)
[![CI](https://github.com/mutantkeyboard/spindle/actions/workflows/ci.yml/badge.svg)](https://github.com/mutantkeyboard/spindle/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mutantkeyboard/spindle/branch/main/graph/badge.svg)](https://codecov.io/gh/mutantkeyboard/spindle)

Pagination middleware for [Fiber](https://github.com/gofiber/fiber) v3.

Spindle extracts `page`, `limit`, `offset`, and `sort` parameters from query strings and makes them available to your handlers via context.

## Install

```bash
go get github.com/mutantkeyboard/spindle
```

Requires Go 1.25+ and Fiber v3.

## Usage

### Basic

```go
package main

import (
    "github.com/gofiber/fiber/v3"
    "github.com/mutantkeyboard/spindle"
)

func main() {
    app := fiber.New()

    app.Use(spindle.New())

    app.Get("/users", func(c fiber.Ctx) error {
        pageInfo, ok := spindle.FromContext(c)
        if !ok {
            return fiber.ErrBadRequest
        }

        // pageInfo.Page   - current page (default: 1)
        // pageInfo.Limit  - items per page (default: 10, max: 100)
        // pageInfo.Offset - direct offset (default: 0)
        // pageInfo.Start() - calculated start index
        // pageInfo.Sort   - sort fields

        return c.JSON(pageInfo)
    })

    app.Listen(":3000")
}
```

Request: `GET /users?page=2&limit=20`

### With Sorting

```go
app.Use(spindle.New(spindle.Config{
    SortKey:      "sort",
    DefaultSort:  "created_at",
    AllowedSorts: []string{"created_at", "name", "email"},
}))
```

Request: `GET /users?page=1&limit=10&sort=name,-created_at`

Sort fields are comma-separated. Prefix with `-` for descending order.

### Cursor Pagination

For infinite scroll and keyset pagination:

```go
app.Use(spindle.New(spindle.Config{
    CursorKey: "cursor", // default
}))

app.Get("/users", func(c fiber.Ctx) error {
    pageInfo, ok := spindle.FromContext(c)
    if !ok {
        return fiber.ErrBadRequest
    }

    query := db.Model(&User{}).OrderBy("id ASC").Limit(pageInfo.Limit + 1)

    if vals := pageInfo.CursorValues(); vals != nil {
        query = query.Where("id > ?", vals["id"])
    }

    var users []User
    query.Find(&users)

    hasMore := len(users) > pageInfo.Limit
    if hasMore {
        users = users[:pageInfo.Limit]
        last := users[len(users)-1]
        pageInfo.SetNextCursor(map[string]any{"id": last.ID})
    }

    return c.JSON(fiber.Map{
        "data":        users,
        "has_more":    pageInfo.HasMore,
        "next_cursor": pageInfo.NextCursor,
    })
})
```

First request: `GET /users?limit=20`
Next request: `GET /users?cursor=<next_cursor>&limit=20`

Cursor tokens are opaque base64-encoded values. Invalid cursors return 400.

### Custom Config

```go
app.Use(spindle.New(spindle.Config{
    PageKey:      "p",
    LimitKey:     "size",
    DefaultPage:  1,
    DefaultLimit: 25,
    DefaultSort:  "id",
    AllowedSorts: []string{"id", "name", "date"},
    Next: func(c fiber.Ctx) bool {
        return c.Path() == "/health"
    },
}))
```

## Config

| Property | Type | Description | Default |
| -------- | ---- | ----------- | ------- |
| Next | `func(c fiber.Ctx) bool` | Skip middleware when returns true | `nil` |
| PageKey | `string` | Query key for page number | `"page"` |
| DefaultPage | `int` | Default page number | `1` |
| LimitKey | `string` | Query key for limit | `"limit"` |
| DefaultLimit | `int` | Default items per page | `10` |
| SortKey | `string` | Query key for sort | `""` |
| DefaultSort | `string` | Default sort field | `"id"` |
| AllowedSorts | `[]string` | Allowed sort field names | `[]` |
| CursorKey | `string` | Query key for cursor token | `"cursor"` |
| CursorParam | `string` | Optional alias for cursor key | `""` |

## PageInfo

Retrieved via `spindle.FromContext(c)`:

```go
type PageInfo struct {
    Page       int         // Current page number
    Limit      int         // Items per page (capped at 100)
    Offset     int         // Direct offset
    Sort       []SortField // Sort fields with direction
    Cursor     string      // Cursor token (empty if not in cursor mode)
    HasMore    bool        // True if more results exist (set by handler)
    NextCursor string      // Opaque cursor for next page (set by handler)
}
```

### Methods

- `Start() int` - Returns the start index. Uses `Offset` if set, otherwise `(Page-1) * Limit`.
- `SortBy(field string, order SortOrder) *PageInfo` - Adds a sort field. Chainable.
- `NextPageURL(baseURL string) string` - Returns the URL for the next page.
- `PreviousPageURL(baseURL string) string` - Returns the URL for the previous page. Empty string if on page 1.
- `CursorValues() map[string]any` - Decodes the cursor into key-value pairs. Returns nil if empty or invalid.
- `SetNextCursor(values map[string]any) *PageInfo` - Encodes values into an opaque cursor and sets HasMore. Chainable.
- `NextCursorURL(baseURL string) string` - Returns the URL for the next cursor page. Empty string if HasMore is false.

## Safety

- Limit is capped at `MaxLimit` (100) to prevent excessive memory usage
- Page values below 1 are reset to 1
- Negative offsets are reset to 0
- Sort fields are validated against `AllowedSorts`
- Invalid cursor tokens return 400 Bad Request

## Development

### Run tests locally

```bash
go test -race -v ./...
```

### Run tests in Docker

```bash
docker build -f Dockerfile.test -t spindle-test .
docker run --rm spindle-test
```

### Dev container

Open this project in VS Code with the Dev Containers extension to get a pre-configured Go development environment.

## Acknowledgements

Heavily inspired by [fiberpaginate](https://github.com/garrettladley/fiberpaginate) by Garrett Ladley.

## License

[MIT](LICENSE)

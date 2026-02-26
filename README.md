# Spindle

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
|----------|------|-------------|---------|
| Next | `func(c fiber.Ctx) bool` | Skip middleware when returns true | `nil` |
| PageKey | `string` | Query key for page number | `"page"` |
| DefaultPage | `int` | Default page number | `1` |
| LimitKey | `string` | Query key for limit | `"limit"` |
| DefaultLimit | `int` | Default items per page | `10` |
| SortKey | `string` | Query key for sort | `""` |
| DefaultSort | `string` | Default sort field | `"id"` |
| AllowedSorts | `[]string` | Allowed sort field names | `[]` |

## PageInfo

Retrieved via `spindle.FromContext(c)`:

```go
type PageInfo struct {
    Page   int         // Current page number
    Limit  int         // Items per page (capped at 100)
    Offset int         // Direct offset
    Sort   []SortField // Sort fields with direction
}
```

### Methods

- `Start() int` - Returns the start index. Uses `Offset` if set, otherwise `(Page-1) * Limit`.
- `SortBy(field string, order SortOrder) *PageInfo` - Adds a sort field. Chainable.
- `NextPageURL(baseURL string) string` - Returns the URL for the next page.
- `PreviousPageURL(baseURL string) string` - Returns the URL for the previous page. Empty string if on page 1.

## Safety

- Limit is capped at `MaxLimit` (100) to prevent excessive memory usage
- Page values below 1 are reset to 1
- Negative offsets are reset to 0
- Sort fields are validated against `AllowedSorts`

## License

[MIT](LICENSE)

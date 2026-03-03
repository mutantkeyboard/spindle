package spindle

import (
	"encoding/base64"
	"encoding/json"
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

		limit := fiber.Query(c, cfg.LimitKey, cfg.DefaultLimit)
		if limit < 1 {
			limit = cfg.DefaultLimit
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		sorts := parseSortQuery(c.Query(cfg.SortKey), cfg.AllowedSorts, cfg.DefaultSort)

		cursorRaw := c.Query(cfg.CursorKey)
		if cursorRaw == "" && cfg.CursorParam != "" {
			cursorRaw = c.Query(cfg.CursorParam)
		}

		if cursorRaw != "" {
			data, err := base64.RawURLEncoding.DecodeString(cursorRaw)
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid cursor"})
			}
			var obj map[string]any
			if err := json.Unmarshal(data, &obj); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid cursor"})
			}

			pageInfo := &PageInfo{
				Limit:  limit,
				Sort:   sorts,
				Cursor: cursorRaw,
			}
			c.Locals(pageInfoKey, pageInfo)
			return c.Next()
		}

		page := max(fiber.Query(c, cfg.PageKey, cfg.DefaultPage), 1)
		offset := max(fiber.Query(c, "offset", 0), 0)

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

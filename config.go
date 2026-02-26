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

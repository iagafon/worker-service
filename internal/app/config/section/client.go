package section

import "time"

type (
	// Client — конфигурация внешних клиентов.
	Client struct {
		Fixer ClientFixer
	}

	// ClientFixer — конфигурация клиента Fixer API.
	ClientFixer struct {
		ApiKey   string        `envconfig:"APP_CLIENT_FIXER_API_KEY"`
		BaseURL  string        `envconfig:"APP_CLIENT_FIXER_BASE_URL" default:"http://data.fixer.io/api"`
		CacheTTL time.Duration `envconfig:"APP_CLIENT_FIXER_CACHE_TTL" default:"30m"`
	}
)

package entity

import "errors"

// Доменные ошибки приложения.
// Используются в handlers и modules, маппятся на HTTP коды в respondent.
var (
	// ErrNotFound — ресурс не найден.
	ErrNotFound = errors.New("entity: not found")

	// ErrAlreadyExists — ресурс уже существует.
	ErrAlreadyExists = errors.New("entity: already exists")

	// ErrInvalidInput — некорректные входные данные.
	ErrInvalidInput = errors.New("entity: invalid input")

	// ErrForbidden — доступ запрещён.
	ErrForbidden = errors.New("entity: forbidden")

	// ErrUnauthorized — требуется авторизация.
	ErrUnauthorized = errors.New("entity: unauthorized")

	// ErrFixerInvalidApiKey — неверный API ключ Fixer.
	ErrFixerInvalidApiKey = errors.New("fixer: invalid api key")

	// ErrFixerRateLimitExceeded — превышен лимит запросов к Fixer API.
	ErrFixerRateLimitExceeded = errors.New("fixer: rate limit exceeded")

	// ErrFixerUnavailable — сервис Fixer недоступен.
	ErrFixerUnavailable = errors.New("fixer: service unavailable")

	// ErrFixerInvalidResponse — некорректный ответ от Fixer API.
	ErrFixerInvalidResponse = errors.New("fixer: invalid response")

	// ErrFixerCurrencyNotFound — валюта не найдена в Fixer API.
	ErrFixerCurrencyNotFound = errors.New("fixer: currency not found")
)

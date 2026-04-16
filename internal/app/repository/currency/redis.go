package currency

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/repository"
	rcredis "github.com/iagafon/worker-service/internal/app/repository/conn/redis"
)

const rateKeyPrefix = "rates:"

// Проверка реализации интерфейса на этапе компиляции.
var _ repository.CurrencyRate = (*RedisRepository)(nil)

// RedisRepository — реализация репозитория курсов валют на Redis.
type RedisRepository struct {
	client   *rcredis.Client
	cacheTTL time.Duration
}

// NewRedisRepository создаёт новый репозиторий курсов валют.
func NewRedisRepository(client *rcredis.Client, cacheTTL time.Duration) *RedisRepository {
	return &RedisRepository{client: client, cacheTTL: cacheTTL}
}

// buildKey создаёт ключ для кэша.
func (r *RedisRepository) buildKey(from, to string) string {
	return fmt.Sprintf("%s%s:%s", rateKeyPrefix, from, to)
}

// GetRate возвращает курс валюты из кэша.
func (r *RedisRepository) GetRate(ctx context.Context, from, to string) (float64, error) {
	key := r.buildKey(from, to)
	value, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	rate, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	log.Debug().Str("key", key).Str("value", value).Msg("Successfully get rate from cache")

	return rate, nil
}

// SetRate сохраняет курс валюты в кэш.
func (r *RedisRepository) SetRate(ctx context.Context, from, to string, rate float64) error {
	key := r.buildKey(from, to)
	value := strconv.FormatFloat(rate, 'f', -1, 64)
	return r.client.Set(ctx, key, value, r.cacheTTL).Err()
}

// SetRates сохраняет множество курсов валют в кэш.
func (r *RedisRepository) SetRates(ctx context.Context, from string, rates map[string]float64) error {
	pipe := r.client.Pipeline()
	for to, rate := range rates {
		pipe.Set(ctx, r.buildKey(from, to), strconv.FormatFloat(rate, 'f', -1, 64), r.cacheTTL)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	log.Debug().Int("Count:", len(rates)).Msg("Successfully set rates to cache")
	return nil
}

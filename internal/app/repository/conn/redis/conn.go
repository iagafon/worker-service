package rcredis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"

	"github.com/iagafon/worker-service/internal/app/config/section"
)

type (
	Client struct {
		client
		cfg section.RepositoryRedis
	}

	client = redis.Client
)

func NewConn(ctx context.Context, cfg section.RepositoryRedis) (*Client, error) {
	log.Debug().Msg("redis connection started")

	c := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if _, err := c.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	log.Debug().
		Str("address", cfg.Address).
		Int("db", cfg.DB).
		Msg("Redis connected")

	return &Client{client: *c, cfg: cfg}, nil
}

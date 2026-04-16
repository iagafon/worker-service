package section

import "time"

type (
	// Repository — конфигурация источников данных.
	Repository struct {
		Postgres RepositoryPostgres
		Redis    RepositoryRedis
	}

	// RepositoryPostgres — конфигурация подключения к PostgreSQL.
	RepositoryPostgres struct {
		Address        string        `envconfig:"APP_REPOSITORY_POSTGRES_ADDRESS" default:"127.0.0.1:5432"`
		Username       string        `envconfig:"APP_REPOSITORY_POSTGRES_USERNAME"`
		Password       string        `envconfig:"APP_REPOSITORY_POSTGRES_PASSWORD"`
		Name           string        `envconfig:"APP_REPOSITORY_POSTGRES_NAME"`
		MigrationTable string        `envconfig:"APP_REPOSITORY_POSTGRES_MIGRATION_TABLE" default:"schema_migrations"`
		ReadTimeout    time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_READ_TIMEOUT" default:"30s"`
		WriteTimeout   time.Duration `envconfig:"APP_REPOSITORY_POSTGRES_WRITE_TIMEOUT" default:"30s"`
	}

	RepositoryRedis struct {
		Address  string `envconfig:"APP_REPOSITORY_REDIS_ADDRESS" default:"localhost:6380"`
		Password string `envconfig:"APP_REPOSITORY_REDIS_PASSWORD" default:""`
		DB       int    `envconfig:"APP_REPOSITORY_REDIS_DB" default:"0"`
	}
)

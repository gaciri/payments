package utils

import (
	"github.com/go-pg/pg/v10"
	"github.com/redis/go-redis/v9"
	"payments/config"
)

func NewDbConnection(cfg *config.Config) *pg.DB {
	db := pg.Connect(&pg.Options{
		User:     cfg.Database.Username,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
		Addr:     cfg.Database.HostName,
	})
	return db
}

func NewRedisConnection(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.HostName,
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
	})
	return client
}

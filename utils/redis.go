package utils

import (
	"fmt"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

func GetMutexLock(rdb *redis.Client, domain, key string) *redsync.Mutex {
	pool := goredis.NewPool(rdb)
	rs := redsync.New(pool)
	mutexKey := fmt.Sprintf("%s:%s", domain, key)
	mutex := rs.NewMutex(mutexKey)
	return mutex
}

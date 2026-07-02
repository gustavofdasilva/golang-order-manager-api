package cache

import (
	"context"
	"fmt"
	"golang-order-manager-api/internal/config"
	"strings"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func InitRedis() {
	ctx := context.Background()

	addr := fmt.Sprintf("%s:%s", config.REDIS_HOST, strings.TrimLeft(config.REDIS_PORT, ":"))

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.REDIS_PASS,
		DB:       config.REDIS_DB,
	})

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
}

func GetRedisClient() *redis.Client {
	return rdb
}

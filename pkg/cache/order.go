package cache

import (
	"context"
	"encoding/json"
	"golang-order-manager-api/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisOrderCache struct {
	rdb *redis.Client
}

func NewRedisOrderCache() *RedisOrderCache {

	rdb := GetRedisClient()

	return &RedisOrderCache{rdb: rdb}
}

func (c *RedisOrderCache) GetOrder(ctx context.Context, id, userID uuid.UUID) (models.Order, error) {
	val, err := c.rdb.Get(ctx, "user:"+userID.String()+":order:"+id.String()).Bytes()
	if err == redis.Nil {
		return models.Order{}, nil
	}
	var order models.Order
	json.Unmarshal(val, &order)
	return order, nil
}

func (c *RedisOrderCache) SetOrder(ctx context.Context, order models.Order, userID uuid.UUID) error {
	val, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return c.rdb.Set(ctx, "user:"+userID.String()+":order:"+order.ID.String(), val, 0).Err()
}

func (c *RedisOrderCache) DeleteOrder(ctx context.Context, id, userID uuid.UUID) error {
	return c.rdb.Del(ctx, "user:"+userID.String()+":order:"+id.String()).Err()
}

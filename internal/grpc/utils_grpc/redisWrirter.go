package utils

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func WriteStringToRedis(ctx context.Context, redisClient *redis.Client, uuid, column string, data string) error {
	key := fmt.Sprintf("stats:%s", uuid)
	err := redisClient.HSet(ctx, key, column, data).Err()
	if err != nil {
		return fmt.Errorf("failed to write string to Redis: %v", err)
	}
	return nil
}
func WriteJsonToRedis(ctx context.Context, redisClient *redis.Client, uuid, column string, data []byte) error {
	key := fmt.Sprintf("stats:%s", uuid)
	err := redisClient.HSet(ctx, key, column, data).Err()
	if err != nil {
		return fmt.Errorf("failed to write JSON to Redis: %v", err)
	}
	return nil
}

func WriteFloat32ToRedis(ctx context.Context, redisClient *redis.Client, uuid, column string, data float32) error {
	key := fmt.Sprintf("stats:%s", uuid)
	err := redisClient.HSet(ctx, key, column, data).Err()
	if err != nil {
		return fmt.Errorf("failed to write JSON to Redis: %v", err)
	}
	return nil
}

package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ride-hailing/services/driver/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "driver:"
	ttl       = 15 * time.Minute
)

// Cache wraps Redis and implements domain.Cache.
// Keys look like: driver:3f2b1a...
type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, d *domain.Driver) error {
	data, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("marshal driver: %w", err)
	}
	return c.client.Set(ctx, key(d.ID), data, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, id string) (*domain.Driver, error) {
	data, err := c.client.Get(ctx, key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrDriverNotFound // cache miss — caller falls back to DB
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var d domain.Driver
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("unmarshal driver: %w", err)
	}
	return &d, nil
}

func (c *Cache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, key(id)).Err()
}

func key(id string) string {
	return keyPrefix + id
}

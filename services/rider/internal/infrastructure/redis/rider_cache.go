package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ride-hailing/services/rider/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "rider:"
	ttl       = 15 * time.Minute
)

type Cache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) Set(ctx context.Context, r *domain.Rider) error {
	data, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal rider: %w", err)
	}
	return c.client.Set(ctx, key(r.ID), data, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, id string) (*domain.Rider, error) {
	data, err := c.client.Get(ctx, key(id)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrRiderNotFound
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var r domain.Rider
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal rider: %w", err)
	}
	return &r, nil
}

func (c *Cache) Delete(ctx context.Context, id string) error {
	return c.client.Del(ctx, key(id)).Err()
}

func key(id string) string {
	return keyPrefix + id
}

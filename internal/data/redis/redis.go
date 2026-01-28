package redis

import (
	"context"
	"fmt"
	"time"

	"example.com/classic/internal/config"
	"example.com/classic/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Client Redis 客户端
type Client struct {
	client *redis.Client
	config *config.Config
	log    logger.Logger
}

// New 创建 Redis 客户端
func New(cfg *config.Config, log logger.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error(ctx, "failed to connect to Redis", logger.F("error", err))
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info(ctx, "Redis connected successfully",
		logger.F("host", cfg.Redis.Host),
		logger.F("port", cfg.Redis.Port),
		logger.F("database", cfg.Redis.Database))

	return &Client{
		client: rdb,
		config: cfg,
		log:    log,
	}, nil
}

// GetClient 获取 Redis 客户端
func (c *Client) GetClient() *redis.Client {
	return c.client
}

// Close 关闭 Redis 连接
func (c *Client) Close() error {
	return c.client.Close()
}

// Ping 检查 Redis 连接
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Set 设置键值对
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get 获取值
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

// Del 删除键
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists 检查键是否存在
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Incr 递增
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy 按指定值递增
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr 递减
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy 按指定值递减
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// Expire 设置过期时间
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// TTL 获取剩余过期时间
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

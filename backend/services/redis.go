package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"oneops/backend/config"
	"oneops/backend/logger"

	"go.uber.org/zap"
)

var (
	redisClient *redis.Client
	redisEnabled bool
)

// InitRedis 初始化 Redis 客户端
func InitRedis(cfg *config.RedisConfig) error {
	if !cfg.IsEnabled() {
		logger.Info("Redis 未启用，跳过初始化")
		redisEnabled = false
		return nil
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:         cfg.GetAddr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis 连接失败，缓存功能将不可用", zap.Error(err))
		redisEnabled = false
		return fmt.Errorf("Redis 连接失败: %w", err)
	}

	redisEnabled = true
	logger.Info("Redis 连接成功",
		zap.String("addr", cfg.GetAddr()),
		zap.Int("db", cfg.DB))
	return nil
}

// IsRedisEnabled 检查 Redis 是否可用
func IsRedisEnabled() bool {
	return redisEnabled && redisClient != nil
}

// RedisCache Redis 缓存操作封装
type RedisCache struct {
	ctx context.Context
}

// NewRedisCache 创建 Redis 缓存实例
func NewRedisCache() *RedisCache {
	return &RedisCache{
		ctx: context.Background(),
	}
}

// Set 设置缓存（默认 5 分钟过期）
func (c *RedisCache) Set(key string, value interface{}) error {
	return c.SetWithTTL(key, value, 5*time.Minute)
}

// SetWithTTL 设置缓存并指定过期时间
func (c *RedisCache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	if !IsRedisEnabled() {
		return nil // Redis 未启用，静默忽略
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化缓存数据失败: %w", err)
	}

	if err := redisClient.Set(c.ctx, c.buildKey(key), data, ttl).Err(); err != nil {
		logger.Warn("Redis Set 失败", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

// Get 获取缓存
func (c *RedisCache) Get(key string, dest interface{}) error {
	if !IsRedisEnabled() {
		return redis.Nil // Redis 未启用，返回未找到
	}

	data, err := redisClient.Get(c.ctx, c.buildKey(key)).Bytes()
	if err != nil {
		if err != redis.Nil {
			logger.Warn("Redis Get 失败", zap.String("key", key), zap.Error(err))
		}
		return err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("反序列化缓存数据失败: %w", err)
	}

	return nil
}

// Delete 删除缓存
func (c *RedisCache) Delete(key string) error {
	if !IsRedisEnabled() {
		return nil
	}

	if err := redisClient.Del(c.ctx, c.buildKey(key)).Err(); err != nil {
		logger.Warn("Redis Delete 失败", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

// DeleteByPattern 根据模式删除缓存
func (c *RedisCache) DeleteByPattern(pattern string) error {
	if !IsRedisEnabled() {
		return nil
	}

	iter := redisClient.Scan(c.ctx, 0, c.buildKey(pattern), 0).Iterator()
	keys := []string{}

	for iter.Next(c.ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		logger.Warn("Redis Scan 失败", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := redisClient.Del(c.ctx, keys...).Err(); err != nil {
			logger.Warn("Redis Del 失败", zap.Int("count", len(keys)), zap.Error(err))
			return err
		}
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *RedisCache) Exists(key string) bool {
	if !IsRedisEnabled() {
		return false
	}

	result, err := redisClient.Exists(c.ctx, c.buildKey(key)).Result()
	if err != nil {
		logger.Warn("Redis Exists 失败", zap.String("key", key), zap.Error(err))
		return false
	}

	return result > 0
}

// buildKey 构建带前缀的缓存键
func (c *RedisCache) buildKey(key string) string {
	return fmt.Sprintf("oneops:monitoring:%s", key)
}

// ========== 监控数据专用缓存方法 ==========

// CacheServerMetrics 缓存主机指标（默认 5 分钟 TTL）
func (c *RedisCache) CacheServerMetrics(serverID uint, metrics interface{}) error {
	key := fmt.Sprintf("server:%d:metrics", serverID)
	return c.Set(key, metrics)
}

// CacheServerMetricsWithTTL 缓存主机指标（指定 TTL）
func (c *RedisCache) CacheServerMetricsWithTTL(serverID uint, metrics interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("server:%d:metrics", serverID)
	return c.SetWithTTL(key, metrics, ttl)
}

// GetServerMetrics 获取主机指标缓存
func (c *RedisCache) GetServerMetrics(serverID uint, dest interface{}) error {
	key := fmt.Sprintf("server:%d:metrics", serverID)
	return c.Get(key, dest)
}

// CacheOverviewData 缓存监控概览数据
func (c *RedisCache) CacheOverviewData(data interface{}) error {
	key := "overview:data"
	return c.Set(key, data)
}

// GetOverviewData 获取监控概览数据缓存
func (c *RedisCache) GetOverviewData(dest interface{}) error {
	key := "overview:data"
	return c.Get(key, dest)
}

// CacheAlertRules 缓存告警规则
func (c *RedisCache) CacheAlertRules(rules interface{}) error {
	key := "alert:rules:active"
	return c.Set(key, rules)
}

// GetAlertRules 获取告警规则缓存
func (c *RedisCache) GetAlertRules(dest interface{}) error {
	key := "alert:rules:active"
	return c.Get(key, dest)
}

// InvalidateServerCache 使指定主机的缓存失效
func (c *RedisCache) InvalidateServerCache(serverID uint) error {
	pattern := fmt.Sprintf("server:%d:*", serverID)
	return c.DeleteByPattern(pattern)
}

// InvalidateAllCache 使所有监控缓存失效
func (c *RedisCache) InvalidateAllCache() error {
	if !IsRedisEnabled() {
		return nil
	}

	return c.DeleteByPattern("*")
}

// SetAlertSuppress 设置告警抑制标记（防止告警风暴）
func (c *RedisCache) SetAlertSuppress(serverID uint, ruleID string, duration time.Duration) error {
	key := fmt.Sprintf("alert:suppress:%d:%s", serverID, ruleID)
	return c.SetWithTTL(key, true, duration)
}

// IsAlertSuppressed 检查告警是否被抑制
func (c *RedisCache) IsAlertSuppressed(serverID uint, ruleID string) bool {
	key := fmt.Sprintf("alert:suppress:%d:%s", serverID, ruleID)
	return c.Exists(key)
}

// PublishAlert 发布告警事件（用于 WebSocket 推送）
func (c *RedisCache) PublishAlert(alert interface{}) error {
	if !IsRedisEnabled() {
		return nil
	}

	data, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("序列化告警事件失败: %w", err)
	}

	if err := redisClient.Publish(c.ctx, "oneops:alerts", data).Err(); err != nil {
		logger.Warn("发布告警事件失败", zap.Error(err))
		return err
	}

	return nil
}

// SubscribeAlerts 订阅告警事件
func (c *RedisCache) SubscribeAlerts(ctx context.Context) <-chan *redis.Message {
	if !IsRedisEnabled() {
		return nil
	}

	pubsub := redisClient.Subscribe(ctx, "oneops:alerts")
	ch := pubsub.Channel()

	go func() {
		<-ctx.Done()
		pubsub.Close()
	}()

	return ch
}

// CloseRedis 关闭 Redis 连接
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

package rcm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/zemld/config-manager/pkg/cm"
)

type RedisConfigManager struct {
	once sync.Once
	r    *redis.Client

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu          sync.RWMutex
	serviceName string
	config      map[string]string
	updatedAt   time.Time
}

func NewRedisConfigManager(serviceName string, redisOptions *redis.Options) cm.ConfigManager {
	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
	}

	rcm.once.Do(func() {
		r := redis.NewClient(redisOptions)
		status := r.Ping(context.Background())
		if status.Err() != nil {
			os.Exit(1)
		}
		rcm.r = r
	})

	rcm.ctx, rcm.cancel = context.WithCancel(context.Background())
	return rcm
}

func (rcm *RedisConfigManager) StartLoading(interval time.Duration) {
	rcm.wg.Add(1)

	go func() {
		defer rcm.wg.Done()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		rcm.fetchUpdates(ticker)
	}()
}

func (rcm *RedisConfigManager) fetchUpdates(ticker *time.Ticker) {
	for {
		select {
		case <-rcm.ctx.Done():
			return
		case <-ticker.C:
			rcm.LoadConfig(rcm.ctx)
		}
	}
}

func (rcm *RedisConfigManager) LoadConfig(ctx context.Context) error {
	rawConfig, err := rcm.r.Get(ctx, rcm.serviceName).Result()
	if err != nil {
		return fmt.Errorf("failed to get config: %w\n", err)
	}

	rawConfigMap := make(map[string]any)
	if err := json.Unmarshal([]byte(rawConfig), &rawConfigMap); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w\n", err)
	}

	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	for key, value := range rawConfigMap {
		rcm.config[key] = fmt.Sprintf("%v", value)
	}

	rcm.updatedAt = time.Now()

	return nil
}

func (rcm *RedisConfigManager) StopLoading() {
	rcm.cancel()
	rcm.r.Close()
	rcm.wg.Wait()
}

func (rcm *RedisConfigManager) GetInt(key string) (int, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	value, ok := rcm.config[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	return strconv.Atoi(value)
}

func (rcm *RedisConfigManager) GetFloat(key string) (float64, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	value, ok := rcm.config[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	return strconv.ParseFloat(value, 64)
}

func (rcm *RedisConfigManager) GetString(key string) (string, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	value, ok := rcm.config[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}

	return value, nil
}

func (rcm *RedisConfigManager) GetBool(key string) (bool, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	value, ok := rcm.config[key]
	if !ok {
		return false, fmt.Errorf("key %s not found", key)
	}

	return strconv.ParseBool(value)
}

func (rcm *RedisConfigManager) GetDuration(key string) (time.Duration, error) {
	rcm.mu.RLock()
	defer rcm.mu.RUnlock()

	value, ok := rcm.config[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	return time.ParseDuration(value)
}

func (rcm *RedisConfigManager) GetIntWithDefault(key string, defaultValue int) int {
	value, err := rcm.GetInt(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (rcm *RedisConfigManager) GetFloatWithDefault(key string, defaultValue float64) float64 {
	value, err := rcm.GetFloat(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (rcm *RedisConfigManager) GetStringWithDefault(key string, defaultValue string) string {
	value, err := rcm.GetString(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (rcm *RedisConfigManager) GetBoolWithDefault(key string, defaultValue bool) bool {
	value, err := rcm.GetBool(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (rcm *RedisConfigManager) GetDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	value, err := rcm.GetDuration(key)
	if err != nil {
		return defaultValue
	}

	return value
}

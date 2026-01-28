package rcm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func createTestConfig(t *testing.T, serviceName string) map[string]interface{} {
	config := map[string]interface{}{
		"int_key":      42,
		"float_key":   3.14,
		"string_key":  "test_value",
		"bool_key":    true,
		"duration_key": "5s",
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}

	return map[string]interface{}{
		serviceName: string(configJSON),
	}
}

func TestNewRedisConfigManager(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	rcm := NewRedisConfigManager(serviceName, &redis.Options{
		Addr: mr.Addr(),
	})

	if rcm == nil {
		t.Fatal("NewRedisConfigManager returned nil")
	}

	redisCM, ok := rcm.(*RedisConfigManager)
	if !ok {
		t.Fatal("NewRedisConfigManager returned wrong type")
	}

	if redisCM.serviceName != serviceName {
		t.Errorf("expected serviceName %s, got %s", serviceName, redisCM.serviceName)
	}

	if redisCM.config == nil {
		t.Error("config map is nil")
	}

	if redisCM.ctx == nil {
		t.Error("context is nil")
	}

	if redisCM.cancel == nil {
		t.Error("cancel function is nil")
	}
}

func TestLoadConfig(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	err := rcm.LoadConfig(context.Background())
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(rcm.config) == 0 {
		t.Error("config was not loaded")
	}

	if rcm.updatedAt.IsZero() {
		t.Error("updatedAt was not set")
	}

	value, ok := rcm.config["int_key"]
	if !ok {
		t.Error("int_key not found in config")
	}
	if value != "42" {
		t.Errorf("expected int_key to be '42', got '%s'", value)
	}
}

func TestLoadConfig_RedisError(t *testing.T) {
	mr, client := setupTestRedis(t)
	mr.Close()

	rcm := &RedisConfigManager{
		serviceName: "test_service",
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	err := rcm.LoadConfig(context.Background())
	if err == nil {
		t.Error("expected error when Redis is unavailable")
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	if err := mr.Set(serviceName, "invalid json"); err != nil {
		t.Fatalf("failed to set config in miniredis: %v", err)
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	err := rcm.LoadConfig(context.Background())
	if err == nil {
		t.Error("expected error when JSON is invalid")
	}
}

func TestGetInt(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value, err := rcm.GetInt("int_key")
	if err != nil {
		t.Fatalf("GetInt failed: %v", err)
	}

	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	_, err = rcm.GetInt("nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetFloat(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value, err := rcm.GetFloat("float_key")
	if err != nil {
		t.Fatalf("GetFloat failed: %v", err)
	}

	if value != 3.14 {
		t.Errorf("expected 3.14, got %f", value)
	}

	_, err = rcm.GetFloat("nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetString(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value, err := rcm.GetString("string_key")
	if err != nil {
		t.Fatalf("GetString failed: %v", err)
	}

	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	_, err = rcm.GetString("nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetBool(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value, err := rcm.GetBool("bool_key")
	if err != nil {
		t.Fatalf("GetBool failed: %v", err)
	}

	if !value {
		t.Error("expected true, got false")
	}

	_, err = rcm.GetBool("nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetDuration(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value, err := rcm.GetDuration("duration_key")
	if err != nil {
		t.Fatalf("GetDuration failed: %v", err)
	}

	expected := 5 * time.Second
	if value != expected {
		t.Errorf("expected %v, got %v", expected, value)
	}

	_, err = rcm.GetDuration("nonexistent_key")
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestGetIntWithDefault(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := rcm.GetIntWithDefault("int_key", 100)
	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	defaultValue := rcm.GetIntWithDefault("nonexistent_key", 100)
	if defaultValue != 100 {
		t.Errorf("expected default value 100, got %d", defaultValue)
	}
}

func TestGetFloatWithDefault(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := rcm.GetFloatWithDefault("float_key", 1.0)
	if value != 3.14 {
		t.Errorf("expected 3.14, got %f", value)
	}

	defaultValue := rcm.GetFloatWithDefault("nonexistent_key", 1.0)
	if defaultValue != 1.0 {
		t.Errorf("expected default value 1.0, got %f", defaultValue)
	}
}

func TestGetStringWithDefault(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := rcm.GetStringWithDefault("string_key", "default")
	if value != "test_value" {
		t.Errorf("expected 'test_value', got '%s'", value)
	}

	defaultValue := rcm.GetStringWithDefault("nonexistent_key", "default")
	if defaultValue != "default" {
		t.Errorf("expected default value 'default', got '%s'", defaultValue)
	}
}

func TestGetBoolWithDefault(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := rcm.GetBoolWithDefault("bool_key", false)
	if !value {
		t.Error("expected true, got false")
	}

	defaultValue := rcm.GetBoolWithDefault("nonexistent_key", false)
	if defaultValue {
		t.Error("expected default value false, got true")
	}
}

func TestGetDurationWithDefault(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	value := rcm.GetDurationWithDefault("duration_key", time.Second)
	expected := 5 * time.Second
	if value != expected {
		t.Errorf("expected %v, got %v", expected, value)
	}

	defaultValue := rcm.GetDurationWithDefault("nonexistent_key", time.Second)
	if defaultValue != time.Second {
		t.Errorf("expected default value %v, got %v", time.Second, defaultValue)
	}
}

func TestStartLoading(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	rcm.ctx, rcm.cancel = context.WithCancel(context.Background())

	rcm.StartLoading(100 * time.Millisecond)

	time.Sleep(150 * time.Millisecond)

	if len(rcm.config) == 0 {
		t.Error("config was not loaded after StartLoading")
	}

	rcm.StopLoading()
}

func TestStopLoading(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	rcm.ctx, rcm.cancel = context.WithCancel(context.Background())

	rcm.StartLoading(50 * time.Millisecond)

	time.Sleep(100 * time.Millisecond)

	rcm.StopLoading()

	select {
	case <-rcm.ctx.Done():
	default:
		t.Error("context was not cancelled after StopLoading")
	}
}

func TestConcurrentAccess(t *testing.T) {
	mr, client := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	serviceName := "test_service"
	config := createTestConfig(t, serviceName)

	for key, value := range config {
		if err := mr.Set(key, value.(string)); err != nil {
			t.Fatalf("failed to set config in miniredis: %v", err)
		}
	}

	rcm := &RedisConfigManager{
		serviceName: serviceName,
		config:      make(map[string]string),
		r:           client,
		ctx:         context.Background(),
	}

	if err := rcm.LoadConfig(context.Background()); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	done := make(chan bool)
	concurrency := 10

	for i := 0; i < concurrency; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				_, _ = rcm.GetString("string_key")
				_, _ = rcm.GetInt("int_key")
				_, _ = rcm.GetFloat("float_key")
			}
		}()
	}

	for i := 0; i < concurrency; i++ {
		<-done
	}
}

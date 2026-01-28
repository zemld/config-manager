package mcm

import (
	"context"
	"fmt"
	"time"
)

type InMemoryConfigManager struct {
	data map[string]any
}

func NewMockConfigManager(data map[string]any) *InMemoryConfigManager {
	return &InMemoryConfigManager{
		data: data,
	}
}

func (mcm *InMemoryConfigManager) StartLoading(interval time.Duration) {}
func (mcm *InMemoryConfigManager) StopLoading()                        {}
func (mcm *InMemoryConfigManager) LoadConfig(ctx context.Context) error {
	return nil
}

func (mcm *InMemoryConfigManager) GetInt(key string) (int, error) {
	value, ok := mcm.data[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	intValue, ok := value.(int)
	if !ok {
		return 0, fmt.Errorf("key %s is not an int", key)
	}

	return intValue, nil
}

func (mcm *InMemoryConfigManager) GetFloat(key string) (float64, error) {
	value, ok := mcm.data[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	floatValue, ok := value.(float64)
	if !ok {
		return 0, fmt.Errorf("key %s is not a float", key)
	}

	return floatValue, nil
}

func (mcm *InMemoryConfigManager) GetString(key string) (string, error) {
	value, ok := mcm.data[key]
	if !ok {
		return "", fmt.Errorf("key %s not found", key)
	}

	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("key %s is not a string", key)
	}

	return stringValue, nil
}

func (mcm *InMemoryConfigManager) GetBool(key string) (bool, error) {
	value, ok := mcm.data[key]
	if !ok {
		return false, fmt.Errorf("key %s not found", key)
	}

	boolValue, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("key %s is not a bool", key)
	}

	return boolValue, nil
}

func (mcm *InMemoryConfigManager) GetDuration(key string) (time.Duration, error) {
	value, ok := mcm.data[key]
	if !ok {
		return 0, fmt.Errorf("key %s not found", key)
	}

	durationValue, ok := value.(time.Duration)
	if !ok {
		return 0, fmt.Errorf("key %s is not a duration", key)
	}

	return durationValue, nil
}

func (mcm *InMemoryConfigManager) GetIntWithDefault(key string, defaultValue int) int {
	value, err := mcm.GetInt(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (mcm *InMemoryConfigManager) GetFloatWithDefault(key string, defaultValue float64) float64 {
	value, err := mcm.GetFloat(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (mcm *InMemoryConfigManager) GetStringWithDefault(key string, defaultValue string) string {
	value, err := mcm.GetString(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (mcm *InMemoryConfigManager) GetBoolWithDefault(key string, defaultValue bool) bool {
	value, err := mcm.GetBool(key)
	if err != nil {
		return defaultValue
	}

	return value
}

func (mcm *InMemoryConfigManager) GetDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	value, err := mcm.GetDuration(key)
	if err != nil {
		return defaultValue
	}

	return value
}

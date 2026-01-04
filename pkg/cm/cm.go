package cm

import (
	"context"
	"time"
)

type ConfigManager interface {
	ConfigLoader
	ConfigGetter
	ConfigGetterWithDefault
}

type ConfigLoader interface {
	StartLoading(interval time.Duration)
	StopLoading()
	LoadConfig(ctx context.Context) error
}

type ConfigGetter interface {
	GetInt(key string) (int, error)
	GetFloat(key string) (float64, error)
	GetString(key string) (string, error)
	GetBool(key string) (bool, error)
	GetDuration(key string) (time.Duration, error)
}

type ConfigGetterWithDefault interface {
	GetIntWithDefault(key string, defaultValue int) int
	GetFloatWithDefault(key string, defaultValue float64) float64
	GetStringWithDefault(key string, defaultValue string) string
	GetBoolWithDefault(key string, defaultValue bool) bool
	GetDurationWithDefault(key string, defaultValue time.Duration) time.Duration
}

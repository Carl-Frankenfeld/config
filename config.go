package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Configuration holds the Configuration key-value pairs and provides thread-safe access.
type Configuration struct {
	mu      sync.RWMutex
	keyvals map[string]any
}

// configtype defines the types that can be used in the configuration.
type configtype interface {
	string | int | int64 | float64 | bool
}

// config is the global configuration instance.
var config = Configuration{
	keyvals: make(map[string]any),
}

// Config returns the global configuration instance.
func Config() *Configuration {
	return &config
}

// must is a helper function that panics if an error is encountered.
func must[T any](res T, err error) T {
	if err != nil {
		panic(err)
	}
	return res
}

// ReadFile reads a JSON configuration file and updates the global configuration.
func ReadFile(fname string) *Configuration {
	cfgFile := must(os.Open(fname))
	defer cfgFile.Close()
	if err := json.NewDecoder(cfgFile).Decode(&config.keyvals); err != nil {
		panic(err)
	}
	return &config
}

// Get retrieves a value from the configuration by key.
func (c *Configuration) Get(key string) (any, bool) {
	c.mu.RLock()
	val, ok := c.keyvals[key]
	c.mu.RUnlock()
	return val, ok
}

// Exists checks if a key exists in the configuration.
func (c *Configuration) Exists(key string) bool {
	c.mu.RLock()
	_, ok := c.keyvals[key]
	c.mu.RUnlock()
	return ok
}

// Get retrieves a value from the configuration by key and converts it to the specified type.
func Get[T configtype](c *Configuration, key string) T {
	c.mu.RLock()
	val := c.keyvals[key]
	c.mu.RUnlock()
	return ConvertTo[T](val)
}

// GetStr retrieves a string value from the configuration by key.
func (c *Configuration) GetStr(key string) string {
	return Get[string](c, key)
}

// GetInt retrieves an int value from the configuration by key.
func (c *Configuration) GetInt(key string) int {
	return Get[int](c, key)
}

// GetInt64 retrieves an int64 value from the configuration by key.
func (c *Configuration) GetInt64(key string) int64 {
	return Get[int64](c, key)
}

// GetFloat64 retrieves a float64 value from the configuration by key.
func (c *Configuration) GetFloat64(key string) float64 {
	return Get[float64](c, key)
}

// GetBool retrieves a bool value from the configuration by key.
func (c *Configuration) GetBool(key string) bool {
	return Get[bool](c, key)
}

// Set sets a value in the configuration by key.
func (c *Configuration) Set(key string, val any) {
	c.mu.Lock()
	c.keyvals[key] = val
	c.mu.Unlock()
}

// Delete removes a key from the configuration.
func (c *Configuration) Delete(key string) {
	c.mu.Lock()
	delete(c.keyvals, key)
	c.mu.Unlock()
}

// ConvertTo converts a value to the specified type.
func ConvertTo[T configtype](val any) T {

	// type already matches
	if v, ok := val.(T); ok {
		return v
	}

	// create new value of type T (the type to convert to). Holds default value for type T for no-op.
	var t T

	switch v := val.(type) {
	// value to convert is a string
	case string:
		switch any(t).(type) {
		case string:
			return any(v).(T)
		case int:
			r, err := strconv.Atoi(v)
			if err != nil {
				r = 0
			}
			return any(r).(T)
		case int64:
			r, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				r = 0
			}
			return any(r).(T)
		case float64:
			r, err := strconv.ParseFloat(v, 64)
			if err != nil {
				r = 0.0
			}
			return any(r).(T)
		case bool:
			vLower := strings.ToLower(v)
			if (vLower == "true") || (vLower == "1") || (vLower == "yes") {
				return any(true).(T)
			}
			return any(false).(T)
		}

	// value to convert is an int
	case int:
		switch any(t).(type) {
		case string:
			return any(fmt.Sprintf("%v", v)).(T)
		case int:
			return any(v).(T)
		case int64:
			return any(int64(v)).(T)
		case float64:
			return any(float64(v)).(T)
		case bool:
			if v != 0 {
				return any(true).(T)
			}
			return any(false).(T)
		}

	// value to convert is an int64
	case int64:
		switch any(t).(type) {
		case string:
			return any(fmt.Sprintf("%v", v)).(T)
		case int:
			return any(int(v)).(T)
		case int64:
			return any(v).(T)
		case float64:
			return any(float64(v)).(T)
		case bool:
			if v != 0 {
				return any(true).(T)
			}
			return any(false).(T)
		}

	// value to convert is a float64
	case float64:
		switch any(t).(type) {
		case string:
			return any(fmt.Sprintf("%f", v)).(T)
		case int:
			return any(int(v)).(T)
		case int64:
			return any(int64(v)).(T)
		case float64:
			return any(v).(T)
		case bool:
			if v != 0.0 {
				return any(true).(T)
			}
			return any(false).(T)
		}

	// value to convert is a bool
	case bool:
		switch any(t).(type) {
		case string:
			if v {
				return any(v).(T)
			}
			return any(false).(T)
		case int:
			if v {
				return any(int(1)).(T)
			}
			return any(int(0)).(T)
		case int64:
			if v {
				return any(int64(1)).(T)
			}
			return any(int64(0)).(T)
		case float64:
			if v {
				return any(float64(1.0)).(T)
			}
			return any(float64(0.0)).(T)
		case bool:
			return any(v).(T)
		}
	}

	return t
}

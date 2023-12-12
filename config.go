package violet

import (
	"github.com/BurntSushi/toml"
	"github.com/CharLemAznable/violet/internal/resilience"
)

type Config struct {
	Endpoint []EndpointConfig

	Defaults Defaults
}

type Defaults struct {
	Resilience ResilienceConfig
}

type EndpointConfig struct {
	Name string

	Location            string
	StripLocationPrefix string

	TargetURL string

	DumpTarget string
	DumpSource string

	Resilience ResilienceConfig
}

type ResilienceConfig struct {
	Bulkhead       BulkheadConfig
	TimeLimiter    TimeLimiterConfig
	RateLimiter    RateLimiterConfig
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	Cache          CacheConfig
	Fallback       FallbackConfig
}

type (
	BulkheadConfig       = resilience.BulkheadConfig
	TimeLimiterConfig    = resilience.TimeLimiterConfig
	RateLimiterConfig    = resilience.RateLimiterConfig
	CircuitBreakerConfig = resilience.CircuitBreakerConfig
	RetryConfig          = resilience.RetryConfig
	CacheConfig          = resilience.CacheConfig
	FallbackConfig       = resilience.FallbackConfig
)

func LoadConfig(data string) (*Config, error) {
	config := &Config{}
	if _, err := toml.Decode(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func FormatConfig(config *Config) *Config {
	cfg := &Config{}
	cfg.Endpoint = make([]EndpointConfig, len(config.Endpoint))
	for i, originEndpoint := range config.Endpoint {
		cfg.Endpoint[i] = originEndpoint
		formatResilienceConfig(&cfg.Endpoint[i].Resilience, &config.Defaults.Resilience)
	}
	return cfg
}

package resilience_test

import (
	"github.com/CharLemAznable/violet/internal/resilience"
	"github.com/prometheus/client_golang/prometheus"
	"testing"
)

func TestDecorator(t *testing.T) {
	decorator := resilience.NewDecorator("test", &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true", Order: "100"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true", Order: "300"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true", Order: "200"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true", Order: "500"},
		Retry:          resilience.RetryConfig{Disabled: "true", Order: "400"},
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	})
	err := decorator.RegisterFn(prometheus.DefaultRegisterer)
	if err != nil {
		t.Errorf("Expected register return no err, but got %v", err)
	}
	ret := decorator.UnregisterFn(prometheus.DefaultRegisterer)
	if !ret {
		t.Error("Expected unregister succeed, but not")
	}
}

func newBulkheadResilienceConfig(config *resilience.BulkheadConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       *config,
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	}
}

func newTimeLimiterResilienceConfig(config *resilience.TimeLimiterConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    *config,
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	}
}

func newRateLimiterResilienceConfig(config *resilience.RateLimiterConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    *config,
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	}
}

func newCircuitBreakerResilienceConfig(config *resilience.CircuitBreakerConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: *config,
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	}
}

func newRetryResilienceConfig(config *resilience.RetryConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          *config,
		Cache:          resilience.CacheConfig{},
		Fallback:       resilience.FallbackConfig{},
	}
}

func newCacheResilienceConfig(config *resilience.CacheConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          *config,
		Fallback:       resilience.FallbackConfig{},
	}
}

func newFallbackResilienceConfig(config *resilience.FallbackConfig) *resilience.Config {
	return &resilience.Config{
		Bulkhead:       resilience.BulkheadConfig{Disabled: "true"},
		TimeLimiter:    resilience.TimeLimiterConfig{Disabled: "true"},
		RateLimiter:    resilience.RateLimiterConfig{Disabled: "true"},
		CircuitBreaker: resilience.CircuitBreakerConfig{Disabled: "true"},
		Retry:          resilience.RetryConfig{Disabled: "true"},
		Cache:          resilience.CacheConfig{},
		Fallback:       *config,
	}
}

package violet

import "github.com/CharLemAznable/gogo/ext"

func formatResilienceConfig(cfg *ResilienceConfig, def *ResilienceConfig) {
	formatBulkheadConfig(&cfg.Bulkhead, &def.Bulkhead)
	formatTimeLimiterConfig(&cfg.TimeLimiter, &def.TimeLimiter)
	formatRateLimiterConfig(&cfg.RateLimiter, &def.RateLimiter)
	formatCircuitBreakerConfig(&cfg.CircuitBreaker, &def.CircuitBreaker)
	formatRetryConfig(&cfg.Retry, &def.Retry)
	formatCacheConfig(&cfg.Cache, &def.Cache)
	formatFallbackConfig(&cfg.Fallback, &def.Fallback)
}

func formatBulkheadConfig(cfg *BulkheadConfig, def *BulkheadConfig) {
	cfg.Disabled = defaultString(cfg.Disabled, def.Disabled)
	cfg.MaxConcurrentCalls = defaultString(cfg.MaxConcurrentCalls, def.MaxConcurrentCalls)
	cfg.MaxWaitDuration = defaultString(cfg.MaxWaitDuration, def.MaxWaitDuration)
	cfg.WhenFullResponse = defaultString(cfg.WhenFullResponse, def.WhenFullResponse)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatTimeLimiterConfig(cfg *TimeLimiterConfig, def *TimeLimiterConfig) {
	cfg.Disabled = defaultString(cfg.Disabled, def.Disabled)
	cfg.TimeoutDuration = defaultString(cfg.TimeoutDuration, def.TimeoutDuration)
	cfg.WhenTimeoutResponse = defaultString(cfg.WhenTimeoutResponse, def.WhenTimeoutResponse)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatRateLimiterConfig(cfg *RateLimiterConfig, def *RateLimiterConfig) {
	cfg.Disabled = defaultString(cfg.Disabled, def.Disabled)
	cfg.TimeoutDuration = defaultString(cfg.TimeoutDuration, def.TimeoutDuration)
	cfg.LimitRefreshPeriod = defaultString(cfg.LimitRefreshPeriod, def.LimitRefreshPeriod)
	cfg.LimitForPeriod = defaultString(cfg.LimitForPeriod, def.LimitForPeriod)
	cfg.WhenOverRateResponse = defaultString(cfg.WhenOverRateResponse, def.WhenOverRateResponse)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatCircuitBreakerConfig(cfg *CircuitBreakerConfig, def *CircuitBreakerConfig) {
	cfg.Disabled = defaultString(cfg.Disabled, def.Disabled)
	cfg.SlidingWindowType = defaultString(cfg.SlidingWindowType, def.SlidingWindowType)
	cfg.SlidingWindowSize = defaultString(cfg.SlidingWindowSize, def.SlidingWindowSize)
	cfg.MinimumNumberOfCalls = defaultString(cfg.MinimumNumberOfCalls, def.MinimumNumberOfCalls)
	cfg.FailureRateThreshold = defaultString(cfg.FailureRateThreshold, def.FailureRateThreshold)
	cfg.SlowCallRateThreshold = defaultString(cfg.SlowCallRateThreshold, def.SlowCallRateThreshold)
	cfg.SlowCallDurationThreshold = defaultString(cfg.SlowCallDurationThreshold, def.SlowCallDurationThreshold)
	cfg.ResponseFailedPredicate = defaultString(cfg.ResponseFailedPredicate, def.ResponseFailedPredicate)
	cfg.ResponseFailedPredicateContext = defaultMap(cfg.ResponseFailedPredicateContext, def.ResponseFailedPredicateContext)
	cfg.AutomaticTransitionFromOpenToHalfOpen = defaultString(cfg.AutomaticTransitionFromOpenToHalfOpen, def.AutomaticTransitionFromOpenToHalfOpen)
	cfg.WaitIntervalInOpenState = defaultString(cfg.WaitIntervalInOpenState, def.WaitIntervalInOpenState)
	cfg.PermittedNumberOfCallsInHalfOpenState = defaultString(cfg.PermittedNumberOfCallsInHalfOpenState, def.PermittedNumberOfCallsInHalfOpenState)
	cfg.MaxWaitDurationInHalfOpenState = defaultString(cfg.MaxWaitDurationInHalfOpenState, def.MaxWaitDurationInHalfOpenState)
	cfg.WhenOverLoadResponse = defaultString(cfg.WhenOverLoadResponse, def.WhenOverLoadResponse)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatRetryConfig(cfg *RetryConfig, def *RetryConfig) {
	cfg.Disabled = defaultString(cfg.Disabled, def.Disabled)
	cfg.MaxAttempts = defaultString(cfg.MaxAttempts, def.MaxAttempts)
	cfg.FailAfterMaxAttempts = defaultString(cfg.FailAfterMaxAttempts, def.FailAfterMaxAttempts)
	cfg.ResponseFailedPredicate = defaultString(cfg.ResponseFailedPredicate, def.ResponseFailedPredicate)
	cfg.ResponseFailedPredicateContext = defaultMap(cfg.ResponseFailedPredicateContext, def.ResponseFailedPredicateContext)
	cfg.WaitInterval = defaultString(cfg.WaitInterval, def.WaitInterval)
	cfg.WhenMaxRetriesResponse = defaultString(cfg.WhenMaxRetriesResponse, def.WhenMaxRetriesResponse)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatCacheConfig(cfg *CacheConfig, def *CacheConfig) {
	cfg.Enabled = defaultString(cfg.Enabled, def.Enabled)
	cfg.Capacity = defaultString(cfg.Capacity, def.Capacity)
	cfg.ItemTTL = defaultString(cfg.ItemTTL, def.ItemTTL)
	cfg.ResponseCachePredicate = defaultString(cfg.ResponseCachePredicate, def.ResponseCachePredicate)
	cfg.ResponseCachePredicateContext = defaultMap(cfg.ResponseCachePredicateContext, def.ResponseCachePredicateContext)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func formatFallbackConfig(cfg *FallbackConfig, def *FallbackConfig) {
	cfg.Enabled = defaultString(cfg.Enabled, def.Enabled)
	cfg.FallbackResponse = defaultString(cfg.FallbackResponse, def.FallbackResponse)
	cfg.FallbackFunction = defaultString(cfg.FallbackFunction, def.FallbackFunction)
	cfg.FallbackFunctionContext = defaultMap(cfg.FallbackFunctionContext, def.FallbackFunctionContext)
	cfg.ResponseFailedPredicate = defaultString(cfg.ResponseFailedPredicate, def.ResponseFailedPredicate)
	cfg.ResponseFailedPredicateContext = defaultMap(cfg.ResponseFailedPredicateContext, def.ResponseFailedPredicateContext)
	cfg.Order = defaultString(cfg.Order, def.Order)
}

func defaultString(v string, def string) string {
	return ext.EmptyThen(v, func() string { return def })
}

func defaultMap(v map[string]string, def map[string]string) map[string]string {
	return ext.MapWithDefault(v, def)
}

package resilience

import (
	"github.com/CharLemAznable/gogo/ext"
	"github.com/CharLemAznable/resilience4go/bulkhead"
	"github.com/CharLemAznable/resilience4go/cache"
	"github.com/CharLemAznable/resilience4go/circuitbreaker"
	"github.com/CharLemAznable/resilience4go/promhelper"
	"github.com/CharLemAznable/resilience4go/ratelimiter"
	"github.com/CharLemAznable/resilience4go/retry"
	"github.com/CharLemAznable/resilience4go/timelimiter"
	. "github.com/CharLemAznable/violet/internal/types"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type Config struct {
	Bulkhead       BulkheadConfig
	TimeLimiter    TimeLimiterConfig
	RateLimiter    RateLimiterConfig
	CircuitBreaker CircuitBreakerConfig
	Retry          RetryConfig
	Cache          CacheConfig
	Fallback       FallbackConfig
}

type Decorator struct {
	Bulkhead       bulkhead.Bulkhead
	TimeLimiter    timelimiter.TimeLimiter
	RateLimiter    ratelimiter.RateLimiter
	CircuitBreaker circuitbreaker.CircuitBreaker
	Retry          retry.Retry
	Cache          cache.Cache[Req, Rsp]

	decorators ext.OrderedSlice[*OrderedDecorator]

	RegisterFn   promhelper.RegisterFn
	UnregisterFn promhelper.UnregisterFn
}

func NewDecorator(name string, config *Config) *Decorator {
	e := &Decorator{decorators: make(ext.OrderedSlice[*OrderedDecorator], 7)}
	e.Bulkhead, e.decorators[0] = NewBulkheadPlugin(name, &config.Bulkhead)
	e.TimeLimiter, e.decorators[1] = NewTimeLimiterPlugin(name, &config.TimeLimiter)
	e.RateLimiter, e.decorators[2] = NewRateLimiterPlugin(name, &config.RateLimiter)
	e.CircuitBreaker, e.decorators[3] = NewCircuitBreakerPlugin(name, &config.CircuitBreaker)
	e.Retry, e.decorators[4] = NewRetryPlugin(name, &config.Retry)
	e.Cache, e.decorators[5] = NewCachePlugin(name, &config.Cache)
	e.decorators[6] = NewFallbackPlugin(&config.Fallback)
	e.decorators.Sort()
	e.RegisterFn, e.UnregisterFn = initRegistryFn(e)
	return e
}

func (e *Decorator) Decorate(rp ReverseProxy) ReverseProxy {
	ret := rp
	for _, decorator := range e.decorators {
		ret = decorator.Decorate(ret)
	}
	return ret
}

func initRegistryFn(e *Decorator) (promhelper.RegisterFn, promhelper.UnregisterFn) {
	var (
		bulkheadRegister       = emptyRegister
		timelimiterRegister    = emptyRegister
		ratelimiterRegister    = emptyRegister
		circuitbreakerRegister = emptyRegister
		retryRegister          = emptyRegister
		cacheRegister          = emptyRegister

		bulkheadUnregister       = emptyUnregister
		timelimiterUnregister    = emptyUnregister
		ratelimiterUnregister    = emptyUnregister
		circuitbreakerUnregister = emptyUnregister
		retryUnregister          = emptyUnregister
		cacheUnregister          = emptyUnregister
	)
	if e.Bulkhead != nil {
		bulkheadRegister, bulkheadUnregister =
			promhelper.BulkheadRegistry(e.Bulkhead)
	}
	if e.TimeLimiter != nil {
		timelimiterRegister, timelimiterUnregister =
			promhelper.TimeLimiterRegistry(e.TimeLimiter)
	}
	if e.RateLimiter != nil {
		ratelimiterRegister, ratelimiterUnregister =
			promhelper.RateLimiterRegistry(e.RateLimiter)
	}
	if e.CircuitBreaker != nil {
		var buckets []float64
		for _, b := range prometheus.DefBuckets {
			buckets = append(buckets, float64(time.Second)*b)
		}
		circuitbreakerRegister, circuitbreakerUnregister =
			promhelper.CircuitBreakerRegistry(e.CircuitBreaker, buckets...)
	}
	if e.Retry != nil {
		retryRegister, retryUnregister =
			promhelper.RetryRegistry(e.Retry)
	}
	if e.Cache != nil {
		cacheRegister, cacheUnregister =
			promhelper.CacheRegistry(e.Cache)
	}
	return func(registerer prometheus.Registerer) error {
			err := prometheus.MultiError{}
			err.Append(bulkheadRegister(registerer))
			err.Append(timelimiterRegister(registerer))
			err.Append(ratelimiterRegister(registerer))
			err.Append(circuitbreakerRegister(registerer))
			err.Append(retryRegister(registerer))
			err.Append(cacheRegister(registerer))
			return err.MaybeUnwrap()
		},
		func(registerer prometheus.Registerer) bool {
			ret := true
			ret = bulkheadUnregister(registerer) && ret
			ret = timelimiterUnregister(registerer) && ret
			ret = ratelimiterUnregister(registerer) && ret
			ret = circuitbreakerUnregister(registerer) && ret
			ret = retryUnregister(registerer) && ret
			ret = cacheUnregister(registerer) && ret
			return ret
		}
}

func emptyRegister(_ prometheus.Registerer) error { return nil }

func emptyUnregister(_ prometheus.Registerer) bool { return true }

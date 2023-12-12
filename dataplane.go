package violet

import (
	"fmt"
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/bulkhead"
	"github.com/CharLemAznable/resilience4go/cache"
	"github.com/CharLemAznable/resilience4go/circuitbreaker"
	"github.com/CharLemAznable/resilience4go/promhelper"
	"github.com/CharLemAznable/resilience4go/ratelimiter"
	"github.com/CharLemAznable/resilience4go/retry"
	"github.com/CharLemAznable/resilience4go/timelimiter"
	"github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DataPlane interface {
	http.Handler
	GetConfig() *Config
	SetConfig(*Config)

	GetRegisterer() prometheus.Registerer
	GetGatherer() prometheus.Gatherer
	SetRegistry(*prometheus.Registry)

	DisableCircuitBreaker(string) error
	ForceOpenCircuitBreaker(string) error
	CloseCircuitBreaker(string) error
	CircuitBreakerState(string) string
}

func NewDataPlane(config *Config) DataPlane {
	plane := &atomicDataPlane{}
	plane.SetConfig(config)
	return plane
}

type atomicDataPlane struct {
	sync.RWMutex
	pointer  atomic.Pointer[dataMux]
	registry *prometheus.Registry
}

func (plane *atomicDataPlane) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	plane.loadDataMux().ServeHTTP(w, r)
}

func (plane *atomicDataPlane) GetConfig() *Config {
	return plane.loadDataMux().rawConfig
}

func (plane *atomicDataPlane) SetConfig(config *Config) {
	plane.Lock()
	defer plane.Unlock()
	if old := plane.swapDataMux(newDataMux(config)); old != nil {
		old.unregisterFn(plane.getRegisterer())
	}
	_ = plane.loadDataMux().registerFn(plane.getRegisterer())
}

func (plane *atomicDataPlane) GetRegisterer() prometheus.Registerer {
	plane.RLock()
	defer plane.RUnlock()
	return plane.getRegisterer()
}

func (plane *atomicDataPlane) GetGatherer() prometheus.Gatherer {
	plane.RLock()
	defer plane.RUnlock()
	return plane.getGatherer()
}

func (plane *atomicDataPlane) SetRegistry(registry *prometheus.Registry) {
	plane.Lock()
	defer plane.Unlock()
	plane.loadDataMux().unregisterFn(plane.getRegisterer())
	plane.registry = registry
	_ = plane.loadDataMux().registerFn(plane.getRegisterer())
}

func (plane *atomicDataPlane) DisableCircuitBreaker(name string) error {
	if endpoint, exists := plane.loadDataMux().endpoint(name); exists {
		return endpoint.disableCircuitBreaker()
	}
	return &endpointNotExists{name: name}
}

func (plane *atomicDataPlane) ForceOpenCircuitBreaker(name string) error {
	if endpoint, exists := plane.loadDataMux().endpoint(name); exists {
		return endpoint.forceOpenCircuitBreaker()
	}
	return &endpointNotExists{name: name}
}

func (plane *atomicDataPlane) CloseCircuitBreaker(name string) error {
	if endpoint, exists := plane.loadDataMux().endpoint(name); exists {
		return endpoint.closeCircuitBreaker()
	}
	return &endpointNotExists{name: name}
}

func (plane *atomicDataPlane) CircuitBreakerState(name string) string {
	if endpoint, exists := plane.loadDataMux().endpoint(name); exists {
		return endpoint.circuitBreakerState()
	}
	return "UNKNOWN"
}

func (plane *atomicDataPlane) loadDataMux() *dataMux {
	return plane.pointer.Load()
}

func (plane *atomicDataPlane) swapDataMux(new *dataMux) *dataMux {
	return plane.pointer.Swap(new)
}

func (plane *atomicDataPlane) getRegisterer() prometheus.Registerer {
	if plane.registry != nil {
		return plane.registry
	}
	return prometheus.DefaultRegisterer
}

func (plane *atomicDataPlane) getGatherer() prometheus.Gatherer {
	if plane.registry != nil {
		return plane.registry
	}
	return prometheus.DefaultGatherer
}

type endpointNotExists struct {
	name string
}

func (e *endpointNotExists) Error() string {
	return fmt.Sprintf("endpoint with name [%s] not exists", e.name)
}

type dataMux struct {
	http.ServeMux

	rawConfig *Config
	endpoints map[string]*endpoint

	registerFn   promhelper.RegisterFn
	unregisterFn promhelper.UnregisterFn
}

func newDataMux(config *Config) *dataMux {
	mux := new(dataMux)
	mux.rawConfig = config
	formatConfig := FormatConfig(config)
	mux.endpoints = make(map[string]*endpoint)
	for _, ec := range formatConfig.Endpoint {
		endpoint := newEndpoint(&ec)
		mux.endpoints[endpoint.name] = endpoint
		mux.HandleFunc(endpoint.location, func(w http.ResponseWriter, r *http.Request) {
			if endpoint.stripLocationPrefix {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, endpoint.location)
			}
			endpoint.proxy.ServeHTTP(w, r)
		})
	}
	mux.registerFn, mux.unregisterFn = mux.registry()
	return mux
}

func (mux *dataMux) registry() (promhelper.RegisterFn, promhelper.UnregisterFn) {
	var registerFns []promhelper.RegisterFn
	var unregisterFns []promhelper.UnregisterFn
	for _, endpoint := range mux.endpoints {
		registerFns = append(registerFns, endpoint.registerFn)
		unregisterFns = append(unregisterFns, endpoint.unregisterFn)
	}
	return func(registerer prometheus.Registerer) error {
			err := prometheus.MultiError{}
			for _, registerFn := range registerFns {
				err.Append(registerFn(registerer))
			}
			return err.MaybeUnwrap()
		},
		func(registerer prometheus.Registerer) bool {
			ret := true
			for _, unregisterFn := range unregisterFns {
				ret = unregisterFn(registerer) && ret
			}
			return ret
		}
}

func (mux *dataMux) endpoint(name string) (*endpoint, bool) {
	endpoint, exists := mux.endpoints[name]
	return endpoint, exists
}

type endpoint struct {
	name string

	location            string
	stripLocationPrefix bool

	proxy *httputil.ReverseProxy

	bulkhead       bulkhead.Bulkhead
	timelimiter    timelimiter.TimeLimiter
	ratelimiter    ratelimiter.RateLimiter
	circuitbreaker circuitbreaker.CircuitBreaker
	retry          retry.Retry
	cache          cache.Cache[*http.Request, *http.Response]

	registerFn   promhelper.RegisterFn
	unregisterFn promhelper.UnregisterFn
}

func newEndpoint(config *EndpointConfig) *endpoint {
	targetURL, err := url.Parse(config.TargetURL)
	ge.PanicIfError(err)
	e := &endpoint{
		name:                config.Name,
		location:            config.Location,
		stripLocationPrefix: ge.ToBool(config.StripLocationPrefix),
		proxy:               proxy.NewReverseProxy(targetURL),
	}
	e.proxy = proxy.DumpDecorator(ge.ToBool(config.DumpTarget), proxy.TargetDump, e.name)(e.proxy)
	bulkheadEntry, bulkheadDecorator := resilience.NewBulkheadPlugin(e.name, &config.Resilience.Bulkhead)
	e.bulkhead = bulkheadEntry
	e.proxy = bulkheadDecorator(e.proxy)
	timelimiterEntry, timelimiterDecorator := resilience.NewTimeLimiterPlugin(e.name, &config.Resilience.TimeLimiter)
	e.timelimiter = timelimiterEntry
	e.proxy = timelimiterDecorator(e.proxy)
	ratelimiterEntry, ratelimiterDecorator := resilience.NewRateLimiterPlugin(e.name, &config.Resilience.RateLimiter)
	e.ratelimiter = ratelimiterEntry
	e.proxy = ratelimiterDecorator(e.proxy)
	circuitbreakerEntry, circuitbreakerDecorator := resilience.NewCircuitBreakerPlugin(e.name, &config.Resilience.CircuitBreaker)
	e.circuitbreaker = circuitbreakerEntry
	e.proxy = circuitbreakerDecorator(e.proxy)
	retryEntry, retryDecorator := resilience.NewRetryPlugin(e.name, &config.Resilience.Retry)
	e.retry = retryEntry
	e.proxy = retryDecorator(e.proxy)
	cacheEntry, cacheDecorator := resilience.NewCachePlugin(e.name, &config.Resilience.Cache)
	e.cache = cacheEntry
	e.proxy = cacheDecorator(e.proxy)
	fallbackDecorator := resilience.NewFallbackPlugin(&config.Resilience.Fallback)
	e.proxy = fallbackDecorator(e.proxy)
	e.proxy = proxy.DumpDecorator(ge.ToBool(config.DumpSource), proxy.SourceDump, e.name)(e.proxy)
	e.registerFn, e.unregisterFn = e.registry()
	return e
}

func (e *endpoint) registry() (promhelper.RegisterFn, promhelper.UnregisterFn) {
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
	if e.bulkhead != nil {
		bulkheadRegister, bulkheadUnregister =
			promhelper.BulkheadRegistry(e.bulkhead)
	}
	if e.timelimiter != nil {
		timelimiterRegister, timelimiterUnregister =
			promhelper.TimeLimiterRegistry(e.timelimiter)
	}
	if e.ratelimiter != nil {
		ratelimiterRegister, ratelimiterUnregister =
			promhelper.RateLimiterRegistry(e.ratelimiter)
	}
	if e.circuitbreaker != nil {
		var buckets []float64
		for _, b := range prometheus.DefBuckets {
			buckets = append(buckets, float64(time.Second)*b)
		}
		circuitbreakerRegister, circuitbreakerUnregister =
			promhelper.CircuitBreakerRegistry(e.circuitbreaker, buckets...)
	}
	if e.retry != nil {
		retryRegister, retryUnregister =
			promhelper.RetryRegistry(e.retry)
	}
	if e.cache != nil {
		cacheRegister, cacheUnregister =
			promhelper.CacheRegistry(e.cache)
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

func (e *endpoint) disableCircuitBreaker() error {
	if e.circuitbreaker == nil {
		return nil
	}
	return e.circuitbreaker.TransitionToDisabled()
}

func (e *endpoint) forceOpenCircuitBreaker() error {
	if e.circuitbreaker == nil {
		return nil
	}
	return e.circuitbreaker.TransitionToForcedOpen()
}

func (e *endpoint) closeCircuitBreaker() error {
	if e.circuitbreaker == nil {
		return nil
	}
	return e.circuitbreaker.TransitionToClosedState()
}

func (e *endpoint) circuitBreakerState() string {
	if e.circuitbreaker == nil {
		return "UNKNOWN"
	}
	return string(e.circuitbreaker.State())
}

package violet

import (
	"fmt"
	"github.com/CharLemAznable/ge"
	"github.com/CharLemAznable/resilience4go/promhelper"
	"github.com/CharLemAznable/violet/internal/proxy"
	"github.com/CharLemAznable/violet/internal/resilience"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
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

	resilience *resilience.Decorator

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
	e.resilience = resilience.NewDecorator(e.name, &config.Resilience)
	e.proxy = e.resilience.Decorate(e.proxy)
	e.proxy = proxy.DumpDecorator(ge.ToBool(config.DumpSource), proxy.SourceDump, e.name)(e.proxy)
	e.registerFn = func(registerer prometheus.Registerer) error {
		err := prometheus.MultiError{}
		err.Append(e.resilience.RegisterFn(registerer))
		return err.MaybeUnwrap()
	}
	e.unregisterFn = func(registerer prometheus.Registerer) bool {
		ret := true
		ret = e.resilience.UnregisterFn(registerer) && ret
		return ret
	}
	return e
}

func (e *endpoint) disableCircuitBreaker() error {
	if e.resilience.CircuitBreaker == nil {
		return nil
	}
	return e.resilience.CircuitBreaker.TransitionToDisabled()
}

func (e *endpoint) forceOpenCircuitBreaker() error {
	if e.resilience.CircuitBreaker == nil {
		return nil
	}
	return e.resilience.CircuitBreaker.TransitionToForcedOpen()
}

func (e *endpoint) closeCircuitBreaker() error {
	if e.resilience.CircuitBreaker == nil {
		return nil
	}
	return e.resilience.CircuitBreaker.TransitionToClosedState()
}

func (e *endpoint) circuitBreakerState() string {
	if e.resilience.CircuitBreaker == nil {
		return "UNKNOWN"
	}
	return string(e.resilience.CircuitBreaker.State())
}

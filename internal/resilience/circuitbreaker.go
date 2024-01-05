package resilience

import (
	"github.com/CharLemAznable/gogo/lang"
	"github.com/CharLemAznable/resilience4go/circuitbreaker"
	"github.com/CharLemAznable/resilience4go/decorator"
	. "github.com/CharLemAznable/violet/internal/types"
	"strconv"
	"strings"
	"time"
)

type CircuitBreakerConfig struct {
	Disabled                              string
	SlidingWindowType                     string
	SlidingWindowSize                     string
	MinimumNumberOfCalls                  string
	FailureRateThreshold                  string
	SlowCallRateThreshold                 string
	SlowCallDurationThreshold             string
	ResponseFailedPredicate               string
	ResponseFailedPredicateContext        map[string]string
	AutomaticTransitionFromOpenToHalfOpen string
	WaitIntervalInOpenState               string
	PermittedNumberOfCallsInHalfOpenState string
	MaxWaitDurationInHalfOpenState        string
	WhenOverLoadResponse                  string
	Order                                 string
}

const CircuitBreakerDefaultOrder = "400"

func NewCircuitBreakerPlugin(name string, config *CircuitBreakerConfig) (circuitbreaker.CircuitBreaker, *OrderedDecorator) {
	if lang.ToBool(config.Disabled) {
		return nil, newOrderedDecorator(ReverseProxyIdentity, config.Order, CircuitBreakerDefaultOrder)
	}
	entry := circuitbreaker.NewCircuitBreaker(name+"_circuitbreaker",
		circuitbreakerConfigBuilders(config)...)
	whenOverLoadFn := responseFn(config.WhenOverLoadResponse)
	return entry, newOrderedDecorator(func(rp ReverseProxy) ReverseProxy {
		decorate := decorator.OfFunction(rp.Transport.RoundTrip).WithCircuitBreaker(entry)
		if whenOverLoadFn != nil {
			decorate = decorate.WhenOverLoad(whenOverLoadFn)
		}
		rp.Transport = RoundTripperFunc(decorate.Decorate())
		return rp
	}, config.Order, CircuitBreakerDefaultOrder)
}

func circuitbreakerConfigBuilders(config *CircuitBreakerConfig) []circuitbreaker.ConfigBuilder {
	var builders []circuitbreaker.ConfigBuilder
	builders = append(builders, configSlidingWindow(config))
	builders = append(builders, configThresholds(config)...)
	builders = append(builders, configFailureResultPredicate(config))
	builders = append(builders, configOpenState(config)...)
	builders = append(builders, configHalfOpenState(config)...)
	return builders
}

func configSlidingWindow(config *CircuitBreakerConfig) circuitbreaker.ConfigBuilder {
	slidingWindowType, slidingWindowSize, minimumNumberOfCalls :=
		circuitbreaker.DefaultSlidingWindowType,
		circuitbreaker.DefaultSlidingWindowSize,
		circuitbreaker.DefaultMinimumNumberOfCalls
	if circuitbreaker.SlidingWindowType(strings.ToUpper(
		config.SlidingWindowType)) == circuitbreaker.TimeBased {
		slidingWindowType = circuitbreaker.TimeBased
	}
	if size, err := strconv.ParseInt(
		config.SlidingWindowSize, 10, 64); err == nil {
		slidingWindowSize = size
	}
	if num, err := strconv.ParseInt(
		config.MinimumNumberOfCalls, 10, 64); err == nil {
		minimumNumberOfCalls = num
	}
	return circuitbreaker.WithSlidingWindow(
		slidingWindowType, slidingWindowSize, minimumNumberOfCalls)
}

func configThresholds(config *CircuitBreakerConfig) []circuitbreaker.ConfigBuilder {
	var builders []circuitbreaker.ConfigBuilder
	if failureRateThreshold, err := strconv.ParseFloat(
		config.FailureRateThreshold, 64); err == nil {
		builders = append(builders, circuitbreaker.
			WithFailureRateThreshold(failureRateThreshold))
	}
	if slowCallRateThreshold, err := strconv.ParseFloat(
		config.SlowCallRateThreshold, 64); err == nil {
		builders = append(builders, circuitbreaker.
			WithSlowCallRateThreshold(slowCallRateThreshold))
	}
	if slowCallDurationThreshold, err := time.ParseDuration(
		config.SlowCallDurationThreshold); err == nil {
		builders = append(builders, circuitbreaker.
			WithSlowCallDurationThreshold(slowCallDurationThreshold))
	}
	return builders
}

func configFailureResultPredicate(config *CircuitBreakerConfig) circuitbreaker.ConfigBuilder {
	predicate := GetRspFailedPredicate(config.ResponseFailedPredicate)
	context := config.ResponseFailedPredicateContext
	return circuitbreaker.WithFailureResultPredicate(
		buildFailureResultPredicate(predicate, context))
}

func configOpenState(config *CircuitBreakerConfig) []circuitbreaker.ConfigBuilder {
	var builders []circuitbreaker.ConfigBuilder
	builders = append(builders, circuitbreaker.
		WithAutomaticTransitionFromOpenToHalfOpenEnabled(
			lang.ToBool(config.AutomaticTransitionFromOpenToHalfOpen)))
	if waitIntervalInOpenState, err := time.ParseDuration(
		config.WaitIntervalInOpenState); err == nil {
		builders = append(builders, circuitbreaker.
			WithWaitIntervalFunctionInOpenState(
				func(_ int64) time.Duration { return waitIntervalInOpenState }))
	}
	return builders
}

func configHalfOpenState(config *CircuitBreakerConfig) []circuitbreaker.ConfigBuilder {
	var builders []circuitbreaker.ConfigBuilder
	if permittedNumberOfCallsInHalfOpenState, err := strconv.ParseInt(
		config.PermittedNumberOfCallsInHalfOpenState, 10, 64); err == nil {
		builders = append(builders, circuitbreaker.
			WithPermittedNumberOfCallsInHalfOpenState(permittedNumberOfCallsInHalfOpenState))
	}
	if maxWaitDurationInHalfOpenState, err := time.ParseDuration(
		config.MaxWaitDurationInHalfOpenState); err == nil {
		builders = append(builders, circuitbreaker.
			WithMaxWaitDurationInHalfOpenState(maxWaitDurationInHalfOpenState))
	}
	return builders
}
